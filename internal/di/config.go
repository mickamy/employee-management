package di

import "github.com/mickamy/employee-management/internal/config"

type Config struct {
	App      config.App      `inject:""`
	Database config.Database `inject:""`
}
