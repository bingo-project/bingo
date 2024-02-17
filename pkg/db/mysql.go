package db

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"bingo/internal/pkg/logger"
)

// MySQLOptions defines options for mysql database.
type MySQLOptions struct {
	Host                  string        `mapstructure:"host" json:"host" yaml:"host"`
	Username              string        `mapstructure:"username" json:"username" yaml:"username"`
	Password              string        `mapstructure:"password" json:"-" yaml:"password"`
	Database              string        `mapstructure:"database" json:"database" yaml:"database"`
	MaxIdleConnections    int           `mapstructure:"maxIdleConnections" json:"maxIdleConnections" yaml:"maxIdleConnections"`
	MaxOpenConnections    int           `mapstructure:"maxOpenConnections" json:"maxOpenConnections" yaml:"maxOpenConnections"`
	MaxConnectionLifeTime time.Duration `mapstructure:"maxConnectionLifeTime" json:"maxConnectionLifeTime" yaml:"maxConnectionLifeTime"`
	LogLevel              int           `mapstructure:"logLevel" json:"logLevel" yaml:"logLevel"`
}

// DSN returns mysql dsn.
func (o *MySQLOptions) DSN() string {
	return fmt.Sprintf(`%s:%s@tcp(%s)/%s?charset=utf8&parseTime=%t&loc=%s`,
		o.Username,
		o.Password,
		o.Host,
		o.Database,
		true,
		"Local",
	)
}

// NewMySQL create a new gorm db instance with the given options.
func NewMySQL(opts *MySQLOptions) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(opts.DSN()), &gorm.Config{
		Logger:                                   logger.New(opts.LogLevel),
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(opts.MaxOpenConnections)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(opts.MaxConnectionLifeTime)

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(opts.MaxIdleConnections)

	return db, nil
}
