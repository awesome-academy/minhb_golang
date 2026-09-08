package main

import (
	"flag"
	"fmt"
	"os"

	"cinema-booking/config"
	"cinema-booking/migrations"
	"cinema-booking/pkg/db"
)

func main() {
	direction := flag.String("direction", "up", "migration direction: up or down")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		fail(err)
	}
	database, err := db.Connect(db.Options{
		DSN:          cfg.DatabaseURL,
		MaxOpenConns: 2,
		MaxIdleConns: 1,
	})
	if err != nil {
		fail(err)
	}
	defer func() { _ = db.Close(database) }()

	switch *direction {
	case "up":
		err = migrations.Run(database)
	case "down":
		err = migrations.RollbackLast(database)
	default:
		fail(fmt.Errorf("unsupported migration direction %q; use up or down", *direction))
	}
	if err != nil {
		fail(err)
	}

	fmt.Printf("migration %s completed successfully\n", *direction)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
