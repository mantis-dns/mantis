package gravity

import "github.com/mantis-dns/mantis/internal/domain"

// DefaultBlocklists returns the default blocklist sources shipped with Mantis.
var DefaultBlocklists = []domain.BlocklistSource{
	{
		ID:      "default-stevenblack",
		Name:    "StevenBlack Unified Hosts",
		URL:     "https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts",
		Enabled: true,
		Format:  domain.FormatHosts,
	},
	{
		ID:      "default-pgl-yoyo",
		Name:    "Peter Lowe's Ad and tracking server list",
		URL:     "https://pgl.yoyo.org/adservers/serverlist.php?hostformat=nohtml&showintro=0&mimetype=plaintext",
		Enabled: true,
		Format:  domain.FormatDomains,
	},
	{
		ID:      "default-easylist",
		Name:    "EasyList",
		URL:     "https://easylist.to/easylist/easylist.txt",
		Enabled: false,
		Format:  domain.FormatAdblock,
	},
	{
		ID:      "default-easyprivacy",
		Name:    "EasyPrivacy",
		URL:     "https://easylist.to/easylist/easyprivacy.txt",
		Enabled: false,
		Format:  domain.FormatAdblock,
	},
	{
		ID:      "default-oisd",
		Name:    "OISD Big",
		URL:     "https://big.oisd.nl/domainswild",
		Enabled: false,
		Format:  domain.FormatDomains,
	},
}
