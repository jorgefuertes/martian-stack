package validation

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

func (ct *ErrorContainer) AppendValidatorErrors(err error) {
	if err == nil {
		return
	}
	for _, e := range err.(validator.ValidationErrors) {
		msg := ErrInvalidField
		switch e.ActualTag() {
		case "oneof":
			msg = fmt.Sprintf("Sólo %s", e.Param())
		}
		ct.Append(strings.ToLower(e.Field()), msg)
	}
}
