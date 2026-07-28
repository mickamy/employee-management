package model

import (
	"github.com/google/uuid"
)

type Department struct {
	ID   uuid.UUID `map:"Id"`
	Code string
	Name string
}
