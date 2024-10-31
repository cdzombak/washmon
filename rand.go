package main

import "math/rand"

func RandAlnumString(n int) string {
	const includeBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	b := make([]byte, n)
	for i := range b {
		b[i] = includeBytes[rand.Intn(len(includeBytes))]
	}
	return string(b)
}
