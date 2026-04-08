package gravity

import "strings"

// ReverseDomain reverses a domain for suffix-based radix tree lookups.
// "ads.example.com" -> "com.example.ads."
func ReverseDomain(d string) string {
	d = strings.TrimSuffix(d, ".")
	d = strings.ToLower(d)

	parts := strings.Split(d, ".")
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}
	return strings.Join(parts, ".") + "."
}
