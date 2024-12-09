package dims

import (
	"strings"
)

func FormatCommand(request *Request, args string) error {
	format := strings.ToLower(args)
	request.format = &format
	return nil
}
