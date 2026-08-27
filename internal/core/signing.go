package core

import "strings"

type Signer interface {
	SignedUrl() string
}

// SignedMessage returns the string a signature covers. It holds one field per
// line: the command path, the image URL, then the canonical query.
//
// A line break cannot appear inside a field. The query is percent-encoded,
// and a control character in the command path or the image URL is refused
// before the request is signed or validated. So one field can never stand in
// for two.
func SignedMessage(commands string, imageUrl string, signedQuery string) string {
	return strings.Join([]string{commands, imageUrl, signedQuery}, "\n")
}
