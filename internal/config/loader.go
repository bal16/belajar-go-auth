package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
)

func Get() *Config {
	err := godotenv.Load()

	if err != nil {
		log.Println("Error when loading configuration", err.Error())
	}

	jwtExp, err := strconv.Atoi(os.Getenv("JWT_EXP"))
	if err != nil {
		log.Fatal("Error converting JWT_EXP to integer", err.Error())
	}

	return &Config{
		Server: Server{
			PORT: os.Getenv("PORT"),
		},
		Database: Database{
			HOST: os.Getenv("BLUEPRINT_DB_HOST"),
			NAME: os.Getenv("BLUEPRINT_DB_DATABASE"),
			PORT: os.Getenv("BLUEPRINT_DB_PORT"),
			USER: os.Getenv("BLUEPRINT_DB_USERNAME"),
			PASS: os.Getenv("BLUEPRINT_DB_PASSWORD"),
			TZ:   os.Getenv("BLUEPRINT_DB_TIME_ZONE"),
		},
		JWT: JWT{
			SECRET: os.Getenv("JWT_SECRET"),
			EXP:    jwtExp,
		},
	}
}
