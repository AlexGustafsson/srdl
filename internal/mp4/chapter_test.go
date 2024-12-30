package mp4

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChapterWrite(t *testing.T) {
	index := Index{
		Chapters: []Chapter{
			{
				Name: "01 - First",
			},
			{
				Name: "02 - Second",
			},
		},
	}

	// target := filepath.Join(t.TempDir(), "with-chapters.m4a")
	target := "with-chapters.m4a"
	require.NoError(t, copyFile("./empty.m4a", target))

	file, err := os.OpenFile(target, os.O_RDWR, 0)
	require.NoError(t, err)
	defer file.Close()

	require.NoError(t, index.Write(file))
}

func TestFormatChapter(t *testing.T) {
	expected := []byte("\x00\x03Foo\x00\x00\x00\x0c\x65encd\x00\x00\x01")
	actual := Chapter{
		Name: "Foo",
	}.Bytes()

	assert.Equal(t, expected, actual)
}

func TestFormatIndex(t *testing.T) {
	expected := []byte("\x00\x03Foo\x00\x00\x00\x0c\x65encd\x00\x00\x01\x00\x03Bar\x00\x00\x00\x0c\x65encd\x00\x00\x01")
	actual := Index{
		Chapters: []Chapter{
			{
				Name: "Foo",
			},
			{
				Name: "Bar",
			},
		},
	}.Bytes()

	assert.Equal(t, expected, actual)
}
