package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func migrationAddBookingsOnePendingPerUserShowtime() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202609210001",
		Migrate: func(tx *gorm.DB) error {
			return execAll(tx,
				`CREATE UNIQUE INDEX IF NOT EXISTS bookings_one_pending_per_user_showtime ON bookings (user_id, showtime_id) WHERE status = 'pending'`,
			)
		},
		Rollback: func(tx *gorm.DB) error {
			return execAll(tx,
				`DROP INDEX IF EXISTS bookings_one_pending_per_user_showtime`,
			)
		},
	}
}
