package core

import (
	"net"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsPublicAddress(t *testing.T) {
	refused := []string{
		"169.254.169.254",        // EC2 and ECS instance metadata
		"::ffff:169.254.169.254", // the same address mapped into IPv6
		"127.0.0.1",
		"::1",
		"10.1.2.3",
		"172.16.0.1",
		"192.168.1.1",
		"0.0.0.0",
		"100.64.0.1", // carrier grade NAT
		"198.18.0.1", // benchmarking
		"224.0.0.1",  // multicast
		"fd00::1",    // IPv6 unique local
		"fe80::1",    // IPv6 link local
		"240.0.0.1",  // reserved
	}
	for _, address := range refused {
		require.False(t, IsPublicAddress(net.ParseIP(address)), "%s must be refused", address)
	}

	allowed := []string{"8.8.8.8", "1.1.1.1", "93.184.216.34", "2606:2800:220:1:248:1893:25c8:1946"}
	for _, address := range allowed {
		require.True(t, IsPublicAddress(net.ParseIP(address)), "%s must be allowed", address)
	}

	require.False(t, IsPublicAddress(nil))
}

func TestHostAllowed(t *testing.T) {
	require.True(t, HostAllowed("anything.example", nil), "an empty list allows any host")

	list := []string{"images.example.com", ".cdn.example.net"}

	require.True(t, HostAllowed("images.example.com", list))
	require.True(t, HostAllowed("IMAGES.EXAMPLE.COM", list), "matching ignores case")
	require.True(t, HostAllowed("images.example.com.", list), "a trailing dot still matches")
	require.True(t, HostAllowed("a.cdn.example.net", list), "a leading dot matches a subdomain")
	require.True(t, HostAllowed("cdn.example.net", list), "a leading dot matches the domain itself")

	require.False(t, HostAllowed("evil.example.com", list))
	require.False(t, HostAllowed("images.example.com.evil.test", list), "a suffix must not match a bare entry")
	require.False(t, HostAllowed("notcdn.example.net", list))
}

func TestValidateImageURL(t *testing.T) {
	open := Network{}

	for _, raw := range []string{"file:///etc/passwd", "gopher://x/1", "ftp://example.com/a.jpg"} {
		u, err := url.Parse(raw)
		require.NoError(t, err)
		require.Error(t, ValidateImageURL(u, open), "%s must be refused", raw)
	}

	u, err := url.Parse("https://example.com/a.jpg")
	require.NoError(t, err)
	require.NoError(t, ValidateImageURL(u, open))

	restricted := Network{AllowedHosts: []string{"images.example.com"}}
	require.Error(t, ValidateImageURL(u, restricted))
}
