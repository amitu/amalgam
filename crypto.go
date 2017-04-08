package acko

import (
	"math/rand"
)

var (
	LETTERS = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	DIGITS  = []byte("1234567890")
)

func GetRandomString(size int) string {
	array := append(LETTERS, DIGITS...)
	b := make([]byte, size)
	for i := range b {
		b[i] = array[rand.Intn(len(array))]
	}
	return string(b)
}
