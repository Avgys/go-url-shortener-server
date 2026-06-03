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

// AuditFile appends audit events as CSV rows to a local file.
type AuditFile struct {
	filepath string
	file     *os.File
	appender *csv.Writer
	logger   *zerolog.Logger
	isClosed atomic.Bool

	closeOnce sync.Once
	openOnce  sync.Once
	appendMux sync.Mutex
}

// NewAuditFile creates a file observer. The file is opened lazily on the first Update.
func NewAuditFile(ctx context.Context, filepath string, logger *zerolog.Logger) *AuditFile {

	newLogger := logger.With().
		Str("audit", "file").
		Str("audit_file_path", filepath).
		Logger()

	a := &AuditFile{filepath: filepath, logger: &newLogger}

	go func() {
		<-ctx.Done()

		_ = a.Close()
	}()

	return a
}

// Update appends event as a CSV row, writing a header row when the file is new.
func (a *AuditFile) Update(ctx context.Context, event AuditEvent) (err error) {
	if err := ctx.Err(); err != nil {
		return err
	}
	if a.appender == nil {
		a.openOnce.Do(func() {
			a.file, a.appender, err = openFile(a.filepath)
		})

		if err != nil {
			a.logger.Err(err).Send()
			return
		}
	}

	if a.isClosed.Load() {
		return errors.New("file closed")
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

	a.appendMux.Lock()
	defer a.appendMux.Unlock()

	pos, err := a.file.Seek(0, io.SeekEnd)

	if err != nil {
		return err
	}

	input := []AuditEvent{event}

	var appendFn func(in any, out gocsv.CSVWriter) (err error)

	if pos == 0 {
		appendFn = gocsv.MarshalCSV
	} else {
		appendFn = gocsv.MarshalCSVWithoutHeaders
	}

	if err := appendFn(input, a.appender); err != nil {
		a.logger.Err(err).Send()
		return err
	}

	a.appender.Flush()

	if err := a.appender.Error(); err != nil {
		a.logger.Err(err).Send()
		return err
	}

	return nil
}

// Close syncs and closes the underlying file. Safe to call more than once.
func (a *AuditFile) Close() (err error) {

	a.closeOnce.Do(func() {

		a.isClosed.Store(true)

		if a.appender == nil {
			return
		}

		a.appender.Flush()
		if err := a.appender.Error(); err != nil {
			a.logger.Err(err).Send()
			return
		}

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
