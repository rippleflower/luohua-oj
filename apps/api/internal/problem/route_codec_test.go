package problem_test

import (
	"testing"

	"github.com/example/oj3/apps/api/internal/problem"
	"github.com/stretchr/testify/require"
)

func TestRouteCodecRoundTrip(t *testing.T) {
	codec := problem.NewRouteCodec("test-salt")
	code := codec.Encode(42)

	require.NotEmpty(t, code)

	decoded, err := codec.Decode(code)
	require.NoError(t, err)
	require.Equal(t, int64(42), decoded)
}

func TestRouteCodecRejectsInvalidCode(t *testing.T) {
	codec := problem.NewRouteCodec("test-salt")

	_, err := codec.Decode("!")
	require.ErrorContains(t, err, "invalid route code")
}
