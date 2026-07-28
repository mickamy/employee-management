package server

import (
	"github.com/mickamy/employee-management/internal/di"
	ehandler "github.com/mickamy/employee-management/internal/feature/employee/handler"
	ohandler "github.com/mickamy/employee-management/internal/feature/organization/handler"
)

type Handlers struct {
	_            di.Infra               `inject:"embed"`
	Employee     *ehandler.Employee     `inject:""`
	Organization *ohandler.Organization `inject:""`
}
