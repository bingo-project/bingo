package seeder

type DatabaseSeeder struct {
}

// Signature The name and signature of the seeder.
func (DatabaseSeeder) Signature() string {
	return "DatabaseSeeder"
}

// Run seed the application's database.
func (DatabaseSeeder) Run() error {
	// Call other seeders.

	return nil
}
