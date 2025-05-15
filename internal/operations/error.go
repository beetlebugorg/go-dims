// Copyright 2025 Jeremy Collins. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package operations

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
