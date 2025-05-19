package v4

import (
	"crypto/md5"
	"fmt"
	"github.com/beetlebugorg/go-dims/internal/core"
	dims "github.com/beetlebugorg/go-dims/internal/http"
	"log/slog"
	"net/http"
	"strings"
)

type Request struct {
	*dims.Request

	clientId  string
	timestamp string
}

func NewRequest(r *http.Request, w http.ResponseWriter, config core.Config) (*Request, error) {
	clientId := r.PathValue("clientId")
	timestamp := r.PathValue("timestamp")

	request, err := dims.NewRequest(r, w, config)
	if err != nil {
		return nil, err
	}

	request.Signature = r.PathValue("signature")

	return &Request{
		Request: request,

		clientId:  clientId,
		timestamp: timestamp,
	}, nil
}

func (v4 *Request) HashId() string {
	h := md5.New()
	h.Write([]byte(v4.clientId))
	h.Write([]byte(v4.RawCommands))
	h.Write([]byte(v4.ImageUrl))

	return fmt.Sprintf("%x", h.Sum(nil))
}

func (v4 *Request) Validate() bool {
	expectedSignature := v4.sign(v4.RawCommands, v4.timestamp, v4.ImageUrl, v4.SignedParams, v4.Config().SigningKey)

	if expectedSignature == v4.Signature {
		return true
	}

	slog.Error("verifySignature failed.",
		"expected", expectedSignature,
		"got", v4.Signature)

	return false
}

func (v4 *Request) sign(commands, timestamp, imageUrl string, signedParams map[string]string, signingKey string) string {
	h := md5.New()
	h.Write([]byte(timestamp))
	h.Write([]byte(signingKey))
	h.Write([]byte(commands))
	h.Write([]byte(imageUrl))

	for _, signedParam := range signedParams {
		h.Write([]byte(signedParam))
	}

	return fmt.Sprintf("%x", h.Sum(nil))[0:7]
}

func (v4 *Request) SignedUrl() string {
	signature := v4.sign(v4.RawCommands, v4.timestamp, v4.ImageUrl, v4.SignedParams, v4.Config().SigningKey)

	unsignedPath := fmt.Sprintf("/dims4/%s/%s/%s", v4.clientId, v4.Signature, v4.timestamp)
	signedPath := fmt.Sprintf("/dims4/%s/%s/%s", v4.clientId, signature, v4.timestamp)

	u := v4.URL
	u.Path = strings.Replace(u.Path, unsignedPath, signedPath, 1)

	return u.String()
}
