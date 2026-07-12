package crawler

import (
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

type ParsedPage struct {
	URL   string
	Text  string
	Links []string
}

func Parse(rawHTML string, baseURL string) (*ParsedPage, error) {
	doc, err := html.Parse(strings.NewReader(rawHTML))
	if err != nil {
		return nil, err
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	page := &ParsedPage{URL: baseURL}
	var textBuilder strings.Builder

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			text := strings.TrimSpace(n.Data)
			if text != "" {
				textBuilder.WriteString(text)
				textBuilder.WriteString(" ")
			}
		}

		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					resolved := resolveLink(base, attr.Val)
					if resolved != "" {
						page.Links = append(page.Links, resolved)
					}
				}
			}
		}

		if n.Type == html.ElementNode && (n.Data == "script" || n.Data == "style") {
			return
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(doc)
	page.Text = strings.TrimSpace(textBuilder.String())

	return page, nil
}

func resolveLink(base *url.URL, href string) string {
	href = strings.TrimSpace(href)

	if href == "" || strings.HasPrefix(href, "#") {
		return ""
	}

	parsed, err := url.Parse(href)
	if err != nil {
		return ""
	}

	resolved := base.ResolveReference(parsed)

	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return ""
	}

	return resolved.String()
}
