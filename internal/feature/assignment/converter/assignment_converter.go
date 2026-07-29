// Package converter maps assignment models to proto messages. Hand-written:
// the enum and the optional release date sit outside what mapgen generates.
package converter

import (
	assignmentv1 "github.com/mickamy/employee-management/gen/assignment/v1"
	"github.com/mickamy/employee-management/internal/feature/assignment/model"
)

//go:generate go tool mapgen -types=model.Assignment:*assignmentv1.Assignment,model.ManagerTenure:*assignmentv1.ManagerTenure -direction=to -converter-pkg=../../../lib/converters -output=.

var (
	_ model.Assignment
	_ assignmentv1.Assignment
)
