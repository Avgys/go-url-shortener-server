package repository

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Avgys/go-url-shortener-server/internal/model"
	"github.com/gocarina/gocsv"
)

var (
	ErrFileClosed = errors.New("file closed, read and write are forbidden")
)

type FileStore struct {
	store    *InMemoryStore
	file     *os.File
	appender *gocsv.SafeCSVWriter
	isClosed bool
}

func NewFileStore(filename string) (*FileStore, error) {
	path := filename

	if !filepath.IsAbs(path) {
		var err error
		path, err = filepath.Abs(path)

		if err != nil {
			err = fmt.Errorf("error getting path to store file, %w", err)
			return nil, err
		}
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)

	if err != nil {
		err = fmt.Errorf("error opening store file, %w", err)
		return nil, err
	}

	records := make([]*model.DBURL, 0)

	if err := gocsv.Unmarshal(file, &records); err != nil {
		panic(err)
	}

	writer := csv.NewWriter(file)
	gocsvWriter := gocsv.NewSafeCSVWriter(writer)

	return &FileStore{file: file, store: NewInMemoryStore(records), appender: gocsvWriter}, nil
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
	if fs.isClosed {
		return errors.New("filestore closed")
	}

	if err := clearFile(fs.file); err != nil {
		return err
	}

	s := fs.getAll()

	for _, v := range s {
		fs.append(v)
	}

	fs.file.Sync()

	fs.isClosed = true
	return fs.file.Close()
}

func (fs *FileStore) StoreBatch(ctx context.Context, input Full2ShortBatch, userID int64) (retryToInsert []string, alreadyExists map[string]string, err error) {
	if fs.isClosed {
		err = ErrFileClosed
		return
	}

	for fullURL, shortURL := range input {
		if err = fs.append(&model.DBURL{OriginalURL: fullURL, ShortURL: shortURL}); err != nil {
			return
		}
	}

	return fs.store.StoreBatch(ctx, input, userID)
}

func (fs *FileStore) ResolveShortURL(ctx context.Context, shortURL string) (string, error) {
	if fs.isClosed {
		return "", ErrFileClosed
	}

	return fs.store.ResolveShortURL(ctx, shortURL)
}

func (fs *FileStore) TestConnection(ctx context.Context) error {
	return nil
}

func (fs *FileStore) append(record *model.DBURL) error {
	pos, err := fs.file.Seek(0, io.SeekCurrent)

	if err != nil {
		return err
	}

	if pos == 0 {
		err = gocsv.MarshalCSV(record, fs.appender)
	} else {
		err = gocsv.MarshalCSVWithoutHeaders(record, fs.appender)
	}

	return err
}

func (fs *FileStore) getAll() []*model.DBURL {
	return fs.store.getAll()
}

func (s *FileStore) GetURLsByUserId(ctx context.Context, userID int64) ([]*model.DBURL, error) {
	return s.store.GetURLsByUserId(ctx, userID)
}
