package helper

import (
	"encoding/json"
	"errors"
	"io"
	"slices"
	"sync"
)

type Writer struct {
	lines [][]byte
	lock  *sync.Mutex
}

func NewWriter() *Writer {
	return &Writer{lines: make([][]byte, 0), lock: &sync.Mutex{}}
}

func (w *Writer) Write(data []byte) (n int, err error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	if data[len(data)-1] == '\n' {
		data = slices.Delete(data, len(data)-1, len(data))
	}
	w.lines = append(w.lines, data)
	return len(data), nil
}

func (w *Writer) WriteJSON(obj any) (n int, err error) {
	b, err := json.Marshal(obj)
	if err != nil {
		return 0, err
	}

	return w.Write(b)
}

// read lines one by one, deleting them
func (w *Writer) Read() ([]byte, error) {
	w.lock.Lock()
	defer w.lock.Unlock()

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

func (w *Writer) ReadJSON(dest any) error {
	b, err := w.Read()
	if err != nil {
		return err
	}

	if len(b) == 0 {
		return nil
	}

	if err := json.Unmarshal(b, dest); err != nil {
		return errors.Join(err, errors.New(string(b)))
	}

	return nil
}

func (w *Writer) Reset() {
	w.lock.Lock()
	defer w.lock.Unlock()

	w.lines = make([][]byte, 0)
}

func (w *Writer) Len() int {
	w.lock.Lock()
	defer w.lock.Unlock()

	return len(w.lines)
}
