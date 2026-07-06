package main_test

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	flagvalues "go-url-shortener/internal/config/flag_values"
	"go-url-shortener/internal/logger"
	"go-url-shortener/internal/repository"
	auditmocks "go-url-shortener/internal/service/audit/mocks"
	"go-url-shortener/internal/service/shortifier"
	shortenercmd "go-url-shortener/cmd/shortener"
	"go-url-shortener/proto"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	grpcTestAddr          = "127.0.0.1:3200"
	headerAuthorization   = "authorization"
)

func TestGRPCServerSuite(t *testing.T) {
	suite.Run(t, new(GRPCServerSuite))
}

type GRPCServerSuite struct {
	suite.Suite

	ctx      context.Context
	cancel   context.CancelFunc
	listener net.Listener
	conn     *grpc.ClientConn
	client   proto.ShortenerServiceClient
	token    string
}

func (s *GRPCServerSuite) SetupSuite() {
	logger.SetDiscardOutput(true)

	lis, err := net.Listen("tcp", grpcTestAddr)
	if err != nil {
		s.T().Skipf("grpc test listener %s unavailable: %v", grpcTestAddr, err)
	}
	s.listener = lis

	sf := s.newTestShortifier()
	s.ctx, s.cancel = context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		errCh <- shortenercmd.ServeGRPCOnListener(s.ctx, lis, sf)
	}()

	s.Require().NoError(waitGRPCReady(grpcTestAddr))

	s.conn, err = grpc.NewClient(
		grpcTestAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	s.Require().NoError(err)

	s.client = proto.NewShortenerServiceClient(s.conn)
}

func (s *GRPCServerSuite) TearDownSuite() {
	if s.cancel != nil {
		s.cancel()
	}
	if s.conn != nil {
		_ = s.conn.Close()
	}
	logger.SetDiscardOutput(false)
}

func (s *GRPCServerSuite) newTestShortifier() *shortifier.Shortifier {
	mockAudit := gomock.NewController(s.T())
	audit := auditmocks.NewMockPublisher(mockAudit)
	audit.EXPECT().Publish(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	redirectAddr := flagvalues.NetAddress{
		Host:           "localhost:3200",
		Scheme:         "http",
		SchemeRequired: true,
	}

	sf, err := shortifier.NewShortifier(
		s.T().Context(),
		&seqStrGen{},
		repository.NewInMemoryStore(nil),
		audit,
		&redirectAddr,
	)
	s.Require().NoError(err)

	return sf
}

func (s *GRPCServerSuite) TestAuthInterceptorIssuesToken() {
	_, header, err := s.shorten(context.Background(), "", "http://example.com/one")
	s.Require().NoError(err)

	tokens := header.Get(headerAuthorization)
	s.Require().NotEmpty(tokens)
	s.token = tokens[0]
}

func (s *GRPCServerSuite) TestShortenURL() {
	s.ensureToken()

	resp, _, err := s.shorten(context.Background(), s.token, "http://example.com/two")
	s.Require().NoError(err)
	s.Contains(resp.GetResult(), "http://localhost:3200")
}

func (s *GRPCServerSuite) TestExpandURL() {
	s.ensureToken()
	ctx := withAuthToken(context.Background(), s.token)

	shortenResp, _, err := s.shorten(ctx, s.token, "http://example.com/three")
	s.Require().NoError(err)

	shortURL, err := url.Parse(shortenResp.GetResult())
	s.Require().NoError(err)

	expandResp, err := s.client.ExpandURL(ctx, proto.URLExpandRequest_builder{
		Id: strings.TrimPrefix(shortURL.Path, "/"),
	}.Build())
	s.Require().NoError(err)
	s.Equal("http://example.com/three", expandResp.GetResult())
}

func (s *GRPCServerSuite) TestExpandURLNotFound() {
	s.ensureToken()
	ctx := withAuthToken(context.Background(), s.token)

	_, err := s.client.ExpandURL(ctx, proto.URLExpandRequest_builder{
		Id: "missing-id",
	}.Build())
	s.Require().Error(err)

	st, ok := status.FromError(err)
	s.Require().True(ok)
	s.Equal(codes.NotFound, st.Code())
}

func (s *GRPCServerSuite) TestShortenURLInvalidArgument() {
	s.ensureToken()

	_, _, err := s.shorten(context.Background(), s.token, "not-a-url")
	s.Require().Error(err)

	st, ok := status.FromError(err)
	s.Require().True(ok)
	s.Equal(codes.InvalidArgument, st.Code())
}

func (s *GRPCServerSuite) TestListUserURLs() {
	s.ensureToken()
	ctx := withAuthToken(context.Background(), s.token)

	_, _, err := s.shorten(ctx, s.token, "http://example.com/list")
	s.Require().NoError(err)

	listResp, err := s.client.ListUserURLs(ctx, &emptypb.Empty{})
	s.Require().NoError(err)

	var found bool
	for _, item := range listResp.GetUrl() {
		if item.GetOriginalUrl() != "http://example.com/list" {
			continue
		}
		found = true
		s.Contains(item.GetShortUrl(), "http://localhost:3200")
	}
	s.True(found)
}

func (s *GRPCServerSuite) shorten(ctx context.Context, token, longURL string) (*proto.URLShortenResponse, metadata.MD, error) {
	if token != "" {
		ctx = withAuthToken(ctx, token)
	}

	var header metadata.MD
	resp, err := s.client.ShortenURL(
		ctx,
		proto.URLShortenRequest_builder{Url: longURL}.Build(),
		grpc.Header(&header),
	)
	return resp, header, err
}

func (s *GRPCServerSuite) ensureToken() {
	if s.token != "" {
		return
	}

	_, header, err := s.shorten(context.Background(), "", "http://example.com/bootstrap")
	s.Require().NoError(err)

	tokens := header.Get(headerAuthorization)
	s.Require().NotEmpty(tokens)
	s.token = tokens[0]
}

func withAuthToken(ctx context.Context, token string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, headerAuthorization, token)
}

func waitGRPCReady(addr string) error {
	const timeout = 5 * time.Second
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		conn, err := grpc.NewClient(
			addr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		time.Sleep(20 * time.Millisecond)
	}

	return fmt.Errorf("grpc server on %s not ready within %s", addr, timeout)
}

type seqStrGen struct {
	seq atomic.Uint64
}

func (g *seqStrGen) GetRandomString(int) string {
	return fmt.Sprintf("hash%x", g.seq.Add(1))
}
