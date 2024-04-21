package token

import (
	"github.com/magiconair/properties/assert"
	"testing"
)

func Test_generateRandomCharacters(t *testing.T) {
	strLen := 6
	randomString := generateRandomCharacters(strLen)
	assert.Equal(t, len(randomString), strLen)

	strLen = 12
	randomString = generateRandomCharacters(strLen)
	assert.Equal(t, len(randomString), strLen)
}
