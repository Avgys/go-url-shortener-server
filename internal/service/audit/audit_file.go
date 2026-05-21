package audit

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

	"github.com/gocarina/gocsv"
	"github.com/rs/zerolog"
)

var (
	ErrFileClosed = errors.New("file closed, read and write are forbidden")
)

type AuditFile struct {
	ID int

	filepath string
	file     *os.File
	appender *csv.Writer
	logger   *zerolog.Logger
	isClosed atomic.Bool

	closeOnce sync.Once
}

func NewAuditFile(ctx context.Context, id int, filepath string, logger *zerolog.Logger) *AuditFile {

	newLogger := logger.With().
		Str("audit", "file").
		Str("audit_file_path", filepath).
		Logger()

	return &AuditFile{ID: id, filepath: filepath, logger: &newLogger}
}

func (a *AuditFile) GetID() int {
	return a.ID
}

func (a *AuditFile) Update(event AuditEvent) (err error) {

	if a.appender == nil {
		a.file, a.appender, err = openFile(a.filepath)
		if err != nil {
			a.logger.Err(err).Send()
			return
		}
	}

	err = a.append(event)
	if err != nil {
		a.logger.Err(err).Send()
	}

	return
}

func openFile(path string) (*os.File, *csv.Writer, error) {

	if !filepath.IsAbs(path) {
		var err error
		path, err = filepath.Abs(path)

		if err != nil {
			err = fmt.Errorf("error getting path to file, %w", err)
			return nil, nil, err
		}
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, os.ModePerm)

	if err != nil {
		err = fmt.Errorf("error opening file, %w", err)
		return nil, nil, err
	}

	writer := csv.NewWriter(file)

	return file, writer, nil
}

func (a *AuditFile) append(event AuditEvent) error {
	pos, err := a.file.Seek(0, io.SeekEnd)

	if err != nil {
		return err
	}

	input := []AuditEvent{event}

	if pos == 0 {
		err = gocsv.MarshalCSV(input, a.appender)
	} else {
		err = gocsv.MarshalCSVWithoutHeaders(input, a.appender)
	}

	return err
}

func (a *AuditFile) Close() error {

	var err error = nil

	a.closeOnce.Do(func() {

		a.isClosed.Store(true)

		if err = a.file.Sync(); err != nil {
			a.logger.Err(err).Send()
			return
		}

		if err = a.file.Close(); err != nil {
			a.logger.Err(err).Send()
			return
		}
	})

	return err
}
