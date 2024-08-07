// Package compressor provides middleware for zipping data.
package compressor

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type compressWriter struct {
	w  http.ResponseWriter
	Zw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{w: w, Zw: gzip.NewWriter(w)}
}

// Header return writer header as it is.
func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

// Write write provided bytes with zipping.
func (c *compressWriter) Write(p []byte) (int, error) {
	return c.Zw.Write(p)
}

// WriteHeader sets headers Content-Encoding value to gzip.
func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}

	c.w.WriteHeader(statusCode)
}

type compressReader struct {
	r  io.ReadCloser
	Zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		Zr: zr,
	}, nil
}

// Read provides reading through gzip.Reader.
func (c *compressReader) Read(p []byte) (n int, err error) {
	return c.Zr.Read(p)
}

// Close closes gzip.Reader.
func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.Zr.Close()
}

// GzipMiddleware handles Accept-Encoding, Content-Encoding and Content-Type header values.
func GzipMiddleware(h http.Handler) http.Handler {
	gzipFn := func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "swag") || strings.Contains(r.URL.Path, "doc") {
			h.ServeHTTP(w, r)
			return
		}

		ow := w
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			cw := newCompressWriter(w)
			ow = cw
			defer cw.Zw.Close()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		jsonType := r.Header.Get("Content-Type")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		jsonContent := strings.Contains(jsonType, "application/json")

		if sendsGzip && jsonContent {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				http.Error(w, "cannot decode json request body", http.StatusInternalServerError)
				return
			}

			r.Body = cr
			defer cr.Zr.Close()
		}

		h.ServeHTTP(ow, r)
	}

	return http.HandlerFunc(gzipFn)
}

// Compress returns gzipped bytes.
func Compress(input []byte) ([]byte, error) {
	var b bytes.Buffer
	w, err := gzip.NewWriterLevel(&b, gzip.BestCompression)
	if err != nil {
		return nil, fmt.Errorf("failed init compress writer. %v", err)
	}

	_, err = w.Write(input)
	if err != nil {
		return nil, fmt.Errorf("failed write data to compress temporary buffer. %v", err)
	}

	err = w.Close()
	if err != nil {
		return nil, fmt.Errorf("failed compress data. %v", err)
	}

	return b.Bytes(), nil
}
