package htmlutil

import (
	"fmt"

	"golang.org/x/net/html"
)

type MetaProperties map[string][]string

func (m MetaProperties) Get(key string) string {
	values, ok := m[key]
	if !ok {
		return ""
	}

	return values[0]
}

func (m MetaProperties) Set(key string, value string) {
	m[key] = []string{value}
}

func (m MetaProperties) Add(key string, value string) {
	values, ok := m[key]
	if !ok {
		values = make([]string, 0)
	}
	values = append(values, value)
	m[key] = values
}

func ParseMetaProperties(root *html.Node) (MetaProperties, error) {
	head := Match(root, func(node *html.Node) bool {
		return node.Data == "head"
	})
	if head == nil {
		return nil, fmt.Errorf("missing head")
	}

	properties := make(MetaProperties)
	for child := head.FirstChild; child != nil; child = child.NextSibling {
		if child.Data == "meta" {
			property, propertyExists := Attr(child, "property")
			content, contentExists := Attr(child, "content")

			if propertyExists && contentExists {
				properties.Add(property, content)
			}
		}
	}

	return properties, nil
}
