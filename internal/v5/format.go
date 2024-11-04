package v5

import (
	"strings"
)

func FormatCommand(request *RequestV5, args string) error {
	format := strings.ToLower(args)
	request.format = &format
	return nil
}
