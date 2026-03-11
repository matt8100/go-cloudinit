package cloudinit

import (
	"fmt"
	"strings"
)

// ValidationError describes a single validation failure at a field path.
type ValidationError struct {
	Path    string
	Message string
}

// Error implements the error interface.
func (e ValidationError) Error() string {
	if e.Path == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Path, e.Message)
}

// ValidationErrors collects multiple validation failures.
type ValidationErrors []ValidationError

// Error implements the error interface.
func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "validation failed"
	}
	lines := make([]string, 0, len(e))
	for _, item := range e {
		lines = append(lines, item.Error())
	}
	return strings.Join(lines, "\n")
}

func (e *ValidationErrors) add(path, message string) {
	*e = append(*e, ValidationError{Path: path, Message: message})
}

// Err returns nil when the collection is empty, or the collection itself.
func (e ValidationErrors) Err() error {
	if len(e) == 0 {
		return nil
	}
	return e
}
