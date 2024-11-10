package htmlutil

import "golang.org/x/net/html"

// Match matches a node by performing a depth-first search.
func Match(node *html.Node, matchFunc func(node *html.Node) bool) *html.Node {
	if matchFunc(node) {
		return node
	}

	// Check subtree
	if node.FirstChild != nil {
		result := Match(node.FirstChild, matchFunc)
		if result != nil {
			return result
		}
	}

	// Go to next child
	if node.NextSibling != nil {
		result := Match(node.NextSibling, matchFunc)
		if result != nil {
			return result
		}
	}

	return nil
}

// Attr returns the value of an attribute by key and whether or not it exists.
func Attr(node *html.Node, key string) (string, bool) {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val, true
		}
	}

	return "", false
}
