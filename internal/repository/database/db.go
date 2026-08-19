package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/XSAM/otelsql"
	_ "github.com/go-sql-driver/mysql" // mysql driver registration
	_ "github.com/jackc/pgx/v5/stdlib" // postgres driver registration ("pgx")

	"github.com/fauzie/golang-sekeleton/internal/repository"
	repointerface "github.com/fauzie/golang-sekeleton/internal/repository/interface"
)

// Conf is the RepoFactory input for the database connection. It implements
// repository.RepoFactory (see internal/repository/repository.go), so
// wiring it up is a matter of appending *Conf to the factory list passed
// to repository.NewRepository.
type Conf struct {
	Driver                 string // mysql|postgres
	Host                   string
	Port                   int
	User                   string
	Password               string
	Name                   string
	SSLMode                string // postgres only
	QueryTimeoutSeconds    int
	MaxOpenConns           int
	MaxIdleConns           int
	ConnMaxLifetimeMinutes int
}

// Kind implements repository.RepoFactory.
func (c *Conf) Kind() repository.Kind { return repository.KindDB }

func (c *Conf) dsn() (driverName, dsn string) {
	switch c.Driver {
	case DriverPostgres:
		sslMode := c.SSLMode
		if sslMode == "" {
			sslMode = "disable"
		}
		return "pgx", fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			c.User, c.Password, c.Host, c.Port, c.Name, sslMode)
	default:
		return DriverMySQL, fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local",
			c.User, c.Password, c.Host, c.Port, c.Name)
	}
}

type dbReadWriter struct {
	db           *sql.DB
	dialect      dialect
	queryTimeout time.Duration
}

// Build implements repository.RepoFactory: it opens (otel-instrumented) the
// SQL connection, applies pool settings, and returns interfaces.DBReadWriter.
func (c *Conf) Build() (interface{}, error) {
	driverName, dsn := c.dsn()

	db, err := otelsql.Open(driverName, dsn, otelsql.WithAttributes())
	if err != nil {
		return nil, fmt.Errorf("database: open %s: %w", c.Driver, err)
	}

	maxOpen := c.MaxOpenConns
	if maxOpen <= 0 {
		maxOpen = 25
	}
	maxIdle := c.MaxIdleConns
	if maxIdle <= 0 {
		maxIdle = 10
	}
	lifetimeMin := c.ConnMaxLifetimeMinutes
	if lifetimeMin <= 0 {
		lifetimeMin = 5
	}
	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxIdle)
	db.SetConnMaxLifetime(time.Duration(lifetimeMin) * time.Minute)

	timeoutSec := c.QueryTimeoutSeconds
	if timeoutSec <= 0 {
		timeoutSec = 5
	}

	rw := &dbReadWriter{
		db:           db,
		dialect:      newDialect(c.Driver),
		queryTimeout: time.Duration(timeoutSec) * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), rw.queryTimeout)
	defer cancel()
	if err := rw.db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("database: ping %s: %w", c.Driver, err)
	}

	var _ repointerface.DBReadWriter = rw // compile-time interface check
	return rw, nil
}

func (rw *dbReadWriter) DB() *sql.DB { return rw.db }

func (rw *dbReadWriter) Close() error { return rw.db.Close() }

func (rw *dbReadWriter) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, rw.queryTimeout)
	defer cancel()
	return rw.db.PingContext(ctx)
}
