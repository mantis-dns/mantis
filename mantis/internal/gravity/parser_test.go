package gravity

import (
	"strings"
	"testing"

	"github.com/mantis-dns/mantis/internal/domain"
)

func TestParseHosts(t *testing.T) {
	input := `# Comment line
127.0.0.1 localhost
0.0.0.0 ads.example.com
0.0.0.0 tracker.net  # inline comment
127.0.0.1 malware.org
0.0.0.0 localhost.localdomain
invalid line without ip
`
	result := Parse(strings.NewReader(input), domain.FormatHosts)

	if len(result.Domains) != 3 {
		t.Errorf("expected 3 domains, got %d: %v", len(result.Domains), result.Domains)
	}
	expected := map[string]bool{"ads.example.com": true, "tracker.net": true, "malware.org": true}
	for _, d := range result.Domains {
		if !expected[d] {
			t.Errorf("unexpected domain %q", d)
		}
	}
}

func TestParseDomains(t *testing.T) {
	input := `# Blocklist
ads.example.com
tracker.net

# Another comment
malware.org
`
	result := Parse(strings.NewReader(input), domain.FormatDomains)
	if len(result.Domains) != 3 {
		t.Errorf("expected 3 domains, got %d", len(result.Domains))
	}
}

func TestParseAdblock(t *testing.T) {
	input := `[Adblock Plus 2.0]
! Title: Test
||ads.example.com^
||tracker.net^
||malware.org^$third-party
||example.com/path^
! Comment
||valid.blocker.io^
`
	result := Parse(strings.NewReader(input), domain.FormatAdblock)

	// "ads.example.com", "tracker.net", "valid.blocker.io" should match.
	// "malware.org^$third-party" - caret is at right place, before $, so malware.org extracted.
	// "example.com/path" has slash, skipped.
	expected := map[string]bool{
		"ads.example.com":  true,
		"tracker.net":      true,
		"valid.blocker.io": true,
	}

	// malware.org^$third-party: ^ is at index 11, before $. So "malware.org" is extracted.
	// Let me check: line = "||malware.org^$third-party", after removing "||" = "malware.org^$third-party"
	// caretIdx = index of '^' = 11, d = "malware.org"
	expected["malware.org"] = true

	if len(result.Domains) != 4 {
		t.Errorf("expected 4 domains, got %d: %v", len(result.Domains), result.Domains)
	}
	for _, d := range result.Domains {
		if !expected[d] {
			t.Errorf("unexpected domain %q", d)
		}
	}
}

func TestParseEmptyAndComments(t *testing.T) {
	input := `
# Only comments
# and empty lines

`
	result := Parse(strings.NewReader(input), domain.FormatDomains)
	if len(result.Domains) != 0 {
		t.Errorf("expected 0 domains, got %d", len(result.Domains))
	}
}

func TestIsValidDomain(t *testing.T) {
	valid := []string{"example.com", "sub.domain.org", "a.b"}
	invalid := []string{"", "nodot", strings.Repeat("a", 254), "a..b"}

	for _, d := range valid {
		if !isValidDomain(d) {
			t.Errorf("%q should be valid", d)
		}
	}
	for _, d := range invalid {
		if isValidDomain(d) {
			t.Errorf("%q should be invalid", d)
		}
	}
}
