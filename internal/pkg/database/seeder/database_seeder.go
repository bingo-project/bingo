package seeder

import (
	"github.com/gookit/color"
	"gorm.io/gorm"

	"github.com/bingo-project/bingo/internal/pkg/store"
)

// Init initializes the seeder with database connection.
func Init(db *gorm.DB) {
	store.NewStore(db)
}

// Seeder defines the interface for database seeders.
type Seeder interface {
	Signature() string
	Run() error
}

// Seeders is the registry of all available seeders.
var Seeders = []Seeder{
	AdminSeeder{},
	ApiSeeder{},
	RoleSeeder{},
	MenuSeeder{},
	RoleMenuSeeder{},
	SystemSeeder{},
}

type DatabaseSeeder struct {
}

// Signature The name and signature of the seeder.
func (DatabaseSeeder) Signature() string {
	return "DatabaseSeeder"
}

// Run seed the application's database.
func (DatabaseSeeder) Run() error {
	return RunSeeders("")
}

// RunSeeders runs seeders. If name is empty, runs all; otherwise runs matching seeder.
func RunSeeders(name string) error {
	for _, s := range Seeders {
		if name != "" && s.Signature() != name {
			continue
		}

		color.Infof("Running %s...\n", s.Signature())
		if err := s.Run(); err != nil {
			color.Redf("%s failed: %s\n", s.Signature(), err.Error())
		} else {
			color.Successf("%s done.\n", s.Signature())
		}
	}

	return nil
}
