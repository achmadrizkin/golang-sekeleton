// Package database implements interfaces.DBReadWriter for both MySQL and
// PostgreSQL behind one set of query files. The only meaningful SQL
// difference between the two drivers this service touches is bind
// placeholder syntax (`?` vs `$1, $2, ...`) — everything else (column
// list, NOW(), LIMIT/OFFSET) is identical, so each operation lives in one
// file that builds its placeholders through dialect instead of being
// duplicated per driver.
package database

import (
	"fmt"
	"strings"
)

const (
	DriverMySQL    = "mysql"
	DriverPostgres = "postgres"
)

type dialect struct {
	driver string
}

func newDialect(driver string) dialect { return dialect{driver: driver} }

// arg returns the placeholder for the pos'th (1-indexed) bind parameter in
// a query.
func (d dialect) arg(pos int) string {
	if d.driver == DriverPostgres {
		return fmt.Sprintf("$%d", pos)
	}
	return "?"
}

// args returns count comma-joined placeholders starting at position start.
func (d dialect) args(start, count int) string {
	ph := make([]string, count)
	for i := 0; i < count; i++ {
		ph[i] = d.arg(start + i)
	}
	return strings.Join(ph, ", ")
}
