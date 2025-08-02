package util

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// Ip2long converts an IPv4 address to a long integer
func Ip2long(ipAddress string) int64 {
	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return -1
	}

	ip = ip.To4()
	if ip == nil {
		return -1
	}

	return int64(ip[0])<<24 + int64(ip[1])<<16 + int64(ip[2])<<8 + int64(ip[3])
}

// Long2ip converts a long integer to an IPv4 address
func Long2ip(properAddress int64) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		(properAddress>>24)&0xFF,
		(properAddress>>16)&0xFF,
		(properAddress>>8)&0xFF,
		properAddress&0xFF)
}

// Gethostbyname gets the IPv4 address of a hostname
func Gethostbyname(hostname string) string {
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return hostname
	}

	for _, ip := range ips {
		if ipv4 := ip.To4(); ipv4 != nil {
			return ipv4.String()
		}
	}

	return hostname
}

// Gethostbyaddr gets the hostname of an IP address
func Gethostbyaddr(ipAddress string) string {
	names, err := net.LookupAddr(ipAddress)
	if err != nil || len(names) == 0 {
		return ipAddress
	}

	// Remove trailing dot if present
	hostname := names[0]
	if strings.HasSuffix(hostname, ".") {
		hostname = hostname[:len(hostname)-1]
	}

	return hostname
}

// Gethostbynamel gets a list of IPv4 addresses of a hostname
func Gethostbynamel(hostname string) []string {
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return nil
	}

	var ipv4s []string
	for _, ip := range ips {
		if ipv4 := ip.To4(); ipv4 != nil {
			ipv4s = append(ipv4s, ipv4.String())
		}
	}

	return ipv4s
}

// Checkdnsrr checks DNS records
func Checkdnsrr(hostname string, recordType ...string) bool {
	rType := "MX"
	if len(recordType) > 0 {
		rType = strings.ToUpper(recordType[0])
	}

	switch rType {
	case "A":
		_, err := net.LookupIP(hostname)
		return err == nil
	case "AAAA":
		ips, err := net.LookupIP(hostname)
		if err != nil {
			return false
		}
		for _, ip := range ips {
			if ip.To16() != nil && ip.To4() == nil {
				return true
			}
		}
		return false
	case "MX":
		_, err := net.LookupMX(hostname)
		return err == nil
	case "NS":
		_, err := net.LookupNS(hostname)
		return err == nil
	case "TXT":
		_, err := net.LookupTXT(hostname)
		return err == nil
	case "CNAME":
		_, err := net.LookupCNAME(hostname)
		return err == nil
	default:
		return false
	}
}

// Getmxrr gets MX records for a hostname
func Getmxrr(hostname string) ([]string, []int, bool) {
	mxRecords, err := net.LookupMX(hostname)
	if err != nil {
		return nil, nil, false
	}

	var hosts []string
	var priorities []int

	for _, mx := range mxRecords {
		host := mx.Host
		if strings.HasSuffix(host, ".") {
			host = host[:len(host)-1]
		}
		hosts = append(hosts, host)
		priorities = append(priorities, int(mx.Pref))
	}

	return hosts, priorities, true
}

