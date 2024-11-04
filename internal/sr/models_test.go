package sr

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimeMarshal(t *testing.T) {
	expected := `"/Date(1728810000000)/"`
	actual, err := json.Marshal(&Time{Time: time.Unix(1728810000, 0).UTC()})
	require.NoError(t, err)
	assert.Equal(t, expected, string(actual))
}

func TestTimeUnmarshal(t *testing.T) {
	expected := time.Unix(1728810000, 0).UTC()
	var actual Time
	require.NoError(t, json.Unmarshal([]byte(`"/Date(1728810000000)/"`), &actual))
	assert.Equal(t, expected, actual.Time)
}
