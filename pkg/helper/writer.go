package helper

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"sync"
)

var ErrNullBytes = errors.New("null bytes detected in input")

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

	fmt.Printf("Data received (%d bytes): '%s'\n", len(data), string(data))

	if len(data) == 0 {
		return 0, nil
	}

	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)

	if dataCopy[len(dataCopy)-1] == '\n' {
		dataCopy = dataCopy[:len(dataCopy)-1]
	}

	if bytes.Contains(dataCopy, []byte{0}) {
		return 0, ErrNullBytes
	}

	w.lines = append(w.lines, dataCopy)
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

	// Hacer una copia profunda de la línea
	line := make([]byte, len(w.lines[0]))
	copy(line, w.lines[0])

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
