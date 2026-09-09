package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func Run(db *gorm.DB) error {
	return New(db).Migrate()
}

func RollbackLast(db *gorm.DB) error {
	return New(db).RollbackLast()
}

func New(db *gorm.DB) *gormigrate.Gormigrate {
	return gormigrate.New(
		db,
		&gormigrate.Options{
			TableName:                 "schema_migrations",
			IDColumnName:              "id",
			IDColumnSize:              255,
			UseTransaction:            true,
			ValidateUnknownMigrations: true,
		},
		[]*gormigrate.Migration{
			migrationEnableExtensions(),
			migrationCreateEnums(),
			migrationCreateUsers(),
			migrationCreateMoviesAndGenres(),
			migrationCreateTheatersRoomsSeats(),
			migrationCreateShowtimesAndPrices(),
			migrationCreateBookingsAndTickets(),
		},
	)
}

func execAll(tx *gorm.DB, statements ...string) error {
	for _, statement := range statements {
		if err := tx.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}
