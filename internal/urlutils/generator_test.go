package urlutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateKey(t *testing.T) {
	key := GenerateShortURL("https://foo.com", 8)
	assert.Len(t, key, 8, "Generated key should have the correct length")
}
