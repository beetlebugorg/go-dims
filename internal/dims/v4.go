package dims

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"net/http"
	"strings"
)

func ParseAndValidV4Request(r *http.Request, config Config) (*Request, error) {
	var timestamp int32
	n, err := fmt.Sscanf(r.PathValue("timestamp"), "%d", &timestamp)
	if err != nil || n != 1 {
		timestamp = 0
	}

	h := md5.New()
	h.Write([]byte(r.PathValue("clientId")))
	h.Write([]byte(r.PathValue("commands")))
	h.Write([]byte(r.URL.Query().Get("url")))
	requestHash := fmt.Sprintf("%x", h.Sum(nil))

	request := Request{
		Id:          requestHash,
		Config:      config,
		ClientId:    r.PathValue("clientId"),
		ImageUrl:    r.URL.Query().Get("url"),
		RawCommands: r.PathValue("commands"),
		Signature:   r.PathValue("signature"),
	}

	// Validate signature
	if !config.DevelopmentMode {
		signature := Sign_md5(request)
		if bytes.Equal([]byte(signature), []byte(request.Signature)) {
			return nil, fmt.Errorf("signature mismatch")
		}
	}

	return &request, nil
}

// Sign returns a signed string using the MD5 algorithm.
func Sign_md5(r Request) string {
	sanitizedCommands := strings.ReplaceAll(r.RawCommands, " ", "+")
	sanitizedCommands = strings.Trim(sanitizedCommands, "/")
	sanitizedCommands += "/"

	hash := md5.New()
	hash.Write([]byte(fmt.Sprintf("%d", r.Timestamp)))
	hash.Write([]byte(r.Config.Signing.SigningKey))
	hash.Write([]byte(sanitizedCommands))
	hash.Write([]byte(r.ImageUrl))

	return fmt.Sprintf("%x", hash.Sum(nil))[0:7]
}
