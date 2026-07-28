package converter

import (
	organizationv1 "github.com/mickamy/employee-management/gen/organization/v1"
	"github.com/mickamy/employee-management/internal/feature/organization/model"
)

//go:generate go tool mapgen -types=model.Department:*organizationv1.Department -converter-pkg=../../../lib/converters -output=.

var (
	_ model.Department
	_ organizationv1.Department
)
