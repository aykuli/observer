package decompressor

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"io"

	"github.com/aykuli/observer/internal/server/models"
)

type CompressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func New(r io.ReadCloser) (*CompressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &CompressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c CompressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *CompressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

// ----------------------------------

type Decompressor struct {
}

//func New() *Decompressor {
//	return &Decompressor{}
//}

func (d *Decompressor) decompress(reader io.ReadCloser) (*bytes.Buffer, error) {
	var data []byte
	_, err := reader.Read(data)
	if err != nil {
		return nil, fmt.Errorf("couldnt read bytes from req body. %+v", err)
	}

	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("couldnt read bytes qhile decompressing. %+v", err)
	}
	defer r.Close()

	var b bytes.Buffer
	_, err = b.ReadFrom(r)
	if err != nil {
		return nil, fmt.Errorf("failed decompress data. %+v", err)
	}

	return &b, nil
}

func (d *Decompressor) Metric(reqBody io.ReadCloser) (models.Metric, error) {
	var metric models.Metric

	b, err := d.decompress(reqBody)
	if err != nil {
		return metric, fmt.Errorf("couldnt read bytes qhile decompressing. %+v", err)
	}

	dec := gob.NewDecoder(b)
	err = dec.Decode(&metric)
	if err != nil {
		return metric, fmt.Errorf("failed decompress data. %+v", err)
	}

	return metric, nil
}

func (d *Decompressor) Metrics(reqBody io.ReadCloser) ([]models.Metric, error) {
	b, err := d.decompress(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed decompress data. %+v", err)
	}

	var metrics []models.Metric
	dec := gob.NewDecoder(b)
	err = dec.Decode(&metrics)
	if err != nil {
		return nil, fmt.Errorf("failed decompress data. %+v", err)
	}

	return metrics, nil
}
