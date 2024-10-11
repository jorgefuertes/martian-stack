package validation

import (
	"fmt"
)

type Error struct {
	Field string `json:"field"`
	Msg   string `json:"msg"`
}

func (e Error) Error() string {
	return e.Msg
}

type ErrorContainer struct {
	Errors []Error `json:"errors"`
}

func NewErrorContainer() *ErrorContainer {
	return new(ErrorContainer)
}

func (e *ErrorContainer) Append(field string, err any) {
	e.Errors = append(e.Errors, Error{Field: field, Msg: fmt.Sprintf("%s", err)})
}

func (e *ErrorContainer) HasErrors() bool {
	return len(e.Errors) > 0
}

func (e *ErrorContainer) FieldHasError(field string) bool {
	for _, cur := range e.Errors {
		if cur.Field == field {
			return true
		}
	}
	return false
}

func (e *ErrorContainer) GetError(field string) string {
	for _, cur := range e.Errors {
		if cur.Field == field {
			return cur.Msg
		}
	}
	return fmt.Sprintf("no errors for field %s", field)
}

func (e *ErrorContainer) GetErrorMap() map[string][]string {
	m := make(map[string][]string, 0)
	for _, cur := range e.Errors {
		m[cur.Field] = append(m[cur.Field], cur.Msg)
	}
	return m
}
