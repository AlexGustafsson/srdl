package mp4

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMetadata(t *testing.T) {
	metadata := Metadata{
		Title:       "A very long title for testing",
		Artist:      "Some artist",
		Album:       "Album",
		Description: "Some description",
		Copyright:   "2024",
		Released:    time.Now(),
	}

	file, err := os.OpenFile("../../output.mp4", os.O_RDWR, 0644)
	require.NoError(t, err)
	defer file.Close()

	n := time.Now()
	require.NoError(t, metadata.Write(file))
	fmt.Println(time.Since(n))
}
