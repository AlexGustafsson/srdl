package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderOutputPathTemplate(t *testing.T) {
	testCases := []struct {
		Name     string
		Template string
		Values   TemplateValues
		Expected string
		Fails    bool
	}{
		{
			Name:     "Plain path",
			Template: "/path/to/file",
			Expected: "/path/to/file",
		},
		{
			Name:     "Jellyfin example",
			Template: "output/{{ .Subscription.Artist }}/{{ .Subscription.Album }}",
			Values: TemplateValues{
				Subscription: SubscriptionTemplateValues{
					Artist: "Foo",
					Album:  "Bar",
				},
			},
			Expected: "output/Foo/Bar",
		},
		{
			Name:     "Audiobookshelf example",
			Template: "output/{{ .Program.Name }}",
			Values: TemplateValues{
				Program: ProgramTemplateValues{
					Name: "Foo",
				},
			},
			Expected: "output/Foo",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			actual, err := renderOutputPathTemplate(testCase.Template, testCase.Values)
			if testCase.Fails {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.Expected, actual)
			}
		})
	}

}
