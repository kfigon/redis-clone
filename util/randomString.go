package util

import "math/rand"

func RandStringRunes(maxLen int) string {
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	ln := rand.Int31n(int32(maxLen))

    b := make([]rune, ln)
    for i := range b {
        b[i] = letterRunes[rand.Intn(len(letterRunes))]
    }
    return string(b)
}