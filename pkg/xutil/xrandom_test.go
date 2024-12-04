package xutil

import "testing"

func TestRandomKey(t *testing.T) {
	key := "dfslkjlewruh4657342543ljkfdxcglkdshkrjwqelterreterwtweq"
	for i := 0; i < 1000; i++ {
		seed := ReverseRandomKey(key)
		t.Log(seed)
		randomKey := GenerateRandomKey(seed, 16)
		t.Log(randomKey)
	}
}
