package dbrepos

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStoreBatch_slicePairs(t *testing.T) {
	t.Parallel()

	input := map[string]string{
		"http://rvl1rax4xfkghd.net/ckkrl": "FQNrnTIk",
		"http://rjslawc35.ru/yw79r4zb":    "bQPvDZBj",
	}

	longURLs := make([]string, 0, len(input))
	shortURLs := make([]string, 0, len(input))
	for longURL, shortURL := range input {
		longURLs = append(longURLs, longURL)
		shortURLs = append(shortURLs, shortURL)
	}

	require.Len(t, longURLs, len(input))
	require.Len(t, shortURLs, len(input))

	for i, longURL := range longURLs {
		require.Equal(t, input[longURL], shortURLs[i])
	}
}
