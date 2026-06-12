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

	return ParseTable(tableNode), true
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

func findTableRows(tableNode *html.Node) []*html.Node {
	var rows []*html.Node

	for child := tableNode.FirstChild; child != nil; child = child.NextSibling {
		if child.Type != html.ElementNode {
			continue
		}

		//nolint:exhaustive // We only care about table-specific atoms; default handles the rest
		switch child.DataAtom {
		case atom.Tr:
			rows = append(rows, child)
		case atom.Thead, atom.Tbody, atom.Tfoot:
			for rc := child.FirstChild; rc != nil; rc = rc.NextSibling {
				if rc.Type == html.ElementNode && rc.DataAtom == atom.Tr {
					rows = append(rows, rc)
				}
			}
		default:
			// No additional action for other tags
		}
	}

	return rows
}

// ParseTable parses an HTML table node and extracts it into a Table model.
func ParseTable(tableNode *html.Node) models.Table {
	var table models.Table
	rows := findTableRows(tableNode)

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

// HTML spec limits for span attributes. Clipboard HTML is arbitrary external
// input; clamping prevents a huge colspan from blowing up the grid allocation
// in FlattenTable.
const (
	maxRowSpan = 65534
	maxColSpan = 1000
)

func getSpans(n *html.Node) (rowSpan, colSpan int) {
	rowSpan, colSpan = 1, 1
	for _, attr := range n.Attr {
		if attr.Key == "rowspan" {
			if val, err := strconv.Atoi(attr.Val); err == nil {
				rowSpan = clampSpan(val, maxRowSpan)
			}
		}
		if attr.Key == "colspan" {
			if val, err := strconv.Atoi(attr.Val); err == nil {
				colSpan = clampSpan(val, maxColSpan)
			}
		}
	}
	return rowSpan, colSpan
}

func clampSpan(val, maxVal int) int {
	if val < 1 {
		return 1
	}
	if val > maxVal {
		return maxVal
	}
	return val
}

func findNodes(n *html.Node, atoms ...atom.Atom) []*html.Node {
	var results []*html.Node
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		// Stop descending if we hit a nested table to avoid collecting its cells
		if c.Type == html.ElementNode && c.DataAtom == atom.Table {
			continue
		}

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
		parts := strings.SplitN(style, ":", 2)
		if len(parts) == 2 && strings.TrimSpace(parts[0]) == "text-align" {
			return parseAlignVal(strings.TrimSpace(parts[1]))
		}
	}
	return models.AlignNone
}

func processCellContent(n *html.Node, converter *htmltomarkdown.Converter) string {
	// FR-005: Strip Confluence macros, then serialize the cell's children
	// via html.Render so void elements and edge cases are handled correctly.
	var innerHTML strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		cleaned := cleanCellNode(c)
		if err := html.Render(&innerHTML, cleaned); err != nil {
			return textOnlyContent(n)
		}
	}

	markdown, err := converter.ConvertString(innerHTML.String())
	if err != nil {
		return textOnlyContent(n)
	}

	return strings.TrimSpace(markdown)
}

func textOnlyContent(n *html.Node) string {
	var sb strings.Builder
	renderTextOnly(n, &sb)
	return strings.TrimSpace(sb.String())
}

// cleanCellNode returns a deep copy of n suitable for html.Render. Confluence
// macro elements are replaced by a text node containing only their inner text.
func cleanCellNode(n *html.Node) *html.Node {
	if isConfluenceMacro(n) {
		var sb strings.Builder
		renderTextOnly(n, &sb)
		return &html.Node{Type: html.TextNode, Data: sb.String()}
	}

	clone := &html.Node{
		Type:      n.Type,
		DataAtom:  n.DataAtom,
		Data:      n.Data,
		Namespace: n.Namespace,
		Attr:      append([]html.Attribute(nil), n.Attr...),
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		clone.AppendChild(cleanCellNode(c))
	}

	return clone
}

func isConfluenceMacro(n *html.Node) bool {
	if n.Type != html.ElementNode {
		return false
	}
	for _, attr := range n.Attr {
		if attr.Key == "class" && strings.Contains(attr.Val, "confluence-jim-macro") {
			return true
		}
	}
	return false
}

func renderTextOnly(n *html.Node, sb *strings.Builder) {
	if n.Type == html.TextNode {
		sb.WriteString(n.Data)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		renderTextOnly(c, sb)
	}
}
