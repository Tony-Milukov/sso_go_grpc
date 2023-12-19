package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	op := "migrations.main"
	var dbLink, operator string

	// getting the dbLink from the flag
	flag.StringVar(&dbLink, "db-link", "", "Database connection string")
	// getting the dbLink from the flag
	flag.StringVar(&operator, "op", "", "( down / up ) Drop or create migrations")

	flag.Parse()

	// check if migrations path is NOT defined in flag
	if operator != "" && operator != "down" && operator != "up" {
		panic("op have to be \"up\" \"down\", or nothing (up by default)")
	}

	// start connection
	m, err := migrate.New("file://../../migrations", dbLink)

	// if an error is there on starting connection
	if err != nil {
		panic(fmt.Sprintf("%s: %w", op, err))
	}

	// migrate up or down (up by default)
	switch operator {

	case "up":
		// get the Migrations UP
		err = m.Up()
		break

	case "down":
		// get the Migrations DOWN
		err = m.Down()
		break

	default:
		// get the Migrations UP
		err = m.Up()
		operator = "up"
		break

	}

	// if there was an error on migration
	if err != nil {
		// if there was no change in DB
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("Database is already up-to-date")
		} else {
			// PANIC IF SOMETHING ELSE WENT WRONG
			panic(fmt.Sprintf("%s: %w", op, err))
		}
	}
	//logging the success
	fmt.Printf("Migrated Successfully op: %s", operator)
}
