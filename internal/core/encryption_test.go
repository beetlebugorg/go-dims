package core

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const testURL = "https://example.com/some/image.jpg?a=1&b=2"

func TestEncryptDecryptRoundTrip(t *testing.T) {
	for _, key := range []string{"a-long-enough-secret", "sha1:t3st", "hkdf:another-secret"} {
		encrypted, err := EncryptURLKey(key, testURL)
		require.NoError(t, err, "key %q", key)

		decrypted, err := DecryptURLKey(key, encrypted)
		require.NoError(t, err, "key %q", key)
		require.Equal(t, testURL, decrypted, "key %q", key)
	}
}

func TestDecryptRejectsTheWrongKey(t *testing.T) {
	encrypted, err := EncryptURLKey("the-right-secret", testURL)
	require.NoError(t, err)

	_, err = DecryptURLKey("the-wrong-secret", encrypted)
	require.Error(t, err)
}

// A query string turns '+' into a space, so an eurl often arrives altered.
// Both alphabets have to decode.
func TestDecryptAcceptsBothBase64Alphabets(t *testing.T) {
	key, err := deriveKey("a-long-enough-secret")
	require.NoError(t, err)

	standard, err := EncryptAES128GCM(key, testURL)
	require.NoError(t, err)

	raw, err := base64.StdEncoding.DecodeString(standard)
	require.NoError(t, err)

	for name, encoded := range map[string]string{
		"standard":     standard,
		"raw standard": base64.RawStdEncoding.EncodeToString(raw),
		"url safe":     base64.URLEncoding.EncodeToString(raw),
		"raw url safe": base64.RawURLEncoding.EncodeToString(raw),
	} {
		decrypted, err := DecryptAES128GCM(key, encoded)
		require.NoError(t, err, "%s", name)
		require.Equal(t, testURL, decrypted, "%s", name)
	}
}

func TestDecryptRejectsShortInput(t *testing.T) {
	key, err := deriveKey("a-long-enough-secret")
	require.NoError(t, err)

	// Shorter than a nonce plus a tag.
	_, err = DecryptAES128GCM(key, base64.StdEncoding.EncodeToString(make([]byte, 8)))
	require.Error(t, err)
}

func TestDecryptRejectsGarbage(t *testing.T) {
	key, err := deriveKey("a-long-enough-secret")
	require.NoError(t, err)

	_, err = DecryptAES128GCM(key, "not base64 at all !!!")
	require.Error(t, err)
}

// The mod_dims path uses 16 hex characters of a SHA-1 digest, so the key holds
// 64 bits of material. Kept for compatibility, and reported at startup.
func TestLegacyKey(t *testing.T) {
	require.True(t, LegacyKey("sha1:t3st"))
	require.False(t, LegacyKey("hkdf:t3st"))
	require.False(t, LegacyKey("t3st"))

	derived, err := deriveKey("sha1:t3st")
	require.NoError(t, err)
	require.Len(t, derived, 16)
	require.Equal(t, strings.ToUpper(string(derived)), string(derived), "the fragment is uppercase hex")
}

// A query string decodes '+' to a space, so an eurl commonly arrives with
// spaces where the standard alphabet had plus signs.
func TestDecryptURLKeyRepairsSpaces(t *testing.T) {
	const key = "a-long-enough-secret"

	// Find a ciphertext whose encoding contains a plus sign, since the nonce
	// is random and not every one will.
	var encrypted string
	for i := 0; i < 200; i++ {
		candidate, err := EncryptURLKey(key, testURL)
		require.NoError(t, err)

		if strings.Contains(candidate, "+") {
			encrypted = candidate
			break
		}
	}
	require.NotEmpty(t, encrypted, "no ciphertext containing a plus sign was produced")

	mangled := strings.ReplaceAll(encrypted, "+", " ")
	require.NotEqual(t, encrypted, mangled)

	decrypted, err := DecryptURLKey(key, mangled)
	require.NoError(t, err)
	require.Equal(t, testURL, decrypted)
}
