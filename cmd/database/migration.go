package main

import (
	"auth/internal/config"
	"auth/internal/connection"
	"log"
	"os"
)

func main() {
	cnf := config.Get()
	_, db := connection.GetDatabase(cnf.Database)
	defer db.Close()

	query, err := os.ReadFile("sql/migration.sql")
	if err != nil {
		log.Fatal("Error reading migration file: ", err.Error())
	}

	log.Println("Migrating database...")
	_, err = db.Exec(string(query))
	if err != nil {
		log.Fatal("Error executing migration: ", err.Error())
	}
	
	log.Println("Database migration completed successfully.")
}
