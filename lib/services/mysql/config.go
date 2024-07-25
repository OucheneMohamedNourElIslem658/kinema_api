package mysql

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type config struct {
	dbUser     string
	dbPassword string
	dbHost     string
	dbPort     string
	dbName     string
}

var envs = initAPI()

func initAPI() config {
	godotenv.Load()
	return config{
		dbUser:     os.Getenv("DB_USER"),
		dbPassword: os.Getenv("DB_PASSWORD"),
		dbHost:     os.Getenv("DB_HOST"),
		dbPort:     os.Getenv("DB_PORT"),
		dbName:     os.Getenv("DB_NAME"),
	}
}

func (config config) getDatabaseDSN() string {
	return fmt.Sprintf(
		"%v:%v@tcp(%v:%v)/%v?charset=utf8mb4&parseTime=True&loc=Local",
		config.dbUser,
		config.dbPassword,
		config.dbHost,
		config.dbPort,
		config.dbName)
}
