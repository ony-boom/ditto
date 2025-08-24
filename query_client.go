package main

import (
	"context"
	"database/sql"
	_ "embed"
	"log"
	"path/filepath"
	"sync"

	"github.com/adrg/xdg"
	"github.com/ony-boom/ditto/database"
	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var ddl string

var (
	queries *database.Queries
	once        sync.Once
)

func NewQueryClient() *database.Queries {
	once.Do(func() {
		ctx := context.Background()
		db, err := sql.Open("sqlite", filepath.Join(xdg.ConfigHome, "ditto", "ditto.db"))
		if err != nil {
			log.Fatal(err)
		}
		if _, err := db.ExecContext(ctx, ddl); err != nil {
			log.Fatal(err)
		}
		queries = database.New(db)
	})
	return queries
}
