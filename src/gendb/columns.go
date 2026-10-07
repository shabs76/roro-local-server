package gendb

import (
	"log/slog"
	"sync"
)

var columnCache sync.Map

// ColumnExists reports whether table has column in the current database. The answer
// is cached; a failed lookup is not cached and reports false.
//
// Code that uses columns added by migrations checks them here, so it keeps working
// when the database user cannot run ALTER TABLE and the migration has not been
// applied yet.
func ColumnExists(table, column string) bool {
	key := table + "." + column
	if v, ok := columnCache.Load(key); ok {
		return v.(bool)
	}

	db, err := InitDb()
	if err != nil {
		return false
	}

	var n int
	err = db.QueryRow(
		"SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?",
		table, column).Scan(&n)
	if err != nil {
		slog.Error("Column lookup failed", "table", table, "column", column, "error", err)
		return false
	}

	columnCache.Store(key, n > 0)
	return n > 0
}

func resetColumnCache() {
	columnCache.Clear()
}
