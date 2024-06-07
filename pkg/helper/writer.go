package helper

import (
	"io"
	"slices"
)

type Writer struct {
	lines [][]byte
}

func NewWriter() *Writer {
	return &Writer{lines: make([][]byte, 0)}
}

func (w *Writer) Write(data []byte) (n int, err error) {
	if data[len(data)-1] == '\n' {
		data = slices.Delete(data, len(data)-1, len(data))
	}
	w.lines = append(w.lines, data)
	return len(data), nil
}

// read lines one by one, deleting them
func (w *Writer) Read() ([]byte, error) {
	if len(w.lines) == 0 {
		return []byte{}, io.EOF
	}
	line := w.lines[0]
	w.lines = slices.Delete(w.lines, 0, 1)
	return line, nil
}

func (w *Writer) ReadString() (string, error) {
	b, err := w.Read()
	return string(b), err
}

func (w *Writer) Reset() {
	w.lines = make([][]byte, 0)
}

func (w *Writer) Len() int { return len(w.lines) }
