package repository

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
)

var (
	ErrFileClosed = errors.New("file closed, read and write are forbidden")
)

type FileStore struct {
	store    Repository
	file     *os.File
	appender *json.Encoder
	isClosed bool
}

type record struct {
	ShortURL string `json:"short_url"`
	FullURL  string `json:"full_url"`
}

func NewFileStore(filename string) (*FileStore, error) {
	path := filename

	if !filepath.IsAbs(path) {

		var err error
		path, err = filepath.Abs(path)

		if err != nil {
			return nil, err
		}
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)

	if err != nil {
		return nil, err
	}

	records := make(storage, 0)
	dec := json.NewDecoder(file)

	for {
		var r record
		err := dec.Decode(&r)

		if err != nil {

			if errors.Is(err, io.EOF) {
				break
			}

			return nil, err
		}

		records[r.ShortURL] = r.FullURL
	}

	appender := json.NewEncoder(file)
	appender.SetEscapeHTML(false)

	return &FileStore{file: file, store: NewStore(records), appender: appender}, nil
}

func clearFile(file *os.File) {
	file.Truncate(0)
	file.Seek(0, 0)
}

func (fs *FileStore) Close() error {
	if fs.isClosed {
		return nil
	}

	clearFile(fs.file)

	s := fs.getAll()

	for k, v := range s {
		fs.append(&record{ShortURL: k, FullURL: v})
	}

	fs.file.Sync()

	fs.isClosed = true
	return fs.file.Close()
}

func (fs *FileStore) StoreURL(url string, shortURL string) error {
	if fs.isClosed {
		return ErrFileClosed
	}

	fs.append(&record{FullURL: url, ShortURL: shortURL})

	return fs.store.StoreURL(url, shortURL)
}

func (fs *FileStore) ResolveShortURL(shortURL string) (string, error) {
	if fs.isClosed {
		return "", ErrFileClosed
	}

	return fs.store.ResolveShortURL(shortURL)
}

func (fs *FileStore) append(record *record) {
	fs.appender.Encode(record)
}

func (fs *FileStore) getAll() storage {
	return fs.store.getAll()
}
