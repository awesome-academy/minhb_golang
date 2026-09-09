package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func migrationEnableExtensions() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202609090001",
		Migrate: func(tx *gorm.DB) error {
			return execAll(tx,
				`CREATE EXTENSION IF NOT EXISTS btree_gist`,
				`CREATE EXTENSION IF NOT EXISTS pg_trgm`,
			)
		},
		Rollback: func(tx *gorm.DB) error {
			return execAll(tx,
				`DROP EXTENSION IF EXISTS pg_trgm`,
				`DROP EXTENSION IF EXISTS btree_gist`,
			)
		},
	}
}
