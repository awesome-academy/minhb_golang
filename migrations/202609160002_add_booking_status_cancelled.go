package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func migrationAddBookingStatusCancelled() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202609160002",
		Migrate: func(tx *gorm.DB) error {
			return tx.Exec(`ALTER TYPE booking_status ADD VALUE IF NOT EXISTS 'cancelled'`).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return nil
		},
	}
}