// Dns_get_record gets DNS records
func DnsGetRecord(hostname string, recordType ...int) []map[string]any {
	const (
		DNS_A     = 1
		DNS_NS    = 2
		DNS_CNAME = 16
		DNS_MX    = 15
		DNS_TXT   = 16384
		DNS_AAAA  = 28
		DNS_ALL   = DNS_A | DNS_NS | DNS_CNAME | DNS_MX | DNS_TXT | DNS_AAAA
	)

	rType := DNS_ALL
	if len(recordType) > 0 {
		rType = recordType[0]
	}

	var records []map[string]any

	// A records
	if rType&DNS_A != 0 {
		ips, err := net.LookupIP(hostname)
		if err == nil {
			for _, ip := range ips {
				if ipv4 := ip.To4(); ipv4 != nil {
					record := map[string]any{
						"host":  hostname,
						"class": "IN",
						"ttl":   86400,
						"type":  "A",
						"ip":    ipv4.String(),
					}
					records = append(records, record)
				}
			}
		}
	}

	// AAAA records
	if rType&DNS_AAAA != 0 {
		ips, err := net.LookupIP(hostname)
		if err == nil {
			for _, ip := range ips {
				if ip.To16() != nil && ip.To4() == nil {
					record := map[string]any{
						"host":  hostname,
						"class": "IN",
						"ttl":   86400,
						"type":  "AAAA",
						"ipv6":  ip.String(),
					}
					records = append(records, record)
				}
			}
		}
	}

	// MX records
	if rType&DNS_MX != 0 {
		mxRecords, err := net.LookupMX(hostname)
		if err == nil {
			for _, mx := range mxRecords {
				target := mx.Host
				if strings.HasSuffix(target, ".") {
					target = target[:len(target)-1]
				}
				record := map[string]any{
					"host":   hostname,
					"class":  "IN",
					"ttl":    86400,
					"type":   "MX",
					"pri":    int(mx.Pref),
					"target": target,
				}
				records = append(records, record)
			}
		}
	}

	// NS records
	if rType&DNS_NS != 0 {
		nsRecords, err := net.LookupNS(hostname)
		if err == nil {
			for _, ns := range nsRecords {
				target := ns.Host
				if strings.HasSuffix(target, ".") {
					target = target[:len(target)-1]
				}
				record := map[string]any{
					"host":   hostname,
					"class":  "IN",
					"ttl":    86400,
					"type":   "NS",
					"target": target,
				}
				records = append(records, record)
			}
		}
	}

	// TXT records
	if rType&DNS_TXT != 0 {
		txtRecords, err := net.LookupTXT(hostname)
		if err == nil {
			for _, txt := range txtRecords {
				record := map[string]any{
					"host":  hostname,
					"class": "IN",
					"ttl":   86400,
					"type":  "TXT",
					"txt":   txt,
				}
				records = append(records, record)
			}
		}
	}

	// CNAME records
	if rType&DNS_CNAME != 0 {
		cname, err := net.LookupCNAME(hostname)
		if err == nil && cname != hostname+"." {
			if strings.HasSuffix(cname, ".") {
				cname = cname[:len(cname)-1]
			}
			record := map[string]any{
				"host":   hostname,
				"class":  "IN",
				"ttl":    86400,
				"type":   "CNAME",
				"target": cname,
			}
			records = append(records, record)
		}
	}

	return records
}

// Helper functions for validation

