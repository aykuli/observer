package file

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io/fs"
	"os"

	"github.com/aykuli/observer/internal/server/storage"
)

// Consumer reads data from file in JSON
type Consumer struct {
	file    *os.File
	scanner *bufio.Scanner
}

func NewConsumer(filename string) (*Consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, fs.ModePerm)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		file:    file,
		scanner: bufio.NewScanner(file),
	}, nil
}

func (c *Consumer) ReadMetrics() (*storage.MemStorage, error) {
	c.scanner.Split(scanLastNonEmptyLine)
	var line string
	for c.scanner.Scan() {
		line = c.scanner.Text()
	}

	if len(line) < 1 {
		return nil, nil
	}

	defer c.file.Close()

	var mStore *storage.MemStorage
	err := json.Unmarshal([]byte(line), &mStore)
	if err != nil {
		return nil, err
	}

	return mStore, nil
}

func (c *Consumer) Close() error {
	return c.file.Close()
}

// source https://gist.github.com/keegancsmith/54d2325e7a3c6eb78276c884c4208aa6
func scanLastNonEmptyLine(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// Set advance to after our last line
	if atEOF {
		advance = len(data)
	} else {
		// data[advance:] now contains a possibly incomplete line
		advance = bytes.LastIndexAny(data, "\n\r") + 1
	}
	data = data[:advance]

	// Remove empty lines (strip EOL chars)
	data = bytes.TrimRight(data, "\n\r")

	// We have no non-empty lines, so advance but do not return a token.
	if len(data) == 0 {
		return advance, nil, nil
	}

	token = data[bytes.LastIndexAny(data, "\n\r")+1:]
	return advance, token, nil
}
