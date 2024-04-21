package token

import (
	"encoding/hex"
	"math"
	"math/rand"
)

// generateRandomCharacters generates a random string of length "l" as specified by the param
// src: https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
func generateRandomCharacters(l int) string {
	buff := make([]byte, int(math.Ceil(float64(l)/2)))
	_, _ = rand.Read(buff)
	str := hex.EncodeToString(buff)
	return str[:l]
}
