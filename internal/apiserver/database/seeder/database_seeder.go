package seeder

import (
	"github.com/gookit/color"
)

type DatabaseSeeder struct {
}

// Signature The name and signature of the seeder.
func (DatabaseSeeder) Signature() string {
	return "DatabaseSeeder"
}

// Run seed the application's database.
func (DatabaseSeeder) Run() error {
	// Call other seeders.
	err := AdminSeeder{}.Run()
	if err != nil {
		color.Redf("%s failed: %s", AdminSeeder{}.Signature(), err.Error())
	}

	return nil
}
