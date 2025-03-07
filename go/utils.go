package main

import (
	"math/rand"
)

func generateRandomString(length int) string {
	const (
		lowercase = "abcdefghijklmnopqrstuvwxyz"
		uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		digits    = "0123456789"
	)

	allChars := lowercase + uppercase + digits

	b := make([]byte, length)

	for i := range b {
		b[i] = allChars[rand.Intn(len(allChars))]
	}

	return string(b)
}
