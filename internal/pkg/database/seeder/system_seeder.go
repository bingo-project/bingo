// ABOUTME: Composite seeder that runs all system initialization seeders.
// ABOUTME: Executes Admin, Api, Role, Menu, and RoleMenu seeders in order.

package seeder

import "github.com/gookit/color"

type SystemSeeder struct{}

func (SystemSeeder) Signature() string {
	return "SystemSeeder"
}

func (SystemSeeder) Run() error {
	seeders := []Seeder{
		AdminSeeder{},
		ApiSeeder{},
		RoleSeeder{},
		MenuSeeder{},
		RoleMenuSeeder{},
	}

	for _, s := range seeders {
		color.Infof("  Running %s...\n", s.Signature())
		if err := s.Run(); err != nil {
			color.Redf("  %s failed: %s\n", s.Signature(), err.Error())

			return err
		}
		color.Successf("  %s done.\n", s.Signature())
	}

	return nil
}
