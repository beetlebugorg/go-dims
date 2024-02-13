package dims

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"io"
	"strings"
)

/* Sign an image URL
 *
 * The signature is the first 7 characters of an md5 hash of:
 *
 *   1. timestamp
 *   2. secret key (from the configuration)
 *   3. commands (as-is from the URL)
 *   4. image URL
 *
 * Example:
 *
 *   1. timestamp: 1234567890
 *   2. secret: mysecret
 *   3. commands: resize/100x100/crop/100x100
 *   4. image URL: http://example.com/image.jpg
 *
 *   md5(1234567890mysecretresize/100x100crop/100x100http://example.com/image.jpg)
 */
func Sign(timestamp string, secret string, commands string, imageUrl string) string {
	sanitizedCommands := strings.ReplaceAll(commands, " ", "+")

	hash := md5.New()
	io.WriteString(hash, timestamp)
	io.WriteString(hash, secret)
	io.WriteString(hash, sanitizedCommands)
	io.WriteString(hash, imageUrl)

	return fmt.Sprintf("%x", hash.Sum(nil))[0:7]
}

func SignHmacSha256(timestamp string, secret string, commands string, imageUrl string) string {
	sanitizedCommands := strings.ReplaceAll(commands, " ", "+")

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write([]byte(sanitizedCommands))
	mac.Write([]byte(imageUrl))

	return fmt.Sprintf("%x", mac.Sum(nil))[0:24]
}
