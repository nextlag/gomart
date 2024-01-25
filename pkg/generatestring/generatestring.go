package generatestring

import (
	"math/rand"
)

// NewRandomString генерирует случайную строку заданной длины.
func NewRandomString(length int) string {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

func DigitString(minLen, maxLen int) string {
	var rnd *rand.Rand
	var letters = "0123456789"

	slen := rnd.Intn(maxLen-minLen) + minLen

	s := make([]byte, 0, slen)
	i := 0
	for len(s) < slen {
		idx := rnd.Intn(len(letters) - 1)
		char := letters[idx]
		if i == 0 && '0' == char {
			continue
		}
		s = append(s, char)
		i++
	}

	return string(s)
}
