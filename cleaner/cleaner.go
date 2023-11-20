package cleaner

import (
	"bytes"
	"strings"
	"unicode"

	"golang.org/x/net/html"
)

func cleanNode(n *html.Node) {
	var next *html.Node
	for c := n.FirstChild; c != nil; c = next {
		next = c.NextSibling

		if c.Type == html.ElementNode && (c.Data == "script" || c.Data == "style") {
			n.RemoveChild(c)
			continue
		}

		if c.Type == html.CommentNode {
			n.RemoveChild(c)
			continue
		}

		if c.Type == html.ElementNode {
			c.Attr = nil
		}

		cleanNode(c)
	}

	if n.Type == html.ElementNode && isEmptyNode(n) {
		n.Parent.RemoveChild(n)
	}
}

func isEmptyNode(n *html.Node) bool {
	return n.Type == html.TextNode && strings.TrimSpace(n.Data) == "" || n.FirstChild == nil
}

func CleanHTML(htmlContent string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", err
	}

	cleanNode(doc)

	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return "", err
	}

	return removeExtraWhitespace(buf.String()), nil
}

func removeExtraWhitespace(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	wasSpace := false

	for _, r := range s {
		if unicode.IsSpace(r) {
			if !wasSpace {
				b.WriteRune(' ')
				wasSpace = true
			}
		} else {
			b.WriteRune(r)
			wasSpace = false
		}
	}

	return b.String()
}
