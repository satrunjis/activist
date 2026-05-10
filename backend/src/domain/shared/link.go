package shared

// Link stores an external URL with a platform tag (e.g. "vk", "tg").
type Link struct {
	Platform string
	Value    string
}

func (l Link) Equal(other Link) bool {
	return l.Platform == other.Platform && l.Value == other.Value
}

// NormalizeLinks trims whitespace from each link field in place.
func NormalizeLinks(links []Link) {
	for i := range links {
		links[i].Platform = Trim(links[i].Platform)
		links[i].Value = Trim(links[i].Value)
	}
}

// ValidateLinks checks that each link has non-blank platform and value.
// Call after NormalizeLinks.
func ValidateLinks(links []Link) error {
	for _, l := range links {
		if err := RequireText("link.platform", l.Platform, MaxShortText); err != nil {
			return err
		}
		if err := RequireText("link.value", l.Value, MaxURL); err != nil {
			return err
		}
	}
	return nil
}
