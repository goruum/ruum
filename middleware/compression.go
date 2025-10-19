package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/goruum/ruum/core"
)

// CompressionConfig holds compression configuration
type CompressionConfig struct {
	Level        int // Compression level (gzip.DefaultCompression, gzip.BestSpeed, gzip.BestCompression)
	MinLength    int // Minimum response size to compress (bytes)
	SkipFunc     func(core.Context) bool
	ContentTypes []string // Content types to compress
}

// DefaultCompressionConfig returns default configuration
func DefaultCompressionConfig() CompressionConfig {
	return CompressionConfig{
		Level:     gzip.DefaultCompression,
		MinLength: 1024, // 1KB
		SkipFunc: func(_ core.Context) bool {
			return false
		},
		ContentTypes: []string{
			"text/html",
			"text/css",
			"text/plain",
			"text/javascript",
			"application/javascript",
			"application/json",
			"application/xml",
			"text/xml",
		},
	}
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
	wroteHeader bool
}

func (w *gzipResponseWriter) WriteHeader(code int) {
	if !w.wroteHeader {
		w.ResponseWriter.Header().Del("Content-Length")
		w.wroteHeader = true
		w.ResponseWriter.WriteHeader(code)
	}
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.Writer.Write(b)
}

var gzipPool = sync.Pool{
	New: func() interface{} {
		w, _ := gzip.NewWriterLevel(io.Discard, gzip.DefaultCompression)
		return w
	},
}

// Compression creates a compression middleware
func Compression(config CompressionConfig) core.MiddlewareFunc {
	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(ctx core.Context) error {
			if config.SkipFunc(ctx) {
				return next(ctx)
			}

			// Check if client accepts gzip
			if !strings.Contains(ctx.Header("Accept-Encoding"), "gzip") {
				return next(ctx)
			}

			// Get gzip writer from pool
			gz := gzipPool.Get().(*gzip.Writer)
			defer gzipPool.Put(gz)
			gz.Reset(ctx.Response())
			defer func() {
				_ = gz.Close()
			}()

			ctx.SetHeader("Content-Encoding", "gzip")
			ctx.SetHeader("Vary", "Accept-Encoding")

			// Wrap response writer
			gzw := &gzipResponseWriter{Writer: gz, ResponseWriter: ctx.Response()}

			// Create new context with wrapped response
			wrappedCtx := core.NewContext(
				ctx,
				ctx.Request(),
				gzw,
				ctx.Container(),
			)

			return next(wrappedCtx)
		}
	}
}
