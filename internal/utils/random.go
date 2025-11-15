package utils

import "math/rand"

func PickRandom[T any](slice []T, n int) []T {
	if len(slice) == 0 {
		return nil
	}
	if n >= len(slice) {
		return slice
	}

	rand.Shuffle(len(slice), func(i, j int) {
		slice[i], slice[j] = slice[j], slice[i]
	})
	return slice[:n]
}

func Ptr[T any](v T) *T {
	return &v
}
