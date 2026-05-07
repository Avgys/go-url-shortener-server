package repository

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"go-url-shortener/internal/model"
)

var (
	ErrFileClosed = errors.New("file closed, read and write are forbidden")
)

type FileStore struct {
	store    *InMemoryStore
	file     *os.File
	appender *gocsv.SafeCSVWriter
	logger   *zerolog.Logger
	isClosed atomic.Bool

	closeOnce sync.Once
}

func NewFileStore(ctx context.Context, filename string, logger *zerolog.Logger) (*FileStore, error) {
	file, err := openFile(filename)

	if err != nil {
		return nil, err
	}

	records, err := readRows(file)

	if err != nil {
		return nil, err
	}

	writer := csv.NewWriter(file)
	gocsvWriter := gocsv.NewSafeCSVWriter(writer)

	rep := &FileStore{file: file, store: NewInMemoryStore(records), appender: gocsvWriter, logger: logger}

	go func() {
		<-ctx.Done()

		rep.Close()
	}()

	return rep, nil
}

func readRows(file *os.File) ([]*model.DBURL, error) {
	var records []*model.DBURL

	reader := csv.NewReader(file)

	err := gocsv.UnmarshalCSV(reader, &records)
	if err != nil && !errors.Is(err, gocsv.ErrEmptyCSVFile) {
		return nil, err
	}

	return records, nil
}

func openFile(path string) (*os.File, error) {

	if !filepath.IsAbs(path) {
		var err error
		path, err = filepath.Abs(path)

		if err != nil {
			err = fmt.Errorf("error getting path to store file, %w", err)
			return nil, err
		}
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, os.ModePerm)

	if err != nil {
		err = fmt.Errorf("error opening store file, %w", err)
		return nil, err
	}

	return file, err
}

func clearFile(file *os.File) error {
	if err := file.Truncate(0); err != nil {
		return fmt.Errorf("failed to truncate file: %w", err)
	}
	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek file: %w", err)
	}
	return nil
}

func (fs *FileStore) Close() error {

	var err error = nil

	fs.closeOnce.Do(func() {

		fs.isClosed.Store(true)

		if err = clearFile(fs.file); err != nil {
			fs.logger.Err(err).Send()
			return
		}

		s := fs.getAll()

		if err = fs.append(s); err != nil {
			fs.logger.Err(err).Send()
			return
		}

		if err = fs.file.Sync(); err != nil {
			fs.logger.Err(err).Send()
			return
		}

		if err = fs.file.Close(); err != nil {
			fs.logger.Err(err).Send()
			return
		}
	})

	return err
}

func (fs *FileStore) StoreBatch(ctx context.Context, input Full2ShortBatch, userID int64) (retryToInsert []string, alreadyExists map[string]string, err error) {
	if fs.isClosed.Load() {
		err = ErrFileClosed
		return
	}

	retryToInsert, alreadyExists, err = fs.store.StoreBatch(ctx, input, userID)

	if err != nil {
		return
	}

	urlsToSave := lo.FilterMapToSlice(input, func(origin string, short string) (model.DBURL, bool) {
		if _, exists := alreadyExists[origin]; exists {
			return model.DBURL{}, false
		}

		if lo.Contains(retryToInsert, origin) {
			return model.DBURL{}, false
		}

		return model.DBURL{OriginalURL: origin, ShortURL: short, UserID: userID, CreatedAt: time.Now().UTC()}, true
	})

	if err = fs.append(urlsToSave); err != nil {
		return
	}

	return retryToInsert, alreadyExists, err
}

func (fs *FileStore) ResolveShortURL(ctx context.Context, shortURL string) (model.DBURL, error) {
	if fs.isClosed.Load() {
		return model.DBURL{}, ErrFileClosed
	}

	return fs.store.ResolveShortURL(ctx, shortURL)
}

func (fs *FileStore) TestConnection(ctx context.Context) error {
	return nil
}

func (fs *FileStore) append(records []model.DBURL) error {
	pos, err := fs.file.Seek(0, io.SeekEnd)

	if err != nil {
		return err
	}

	if pos == 0 {
		err = gocsv.MarshalCSV(records, fs.appender)
	} else {
		err = gocsv.MarshalCSVWithoutHeaders(records, fs.appender)
	}

	return err
}

func (fs *FileStore) getAll() []model.DBURL {
	return fs.store.getAll()
}

func (fs *FileStore) GetURLsByUserID(ctx context.Context, userID int64) ([]model.DBURL, error) {
	return fs.store.GetURLsByUserID(ctx, userID)
}

func (fs *FileStore) DeleteURLS(context context.Context, groupedByUser map[int64][]string) ([]model.DBURL, error) {

	deleted, err := fs.store.DeleteURLS(context, groupedByUser)

	if err != nil {
		return nil, err
	}

	if err = fs.append(deleted); err != nil {
		return nil, err
	}

	return deleted, err
}
