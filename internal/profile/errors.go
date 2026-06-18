package profile

import "fmt"

// ValidationError is a field-scoped validation failure for UI highlighting.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return ""
	}
	return v[0].Error()
}

func (v ValidationErrors) HasErrors() bool {
	return len(v) > 0
}
