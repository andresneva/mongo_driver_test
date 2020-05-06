package stage

import (
	"math/rand"
	"time"
)

const idLength = 32
const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

//GenerateID for a new document
func GenerateID() string {
	b := make([]byte, idLength)

	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}

	return string(b)
}

//GenerateString to fill a new document
func GenerateString(size int) string {
	b := make([]byte, size*1024)

	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}

	return string(b)
}
