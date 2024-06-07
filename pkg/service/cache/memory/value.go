package memory

import (
	"fmt"
	"strconv"
)

type value string

func (v value) String() string {
	return string(v)
}

func newValue(v any) value {
	return value(fmt.Sprintf("%v", v))
}

func (v value) int() (int, error) {
	i, err := strconv.ParseInt(v.String(), 10, 64)
	if err != nil {
		return 0, err
	}

	return int(i), nil
}

func (v value) float64() (float64, error) {
	i, err := strconv.ParseFloat(v.String(), 64)
	if err != nil {
		return 0, err
	}

	return i, nil
}

func (v value) bytes() []byte {
	return []byte(v.String())
}
