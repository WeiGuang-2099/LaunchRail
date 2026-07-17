package domain

import (
	"errors"
	"strings"
)

var ErrArchivedService = errors.New("service is archived")

type FieldViolation struct {
	Field   string
	Message string
}

type ValidationErrors struct {
	Violations []FieldViolation
}

func (e *ValidationErrors) Error() string {
	messages := make([]string, 0, len(e.Violations))
	for _, violation := range e.Violations {
		messages = append(messages, violation.Field+": "+violation.Message)
	}
	return strings.Join(messages, "; ")
}

func (e *ValidationErrors) add(field, message string) {
	e.Violations = append(e.Violations, FieldViolation{Field: field, Message: message})
}

func (e *ValidationErrors) orNil() error {
	if len(e.Violations) == 0 {
		return nil
	}
	return e
}
