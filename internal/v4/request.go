package v4

import (
	"crypto/md5"
	"crypto/subtle"
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
	request.EtagAlgorithm = "md5"

	return &Request{
		Request: request,

		clientId:  clientId,
		timestamp: timestamp,
	}, nil
}

func (v4 *Request) Validate() bool {
	expectedSignature := v4.sign(v4.RawCommands, v4.timestamp, v4.ImageUrl, v4.SignedQuery, v4.Config().SigningKey)

	if subtle.ConstantTimeCompare([]byte(expectedSignature), []byte(v4.Signature)) == 1 {
		return true
	}

	// Legacy mode also accepts the mod_dims signature, which covers only the
	// parameters named by _keys. Every other parameter is unprotected in that
	// mode.
	if v4.Config().Compat == "legacy" {
		legacy := v4.signLegacy(v4.RawCommands, v4.timestamp, v4.ImageUrl, v4.LegacyParams, v4.Config().SigningKey)
		if subtle.ConstantTimeCompare([]byte(legacy), []byte(v4.Signature)) == 1 {
			slog.Warn("accepted a legacy signature", "url", v4.ImageUrl)
			return true
		}
	}

	slog.Error("verifySignature failed.",
		"expected", expectedSignature,
		"got", v4.Signature)

	return false
}

func (v4 *Request) sign(commands, timestamp, imageUrl string, signedQuery string, signingKey string) string {
	key := strings.Replace(signingKey, "sha1:", "", 1)

	h := md5.New()
	h.Write([]byte(timestamp))
	h.Write([]byte(key))
	h.Write([]byte(core.SignedMessage(commands, imageUrl, signedQuery)))

	return fmt.Sprintf("%x", h.Sum(nil))[0:7]
}

// signLegacy reproduces the mod_dims construction: the parameters named by
// _keys, in _keys order, concatenated with no separator. It is reachable only
// under DIMS_SIGNING_COMPAT=legacy.
func (v4 *Request) signLegacy(commands, timestamp, imageUrl string, legacyParams []string, signingKey string) string {
	key := strings.Replace(signingKey, "sha1:", "", 1)

	h := md5.New()
	h.Write([]byte(timestamp))
	h.Write([]byte(key))
	h.Write([]byte(commands))
	h.Write([]byte(imageUrl))

	for _, legacyParam := range legacyParams {
		h.Write([]byte(legacyParam))
	}

	return fmt.Sprintf("%x", h.Sum(nil))[0:7]
}

func (v4 *Request) SignedUrl() string {
	signature := v4.sign(v4.RawCommands, v4.timestamp, v4.ImageUrl, v4.SignedQuery, v4.Config().SigningKey)

	// Rebuild the path from its parts. Substituting the placeholder went
	// wrong whenever it happened to equal the client id or the timestamp.
	// The URL is copied so signing does not alter the request it was given.
	signed := *v4.URL
	signed.Path = fmt.Sprintf("/dims4/%s/%s/%s/%s", v4.clientId, signature, v4.timestamp, v4.RawCommands)

	return signed.String()
}
