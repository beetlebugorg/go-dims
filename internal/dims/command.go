// Copyright 2024 Jeremy Collins. All rights reserved.
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

package dims

import (
	"gopkg.in/gographics/imagick.v3/imagick"
	"log/slog"
	"strings"
)

type Operation func(mw *imagick.MagickWand, args string) error

type Command struct {
	Name      string
	Args      string
	Operation Operation
}

func ParseCommands(cmds string, operations map[string]Operation) []Command {
	commands := make([]Command, 0)
	parsedCommands := strings.Split(strings.Trim(cmds, "/"), "/")
	for i := 0; i < len(parsedCommands)-1; i += 2 {
		command := parsedCommands[i]
		args := parsedCommands[i+1]

		commands = append(commands, Command{
			Name:      command,
			Args:      args,
			Operation: operations[command],
		})

		slog.Info("parsedCommand", "command", command, "args", args)
	}

	return commands
}
