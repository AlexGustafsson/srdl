package sr

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClientListEpisodesInProgram(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	result, err := DefaultClient.ListEpisodesInProgram(context.TODO(), 4914, nil)
	require.NoError(t, err)

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	require.NoError(t, encoder.Encode(&result))
}
