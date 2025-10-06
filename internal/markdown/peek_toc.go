package markdown

import "strings"

// GenerateSelectors returns selectors for headings.
func GenerateSelectors(filename string, headings []HeadingInfo, short bool) []string {
	selectors := make([]string, len(headings))
	for i, h := range headings {
		if short {
			selectors[i] = GenerateShortSelector(filename, h, headings)
		} else {
			selectors[i] = filename + "#" + strings.ToLower(strings.Join(h.Path, "/"))
		}
	}
	return selectors
}

// GenerateShortSelector returns the shortest unique selector for a heading.
func GenerateShortSelector(filename string, target HeadingInfo, all []HeadingInfo) string {
	for i := len(target.Path) - 1; i >= 0; i-- {
		seg := target.Path[i]
		segLower := strings.ToLower(seg)
		count := 0
		for _, h := range all {
			if strings.Contains(strings.ToLower(h.Text), segLower) {
				count++
			}
		}
		if count == 1 {
			return filename + "#" + segLower
		}
	}
	return filename + "#" + strings.ToLower(strings.Join(target.Path, "/"))
}
