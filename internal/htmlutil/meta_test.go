package htmlutil

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

func TestParseMetaProperties(t *testing.T) {
	document := `<!DOCTYPE html>
<html lang="sv">
	<head>
		<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
		<link rel="preconnect" href="https://static-cdn.sr.se" />
		<link rel="preconnect" href="https://trafficgateway.research-int.se" />
		<link rel="dns-prefetch" href="https://analytics.codigo.se">

		<meta name="author" content="Sveriges Radio" />

		<meta property="og:url" content="https://sverigesradio.se/textochmusikmedericschuldt" />
		<meta property="og:title" content="Text och musik med Eric Sch&#xFC;ldt - alla avsnitt" />
		<meta property="og:description" content="En timme med den vackraste musiken ackompanjerad av poesi, filosofi och personliga reflektioner." />
		<meta property="og:image" content="https://static-cdn.sr.se/images/4914/74ebbeb2-9948-499b-9bc9-94cffd2d456a.jpg?preset=2048x1152" />
		<meta property="og:type" content="website" />
		<meta property="al:ios:url" content="sesrplay://?json=%7B%22type%22:%22showProgram%22,%22id%22:4914%7D" />
		<meta property="al:android:url" content="sesrplay://play/program/4914" />
		<meta property="al:ios:app_store_id" content="300548244" />
		<meta property="al:android:package" content="se.sr.android" />
		<meta property="al:ios:app_name" content="Sveriges Radio Play" />
		<meta property="al:android:app_name" content="Sveriges Radio Play" />
	</head>
	<body></body>
</html>`

	root, err := html.Parse(strings.NewReader(document))
	require.NoError(t, err)

	actual, err := ParseMetaProperties(root)
	require.NoError(t, err)

	expected := MetaProperties{
		"og:url":              {"https://sverigesradio.se/textochmusikmedericschuldt"},
		"og:title":            {"Text och musik med Eric Schüldt - alla avsnitt"},
		"og:description":      {"En timme med den vackraste musiken ackompanjerad av poesi, filosofi och personliga reflektioner."},
		"og:image":            {"https://static-cdn.sr.se/images/4914/74ebbeb2-9948-499b-9bc9-94cffd2d456a.jpg?preset=2048x1152"},
		"og:type":             {"website"},
		"al:ios:url":          {"sesrplay://?json=%7B%22type%22:%22showProgram%22,%22id%22:4914%7D"},
		"al:android:url":      {"sesrplay://play/program/4914"},
		"al:ios:app_store_id": {"300548244"},
		"al:android:package":  {"se.sr.android"},
		"al:ios:app_name":     {"Sveriges Radio Play"},
		"al:android:app_name": {"Sveriges Radio Play"},
	}

	assert.Equal(t, expected, actual)
}
