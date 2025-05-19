package v5

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"github.com/beetlebugorg/go-dims/internal/core"
	dims "github.com/beetlebugorg/go-dims/internal/http"
	"log/slog"
	"net/http"
)

type Request struct {
	*dims.Request
}

func NewRequest(r *http.Request, w http.ResponseWriter, config core.Config) (*Request, error) {
	request, err := dims.NewRequest(r, w, config)
	if err != nil {
		return nil, err
	}

	request.Signature = r.URL.Query().Get("sig")

	return &Request{
		Request: request,
	}, nil
}

// Validate verifies the signature of the image resize is valid.
func (v5 *Request) Validate() bool {
	expectedSignature := v5.sign(v5.ImageUrl, v5.SignedParams, v5.RawCommands, v5.Config().SigningKey)

	gotSignature, err := hex.DecodeString(v5.Signature)
	if err != nil {
		slog.Error("decoding signature failed.", "error", err)
		return false
	}

	if hmac.Equal(expectedSignature, gotSignature) {
		return true
	}

	slog.Error("verifySignature failed.",
		"expected", hex.EncodeToString(expectedSignature),
		"got", v5.Signature)

	return false
}

func (v5 *Request) sign(imageUrl string, signedParams map[string]string, command string, signingKey string) []byte {
	mac := hmac.New(sha256.New, []byte(signingKey))
	mac.Write([]byte(command))
	mac.Write([]byte(imageUrl))

	for _, signedParam := range signedParams {
		mac.Write([]byte(signedParam))
	}

	return mac.Sum(nil)[0:31]
}

func (v5 *Request) SignedUrl() string {
	signature := hex.EncodeToString(v5.sign(v5.ImageUrl, v5.SignedParams, v5.RawCommands, v5.Config().SigningKey))

	u := v5.URL
	q := u.Query()
	q.Set("sig", signature)
	u.RawQuery = q.Encode()

	return u.String()
}