func isValidEmail(email string) bool {
	// Simple email validation
	if !strings.Contains(email, "@") {
		return false
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	local, domain := parts[0], parts[1]
	if len(local) == 0 || len(domain) == 0 {
		return false
	}

	if !strings.Contains(domain, ".") {
		return false
	}

	return true
}

func isValidURL(url string) bool {
	// Simple URL validation
	if !strings.HasPrefix(url, "http://") &&
		!strings.HasPrefix(url, "https://") &&
		!strings.HasPrefix(url, "ftp://") {
		return false
	}

	return strings.Contains(url, ".")
}

// ParseUrl parses a URL and returns its components
func ParseUrl(url string, component ...int) any {
	const (
		URL_SCHEME   = 0
		URL_HOST     = 1
		URL_PORT     = 2
		URL_USER     = 3
		URL_PASS     = 4
		URL_PATH     = 5
		URL_QUERY    = 6
		URL_FRAGMENT = 7
	)

	// Simple URL parsing
	scheme := ""
	host := ""
	port := 0
	user := ""
	pass := ""
	path := ""
	query := ""
	fragment := ""

	// Extract scheme
	if idx := strings.Index(url, "://"); idx != -1 {
		scheme = url[:idx]
		url = url[idx+3:]
	}

	// Extract fragment
	if idx := strings.Index(url, "#"); idx != -1 {
		fragment = url[idx+1:]
		url = url[:idx]
	}

	// Extract query
	if idx := strings.Index(url, "?"); idx != -1 {
		query = url[idx+1:]
		url = url[:idx]
	}

	// Extract user:pass@host:port/path
	if idx := strings.Index(url, "/"); idx != -1 {
		path = url[idx:]
		url = url[:idx]
	} else {
		path = "/"
	}

	// Extract user:pass@
	if idx := strings.Index(url, "@"); idx != -1 {
		userPass := url[:idx]
		url = url[idx+1:]

		if passIdx := strings.Index(userPass, ":"); passIdx != -1 {
			user = userPass[:passIdx]
			pass = userPass[passIdx+1:]
		} else {
			user = userPass
		}
	}

	// Extract host:port
	if idx := strings.Index(url, ":"); idx != -1 {
		host = url[:idx]
		if p, err := strconv.Atoi(url[idx+1:]); err == nil {
			port = p
		}
	} else {
		host = url
	}

	result := map[string]any{
		"scheme":   scheme,
		"host":     host,
		"port":     port,
		"user":     user,
		"pass":     pass,
		"path":     path,
		"query":    query,
		"fragment": fragment,
	}

	if len(component) > 0 {
		switch component[0] {
		case URL_SCHEME:
			if scheme == "" {
				return nil
			}
			return scheme
		case URL_HOST:
			if host == "" {
				return nil
			}
			return host
		case URL_PORT:
			if port == 0 {
				return nil
			}
			return port
		case URL_USER:
			if user == "" {
				return nil
			}
			return user
		case URL_PASS:
			if pass == "" {
				return nil
			}
			return pass
		case URL_PATH:
			if path == "" {
				return nil
			}
			return path
		case URL_QUERY:
			if query == "" {
				return nil
			}
			return query
		case URL_FRAGMENT:
			if fragment == "" {
				return nil
			}
			return fragment
		}
	}

	return result
}

// HttpBuildQuery builds URL-encoded query string
func HttpBuildQuery(data map[string]any, prefix ...string) string {
	var parts []string
	pfx := ""
	if len(prefix) > 0 {
		pfx = prefix[0]
	}

	for key, value := range data {
		fullKey := key
		if pfx != "" {
			fullKey = pfx + "[" + key + "]"
		}

		switch v := value.(type) {
		case map[string]any:
			subQuery := HttpBuildQuery(v, fullKey)
			if subQuery != "" {
				parts = append(parts, subQuery)
			}
		case []any:
			for i, item := range v {
				itemKey := fullKey + "[" + strconv.Itoa(i) + "]"
				parts = append(parts, itemKey+"="+UrlEncode(fmt.Sprintf("%v", item)))
			}
		default:
			parts = append(parts, fullKey+"="+UrlEncode(fmt.Sprintf("%v", value)))
		}
	}

	return strings.Join(parts, "&")
}

// ParseStr parses a query string into variables
func ParseStr(str string) map[string]any {
	result := make(map[string]any)

	if str == "" {
		return result
	}

	pairs := strings.Split(str, "&")
	for _, pair := range pairs {
		if idx := strings.Index(pair, "="); idx != -1 {
			key := pair[:idx]
			value := pair[idx+1:]

			// URL decode
			if decodedKey, err := UrlDecode(key); err == nil {
				key = decodedKey
			}
			if decodedValue, err := UrlDecode(value); err == nil {
				value = decodedValue
			}

			result[key] = value
		} else {
			if decodedKey, err := UrlDecode(pair); err == nil {
				result[decodedKey] = ""
			} else {
				result[pair] = ""
			}
		}
	}

	return result
}

// GetHeaders gets all HTTP headers (simplified implementation)
func GetHeaders() map[string]string {
	// In a real implementation, this would read from the current HTTP request
	// This is a placeholder that returns common headers
	return map[string]string{
		"Host":       "localhost",
		"User-Agent": "Go HTTP Client",
		"Accept":     "*/*",
	}
}

// GetallHeaders alias for GetHeaders
func GetallHeaders() map[string]string {
	return GetHeaders()
}

// Header sends a raw HTTP header (placeholder - would need HTTP context)
func Header(header string, replace ...bool) {
	// In a real implementation, this would set HTTP response headers
	// This is a placeholder function
	fmt.Printf("Header: %s\n", header)
}

// HttpResponseCode gets or sets the HTTP response status code (placeholder)
func HttpResponseCode(responseCode ...int) int {
	// In a real implementation, this would work with HTTP response
	if len(responseCode) > 0 {
		fmt.Printf("Setting response code: %d\n", responseCode[0])
		return responseCode[0]
	}
	return 200 // Default OK
}

// HeadersSent checks if headers have been sent (placeholder)
func HeadersSent() bool {
	// In a real implementation, this would check if HTTP headers were sent
	return false
}

// HeadersList returns a list of response headers sent (placeholder)
func HeadersList() []string {
	// In a real implementation, this would return actual sent headers
	return []string{"Content-Type: text/html"}
}
