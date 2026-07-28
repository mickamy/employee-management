package converter

import (
	employeev1 "github.com/mickamy/employee-management/gen/employee/v1"
	"github.com/mickamy/employee-management/internal/feature/employee/model"
)

//go:generate go tool mapgen -types=model.Employee:*employeev1.Employee -converter-pkg=../../../lib/converters -output=.

var (
	_ model.Employee
	_ employeev1.Employee
)
