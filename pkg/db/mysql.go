package db

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// MySQLOptions defines options for mysql database.
type MySQLOptions struct {
	Host                  string        `mapstructure:"host" json:"host" yaml:"host"`
	Username              string        `mapstructure:"username" json:"username" yaml:"username"`
	Password              string        `mapstructure:"password" json:"-" yaml:"password"`
	Database              string        `mapstructure:"database" json:"database" yaml:"database"`
	MaxIdleConnections    int           `mapstructure:"max-idle-connections" json:"max_idle_connections" yaml:"max-idle-connections"`
	MaxOpenConnections    int           `mapstructure:"max-open-connections" json:"max_open_connections" yaml:"max-open-connections"`
	MaxConnectionLifeTime time.Duration `mapstructure:"max-connection-life-time" json:"max_connection_life_time" yaml:"max-connection-life-time"`
	LogLevel              int           `mapstructure:"log-level" json:"log_level" yaml:"log-level"`
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
	logLevel := logger.Silent
	if opts.LogLevel != 0 {
		logLevel = logger.LogLevel(opts.LogLevel)
	}
	db, err := gorm.Open(mysql.Open(opts.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
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
