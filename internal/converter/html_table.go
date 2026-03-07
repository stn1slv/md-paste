package converter

import (
	"strconv"
	"strings"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/stn1slv/md-paste/internal/models"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// ExtractTableFromHTML parses an HTML string and extracts the first <table> it finds into a Table model.
func ExtractTableFromHTML(rawHTML string) (models.Table, bool) {
	doc, err := html.Parse(strings.NewReader(rawHTML))
	if err != nil {
		return models.Table{}, false
	}

	tableNode := findFirstTable(doc)
	if tableNode == nil {
		return models.Table{}, false
	}

	return parseTable(tableNode), true
}

func findFirstTable(n *html.Node) *html.Node {
	if n.Type == html.ElementNode && n.DataAtom == atom.Table {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if res := findFirstTable(c); res != nil {
			return res
		}
	}
	return nil
}

func parseTable(tableNode *html.Node) models.Table {
	var table models.Table
	rows := findNodes(tableNode, atom.Tr)

	converter := htmltomarkdown.NewConverter("", true, nil)

	for _, tr := range rows {
		row := parseRow(tr, &table, converter)
		if len(row.Cells) > 0 {
			table.Rows = append(table.Rows, row)
		}
	}

	return FlattenTable(table)
}

func parseRow(tr *html.Node, table *models.Table, converter *htmltomarkdown.Converter) models.Row {
	var row models.Row
	cells := findNodes(tr, atom.Th, atom.Td)

	if len(cells) == 0 {
		return row
	}

	isHeaderRow := false
	for _, cellNode := range cells {
		if cellNode.DataAtom == atom.Th {
			isHeaderRow = true
			break
		}
	}

	// FR-010: Only the first encountered row of table headers is the GFM header row.
	if isHeaderRow && !table.HasHeader {
		row.IsHeader = true
		table.HasHeader = true
	}

	for _, cellNode := range cells {
		rs, cs := getSpans(cellNode)
		cell := models.Cell{
			Alignment: getAlignment(cellNode),
			Content:   processCellContent(cellNode, converter),
			RowSpan:   rs,
			ColSpan:   cs,
		}
		row.Cells = append(row.Cells, cell)
	}
	return row
}

func getSpans(n *html.Node) (rowSpan int, colSpan int) {
	rowSpan, colSpan = 1, 1
	for _, attr := range n.Attr {
		if attr.Key == "rowspan" {
			if val, err := strconv.Atoi(attr.Val); err == nil {
				rowSpan = val
			}
		}
		if attr.Key == "colspan" {
			if val, err := strconv.Atoi(attr.Val); err == nil {
				colSpan = val
			}
		}
	}
	return rowSpan, colSpan
}

func findNodes(n *html.Node, atoms ...atom.Atom) []*html.Node {
	var results []*html.Node
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		match := false
		for _, a := range atoms {
			if c.DataAtom == a {
				match = true
				break
			}
		}
		if match {
			results = append(results, c)
		} else {
			results = append(results, findNodes(c, atoms...)...)
		}
	}
	return results
}

func getAlignment(n *html.Node) models.Alignment {
	for _, attr := range n.Attr {
		if attr.Key == "align" {
			if align := parseAlignVal(attr.Val); align != models.AlignNone {
				return align
			}
		}
		if attr.Key == "style" {
			if align := parseStyleAttr(attr.Val); align != models.AlignNone {
				return align
			}
		}
	}
	return models.AlignNone
}

func parseAlignVal(val string) models.Alignment {
	switch strings.ToLower(val) {
	case "left":
		return models.AlignLeft
	case "center":
		return models.AlignCenter
	case "right":
		return models.AlignRight
	default:
		return models.AlignNone
	}
}

func parseStyleAttr(val string) models.Alignment {
	styles := strings.Split(val, ";")
	for _, style := range styles {
		parts := strings.Split(style, ":")
		if len(parts) == 2 && strings.TrimSpace(parts[0]) == "text-align" {
			return parseAlignVal(strings.TrimSpace(parts[1]))
		}
	}
	return models.AlignNone
}

func processCellContent(n *html.Node, converter *htmltomarkdown.Converter) string {
	var innerHTML strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		_ = html.Render(&innerHTML, c) // Ignoring error as we have fallback
	}

	markdown, err := converter.ConvertString(innerHTML.String())
	if err != nil {
		var sb strings.Builder
		renderTextOnly(n, &sb)
		return strings.TrimSpace(sb.String())
	}

	return strings.TrimSpace(markdown)
}

func renderTextOnly(n *html.Node, sb *strings.Builder) {
	if n.Type == html.TextNode {
		sb.WriteString(n.Data)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		renderTextOnly(c, sb)
	}
}
