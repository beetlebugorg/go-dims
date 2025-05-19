package commands

import (
	"fmt"
	"github.com/beetlebugorg/go-dims/internal/core"
)

type OperationError struct {
	core.StatusError
	Command string
	Args    string
}

func NewOperationError(command string, args string, message string) *OperationError {
	return &OperationError{
		StatusError: *core.NewStatusError(400, message),
		Command:     command,
		Args:        args,
	}
}

func (e *OperationError) Error() string {
	return fmt.Sprintf("OperationError: %s (status: %d) (command: %s) (args: %s)", e.Message, e.StatusCode, e.Command, e.Args)
}
