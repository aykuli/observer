package storage

import (
	"bufio"
	"encoding/json"
	"io/fs"
	"os"
)

type Producer struct {
	file   *os.File
	writer *bufio.Writer
}

func NewProducer(filename string) (*Producer, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, fs.ModePerm)
	if err != nil {
		return nil, err
	}

	return &Producer{
		file:   file,
		writer: bufio.NewWriter(file),
	}, nil
}

func (p *Producer) WriteMetrics(mStore Metrics) error {
	data, err := json.Marshal(&mStore)
	if err != nil {
		return err
	}

	if _, err := p.writer.Write(data); err != nil {
		return err
	}

	if err := p.writer.WriteByte('\n'); err != nil {
		return err
	}

	return p.writer.Flush()
}

func (p *Producer) Close() error {
	return p.file.Close()
}
