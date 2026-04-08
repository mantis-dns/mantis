package gravity

import (
	"bufio"
	"io"
	"strings"

	"github.com/mantis-dns/mantis/internal/domain"
)

// ParseResult holds the result of parsing a blocklist.
type ParseResult struct {
	Domains []string
	Errors  int
}

// Parse reads a blocklist in the given format and returns extracted domains.
func Parse(r io.Reader, format domain.BlocklistFormat) ParseResult {
	switch format {
	case domain.FormatHosts:
		return parseHosts(r)
	case domain.FormatDomains:
		return parseDomains(r)
	case domain.FormatAdblock:
		return parseAdblock(r)
	default:
		return parseDomains(r)
	}
}

func parseHosts(r io.Reader) ParseResult {
	var result ParseResult
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || line[0] == '#' {
			continue
		}

		// Remove inline comments.
		if idx := strings.IndexByte(line, '#'); idx >= 0 {
			line = strings.TrimSpace(line[:idx])
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			result.Errors++
			continue
		}

		ip := fields[0]
		if ip != "0.0.0.0" && ip != "127.0.0.1" {
			result.Errors++
			continue
		}

		d := strings.ToLower(fields[1])
		if d == "localhost" || d == "localhost.localdomain" || d == "broadcasthost" || d == "local" {
			continue
		}
		if !isValidDomain(d) {
			result.Errors++
			continue
		}
		result.Domains = append(result.Domains, d)
	}
	return result
}

func parseDomains(r io.Reader) ParseResult {
	var result ParseResult
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || line[0] == '#' {
			continue
		}
		if idx := strings.IndexByte(line, '#'); idx >= 0 {
			line = strings.TrimSpace(line[:idx])
		}

		d := strings.ToLower(line)
		if !isValidDomain(d) {
			result.Errors++
			continue
		}
		result.Domains = append(result.Domains, d)
	}
	return result
}

func parseAdblock(r io.Reader) ParseResult {
	var result ParseResult
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || line[0] == '!' || line[0] == '[' {
			continue
		}

		// Match ||domain^ pattern.
		if !strings.HasPrefix(line, "||") {
			continue
		}
		line = line[2:]

		caretIdx := strings.IndexByte(line, '^')
		if caretIdx < 0 {
			continue
		}
		d := strings.ToLower(line[:caretIdx])

		// Skip entries with path separators (URL rules, not domain rules).
		if strings.ContainsAny(d, "/*") {
			continue
		}

		if !isValidDomain(d) {
			result.Errors++
			continue
		}
		result.Domains = append(result.Domains, d)
	}
	return result
}

func isValidDomain(d string) bool {
	if len(d) == 0 || len(d) > 253 {
		return false
	}
	d = strings.TrimSuffix(d, ".")
	for _, label := range strings.Split(d, ".") {
		if len(label) == 0 || len(label) > 63 {
			return false
		}
	}
	return strings.Contains(d, ".")
}
