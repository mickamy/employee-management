package model

import (
	"time"

	"github.com/google/uuid"
)

type Employee struct {
	ID      uuid.UUID `map:"Id"`
	Code    string
	Name    string
	Email   string
	HiredOn time.Time
}

type EmployeeHire struct {
	ID         uuid.UUID
	EmployeeID uuid.UUID
	HiredOn    time.Time
}
