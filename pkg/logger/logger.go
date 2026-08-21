package logger

import (
	"context"
	"io"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

const (
	archivePeriod = 12 * time.Hour
)

var (
	DefaultLogger zerolog.Logger

	archiveJobsMu   sync.Mutex
	archiveJobsDirs = make(map[string]struct{})
)

func init() {
	DefaultLogger = zerolog.New(os.Stdout).Level(zerolog.InfoLevel).With().Timestamp().Logger()
}

func createLogger(service string, isDebug bool, w io.Writer) *zerolog.Logger {
	level := zerolog.InfoLevel
	if isDebug {
		level = zerolog.DebugLevel
	}

	log := zerolog.New(w).Level(level).With().Timestamp().Logger()

	if isDebug {
		colorWr := zerolog.ConsoleWriter{
			Out:        w,
			NoColor:    false,
			TimeFormat: time.RFC3339,
		}
		log = zerolog.New(colorWr).Level(level).With().Timestamp().Caller().Logger()
	}

	log.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Str("service", service)
	})

	return &log
}

func NewLogger(service string, isDebug bool, w io.Writer) *zerolog.Logger {
	return createLogger(service, isDebug, w)
}

func NewLoggerWithFileOut(service string, isDebug bool, filePath string) (*zerolog.Logger, error) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	w := io.MultiWriter(os.Stdout, file)
	return createLogger(service, isDebug, w), nil
}

func RunArchivateJob(ctx context.Context, dir string, days int) {
	archiveJobsMu.Lock()
	if _, running := archiveJobsDirs[dir]; running {
		archiveJobsMu.Unlock()
		return
	}
	archiveJobsDirs[dir] = struct{}{}
	archiveJobsMu.Unlock()

	if err := makeArchive(dir, days); err != nil {
		DefaultLogger.Error().Err(err).Msg("initial archivation failed")
	} else {
		DefaultLogger.Info().Msg("successfully make archive")
	}

	ticker := time.NewTicker(archivePeriod)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				archiveJobsMu.Lock()
				delete(archiveJobsDirs, dir)
				archiveJobsMu.Unlock()
				DefaultLogger.Info().Msg("archivation job stopped")
				return
			case <-ticker.C:
				if err := makeArchive(dir, days); err != nil {
					DefaultLogger.Error().Err(err).Msg("archivation failed")
				} else {
					DefaultLogger.Info().Msg("archivation job successfully make archive")
				}
			}
		}
	}()
}
