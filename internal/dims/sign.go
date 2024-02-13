package dims

import (
	"crypto/md5"
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

	md5 := md5.New()
	io.WriteString(md5, timestamp)
	io.WriteString(md5, secret)
	io.WriteString(md5, sanitizedCommands)
	io.WriteString(md5, imageUrl)

	return fmt.Sprintf("%x", md5.Sum(nil))[0:7]
}
