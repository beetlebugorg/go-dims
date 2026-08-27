package core

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// reservedBlocks are ranges that are never a legitimate image origin and that
// the net package does not already classify. Loopback, link local, private,
// multicast, and unspecified addresses are handled by the net.IP methods.
var reservedBlocks = []string{
	"100.64.0.0/10",   // carrier grade NAT
	"192.0.0.0/24",    // IETF protocol assignments
	"192.0.2.0/24",    // documentation
	"198.18.0.0/15",   // benchmarking
	"198.51.100.0/24", // documentation
	"203.0.113.0/24",  // documentation
	"240.0.0.0/4",     // reserved
	"::/128",          // unspecified
	"64:ff9b::/96",    // IPv4/IPv6 translation
	"2001:db8::/32",   // documentation
}

var reservedNets = func() []*net.IPNet {
	nets := make([]*net.IPNet, 0, len(reservedBlocks))
	for _, block := range reservedBlocks {
		if _, network, err := net.ParseCIDR(block); err == nil {
			nets = append(nets, network)
		}
	}
	return nets
}()

// IsPublicAddress reports whether an address is routable on the public
// internet. An IPv4 address mapped into IPv6 is judged as IPv4, so
// ::ffff:169.254.169.254 is refused along with 169.254.169.254.
func IsPublicAddress(ip net.IP) bool {
	if ip == nil {
		return false
	}

	if mapped := ip.To4(); mapped != nil {
		ip = mapped
	}

	if ip.IsLoopback() || ip.IsUnspecified() || ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsInterfaceLocalMulticast() || ip.IsMulticast() {
		return false
	}

	for _, network := range reservedNets {
		if network.Contains(ip) {
			return false
		}
	}

	return true
}

// HostAllowed reports whether host passes the configured allowlist. An empty
// allowlist permits every host.
func HostAllowed(host string, allowed []string) bool {
	if len(allowed) == 0 {
		return true
	}

	host = strings.ToLower(strings.TrimSuffix(host, "."))
	for _, entry := range allowed {
		entry = strings.ToLower(strings.TrimSpace(entry))
		if entry == "" {
			continue
		}

		if entry == host {
			return true
		}

		// A leading dot matches the domain and any subdomain of it.
		if strings.HasPrefix(entry, ".") &&
			(strings.HasSuffix(host, entry) || host == entry[1:]) {
			return true
		}
	}

	return false
}

// ValidateImageURL checks the scheme and host of a URL before any connection
// is made. The address itself is checked separately, once it resolves.
func ValidateImageURL(u *url.URL, config Network) error {
	if u.Scheme != "http" && u.Scheme != "https" {
		return NewStatusError(400, fmt.Sprintf("unsupported scheme %q", u.Scheme))
	}

	if !HostAllowed(u.Hostname(), config.AllowedHosts) {
		return NewStatusError(400, fmt.Sprintf("host %q is not allowed", u.Hostname()))
	}

	return nil
}
