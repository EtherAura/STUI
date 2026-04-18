// dns.go implements DNS resolution and port reachability checks for
// domain-based installers. Before running upstream install scripts that
// depend on Let's Encrypt (certbot) for TLS certificates, STUI verifies
// that the configured domain resolves to a public IP address and that
// the required ports (80/443) are reachable. This avoids wasting time
// on install steps that will inevitably fail when certbot cannot
// complete its HTTP-01 challenge.
package installer

import (
	"fmt"
	"net"
	"strings"
	"time"
)

// DNSResult holds the outcome of a domain resolution check.
type DNSResult struct {
	// Domain is the FQDN that was looked up.
	Domain string
	// IPs is the list of resolved IP addresses.
	IPs []net.IP
	// Ports holds per-port reachability results, or nil if port
	// checks were not performed.
	Ports []PortResult
	// Errors lists blocking issues that prevent installation.
	Errors []string
	// Warnings lists non-blocking issues the user should know about.
	Warnings []string
}

// PortResult records whether a single port is reachable.
type PortResult struct {
	// Port is the TCP port number that was tested.
	Port int
	// Open is true if a TCP connection succeeded.
	Open bool
	// Err is non-nil when the connection attempt failed.
	Err error
}

// OK returns true when the domain resolved and no blocking errors
// were found.
func (r *DNSResult) OK() bool {
	return len(r.Errors) == 0
}

// ResolveDomain performs a DNS lookup on the given domain and checks
// the results for common problems that would cause certbot to fail:
//   - domain does not resolve at all
//   - domain resolves only to private/loopback addresses
func ResolveDomain(domain string) *DNSResult {
	result := &DNSResult{Domain: domain}

	addrs, err := net.LookupHost(domain)
	if err != nil {
		result.Errors = append(result.Errors,
			fmt.Sprintf("DNS lookup failed for %q: %v", domain, err))
		return result
	}

	if len(addrs) == 0 {
		result.Errors = append(result.Errors,
			fmt.Sprintf("DNS lookup for %q returned no addresses", domain))
		return result
	}

	// Parse each returned address.
	var hasPublic bool
	for _, addr := range addrs {
		ip := net.ParseIP(addr)
		if ip == nil {
			continue
		}
		result.IPs = append(result.IPs, ip)

		if isPrivateIP(ip) {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("%s resolves to private address %s; Let's Encrypt cannot reach private IPs", domain, ip))
		} else {
			hasPublic = true
		}
	}

	if !hasPublic && len(result.IPs) > 0 {
		result.Errors = append(result.Errors,
			fmt.Sprintf("%s resolves only to private/loopback addresses (%s); "+
				"Let's Encrypt requires a publicly reachable IP for HTTP-01 challenges",
				domain, FormatIPs(result.IPs)))
	}

	return result
}

// portCheckTimeout is the TCP dial timeout for port reachability checks.
const portCheckTimeout = 5 * time.Second

// CheckPorts probes the given TCP ports on each public IP in the
// DNSResult. Results and any errors/warnings are appended directly
// to the DNSResult. Ports that cannot be reached are recorded as
// warnings (non-blocking) because transient network issues or
// firewalls may cause false negatives.
func (r *DNSResult) CheckPorts(ports ...int) {
	for _, ip := range r.IPs {
		if isPrivateIP(ip) {
			continue
		}
		for _, port := range ports {
			addr := net.JoinHostPort(ip.String(), fmt.Sprintf("%d", port))
			conn, err := net.DialTimeout("tcp", addr, portCheckTimeout)
			pr := PortResult{Port: port, Open: err == nil, Err: err}
			r.Ports = append(r.Ports, pr)
			if conn != nil {
				_ = conn.Close()
			}
			if err != nil {
				r.Warnings = append(r.Warnings,
					fmt.Sprintf("port %d on %s is not reachable: %v — check firewall/port forwarding", port, ip, err))
			}
		}
	}
}

// isPrivateIP returns true if ip falls within RFC 1918, RFC 4193,
// or loopback ranges.
func isPrivateIP(ip net.IP) bool {
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast()
}

// FormatIPs returns a comma-separated string of IP addresses.
func FormatIPs(ips []net.IP) string {
	parts := make([]string, len(ips))
	for i, ip := range ips {
		parts[i] = ip.String()
	}
	return strings.Join(parts, ", ")
}
