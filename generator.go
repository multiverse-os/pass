package pass

import (
	"crypto/rand"
	"math/big"
)

// Alpha is a-Z and A-Z
var Alpha = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// Num is 0-9
var Num = []rune("0123456789")

// Symbols is all printable symbols
var Symbols = []rune("!#$%&'()*+,-./:;<=>?@[]^_`{|}~")

// RandSeq returns a random sequence of lenght n
func RandSeq(n int, runes []rune) string {
	b := make([]rune, n)
	for i := range b {
		randomInt, err := rand.Int(rand.Reader, big.NewInt(int64(len(runes))))
		if err != nil {
			panic(err)
		}
		b[i] = runes[randomInt.Int64()]
	}
	return string(b)
}
