package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func migrationCreateEnums() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202609090002",
		Migrate: func(tx *gorm.DB) error {
			return execAll(tx,
				`CREATE TYPE user_role AS ENUM ('user', 'admin')`,
				`CREATE TYPE movie_status AS ENUM ('coming_soon', 'now_showing', 'ended')`,
				`CREATE TYPE age_rating AS ENUM ('P', 'K', 'T13', 'T16', 'T18', 'C')`,
				`CREATE TYPE showtime_format AS ENUM ('2D', '3D', 'IMAX')`,
				`CREATE TYPE showtime_status AS ENUM ('scheduled', 'cancelled')`,
				`CREATE TYPE booking_status AS ENUM ('pending', 'confirmed', 'expired')`,
				`CREATE TYPE ticket_status AS ENUM ('held', 'paid', 'released')`,
			)
		},
		Rollback: func(tx *gorm.DB) error {
			return execAll(tx,
				`DROP TYPE IF EXISTS ticket_status`,
				`DROP TYPE IF EXISTS booking_status`,
				`DROP TYPE IF EXISTS showtime_status`,
				`DROP TYPE IF EXISTS showtime_format`,
				`DROP TYPE IF EXISTS age_rating`,
				`DROP TYPE IF EXISTS movie_status`,
				`DROP TYPE IF EXISTS user_role`,
			)
		},
	}
}
