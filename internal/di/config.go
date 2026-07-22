package di

import "github.com/mickamy/employee-management/internal/config"

type Config struct {
	Database config.DatabaseConfig `inject:""`
}
