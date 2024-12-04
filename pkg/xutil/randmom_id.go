package xutil

import (
	"github.com/google/uuid"
)

func GetRandomID() uint32 {
	uRandom, err := uuid.NewRandom()
	if err != nil {
		return 0
	}
	return uRandom.ID()
}
