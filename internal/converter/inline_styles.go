package converter

import (
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/atom"
)

// boldWeight is the smallest numeric font-weight treated as bold (CSS semibold).
const boldWeight = 600

// normalizeInlineStyles rewrites CSS-only formatting into semantic tags before
// conversion. Google Docs is the main source: it marks bold, italic and
// strikethrough with <span style="..."> only, and wraps the whole selection in
// <b style="font-weight:normal">, which would otherwise turn every line bold.
func normalizeInlineStyles(selec *goquery.Selection) {
	selec.Find("span[style]").Each(func(_ int, s *goquery.Selection) {
		// Heading levels already carry the emphasis; do not add ** to titles.
		if s.Closest("h1, h2, h3, h4, h5, h6").Length() > 0 {
			return
		}
		tags := semanticTags(s.AttrOr("style", ""))
		if len(tags) == 0 {
			return
		}
		// Rename the span itself rather than wrapping its content: the Markdown
		// rules decide on surrounding spaces from the element's siblings, which
		// an inner wrapper would not have.
		n := s.Nodes[0]
		n.Data, n.DataAtom = tags[0].String(), tags[0]
		for _, tag := range tags[1:] {
			s.WrapInnerHtml("<" + tag.String() + "></" + tag.String() + ">")
		}
	})

	selec.Find("b[style], strong[style]").Each(func(_ int, s *goquery.Selection) {
		if weight := cssProperty(s.AttrOr("style", ""), "font-weight"); weight != "" && !isBoldWeight(weight) {
			s.ReplaceWithSelection(s.Contents())
		}
	})
}

// semanticTags returns the tags that express the formatting of an inline style.
func semanticTags(style string) []atom.Atom {
	var tags []atom.Atom
	if isBoldWeight(cssProperty(style, "font-weight")) {
		tags = append(tags, atom.Strong)
	}
	if fontStyle := cssProperty(style, "font-style"); fontStyle == "italic" || fontStyle == "oblique" {
		tags = append(tags, atom.Em)
	}
	if strings.Contains(cssProperty(style, "text-decoration"), "line-through") ||
		strings.Contains(cssProperty(style, "text-decoration-line"), "line-through") {
		tags = append(tags, atom.Del)
	}
	return tags
}

// isBoldWeight reports whether a CSS font-weight value renders as bold.
func isBoldWeight(weight string) bool {
	switch weight {
	case "bold", "bolder":
		return true
	case "", "normal", "lighter":
		return false
	}
	n, err := strconv.Atoi(weight)
	return err == nil && n >= boldWeight
}

// cssProperty returns the lowercased value of prop in an inline style
// attribute, or "" when it is absent.
func cssProperty(style, prop string) string {
	for _, decl := range strings.Split(style, ";") {
		name, value, ok := strings.Cut(decl, ":")
		if ok && strings.EqualFold(strings.TrimSpace(name), prop) {
			return strings.ToLower(strings.TrimSpace(value))
		}
	}
	return ""
}
