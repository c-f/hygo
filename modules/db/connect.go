package db

import (
	"context"
	"database/sql"
	"time"
)

// Connect connects to a db and checks if the credentials are valid or not
func Connect(driverName string, dataSourceName string, timeout time.Duration) (ok bool, err error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return
	}
	defer db.Close()
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return
	}
	return true, nil
}
