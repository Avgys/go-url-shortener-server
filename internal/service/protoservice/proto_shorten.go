package protoservice

import (
	"context"
	"errors"
	"net/http"

	"go-url-shortener/internal/logger"
	"go-url-shortener/internal/service/auth"
	"go-url-shortener/internal/service/shortifier"
	httphelper "go-url-shortener/internal/shared/http"
	grpcstatus "go-url-shortener/internal/shared/grpcstatus"
	"go-url-shortener/proto"

	"google.golang.org/protobuf/types/known/emptypb"
)

type ShortenerServiceServer struct {
	proto.UnimplementedShortenerServiceServer
	Shortifier *shortifier.Shortifier
}

func NewShortenerServiceServer(shortifier *shortifier.Shortifier) *ShortenerServiceServer {
	return &ShortenerServiceServer{
		UnimplementedShortenerServiceServer: proto.UnimplementedShortenerServiceServer{},
		Shortifier:                          shortifier,
	}
}

func (s *ShortenerServiceServer) ShortenURL(ctx context.Context, req *proto.URLShortenRequest) (*proto.URLShortenResponse, error) {
	traceLogger := logger.FromContext(ctx, logger.GetFuncName())
	
	resultURL, err := s.Shortifier.ShortenURL(ctx, req.GetUrl(), traceLogger)
	if err != nil {
		return nil, grpcstatus.FromError(err)
	}

	return proto.URLShortenResponse_builder{
		Result: resultURL.ShortURL,
	}.Build(), nil
}

func (s *ShortenerServiceServer) ExpandURL(ctx context.Context, req *proto.URLExpandRequest) (*proto.URLExpandResponse, error) {
	traceLogger := logger.FromContext(ctx, logger.GetFuncName())

	dbURL, err := s.Shortifier.ResolveShortURL(ctx, req.GetId(), traceLogger)
	if err != nil {
		return nil, grpcstatus.FromError(err)
	}

	return proto.URLExpandResponse_builder{
		Result: dbURL.OriginalURL,
	}.Build(), nil
}

func (s *ShortenerServiceServer) ListUserURLs(ctx context.Context, _ *emptypb.Empty) (*proto.UserURLsResponse, error) {
	claims, err := auth.GetFromContext(ctx)
	if err != nil {
		return nil, grpcstatus.FromError(httphelper.NewError("unauthorized/broken token", http.StatusUnauthorized))
	}

	traceLogger := logger.FromContext(ctx, logger.GetFuncName())

	urls, err := s.Shortifier.GetURLsByUserID(ctx, claims.UserID, traceLogger)
	if err != nil {
		var httpErr *httphelper.ShowHTTPError
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNoContent {
			urls = nil
		} else {
			return nil, grpcstatus.FromError(err)
		}
	}

	urlData := make([]*proto.URLData, 0, len(urls))
	for _, url := range urls {
		urlData = append(urlData, proto.URLData_builder{
			ShortUrl:    url.ShortURL,
			OriginalUrl: url.OriginalURL,
		}.Build())
	}

	return proto.UserURLsResponse_builder{
		Url: urlData,
	}.Build(), nil
}