package fixture

import (
	"time"

	"github.com/mickamy/employee-management/internal/feature/employee/model"
)

//go:generate go tool standin -source ../model -destination ./

func EmployeeAndHire() (model.Employee, model.EmployeeHire) {
	employee := Employee(func(m *model.Employee) {
		m.HiredOn = m.HiredOn.Truncate(24 * time.Hour)
	})
	hire := EmployeeHire(func(m *model.EmployeeHire) {
		m.EmployeeID = employee.ID
		m.HiredOn = employee.HiredOn
	})
	return employee, hire
}
