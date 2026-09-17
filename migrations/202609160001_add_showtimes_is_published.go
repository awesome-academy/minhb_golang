package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func migrationAddShowtimesIsPublished() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202609160001",
		Migrate: func(tx *gorm.DB) error {
			return execAll(tx,
				`ALTER TABLE showtimes ADD COLUMN IF NOT EXISTS is_published BOOLEAN NOT NULL DEFAULT false`,
			)
		},
		Rollback: func(tx *gorm.DB) error {
			return execAll(tx,
				`ALTER TABLE showtimes DROP COLUMN IF EXISTS is_published`,
			)
		},
	}
}
