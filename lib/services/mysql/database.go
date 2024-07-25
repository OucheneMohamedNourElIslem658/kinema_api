package mysql

import (
	"fmt"
	"log"

	"github.com/OucheneMohamedNourElIslem658/kinema_api/lib/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var Instance *gorm.DB

func Init() {
	dsn := envs.getDatabaseDSN()

	var err error
	Instance, err = gorm.Open(mysql.Open(dsn))
	if err != nil {
		log.Fatal(err.Error())
	}

	err = migrateTables()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Database connected succesfully!")
}

func migrateTables() error {
	err := Instance.AutoMigrate(
		&models.User{},
		&models.AuthProvider{},
		&models.Actor{},
		&models.Type{},
		&models.Movie{},
		&models.Seat{},
		&models.Hall{},
		&models.Diffusion{},
		&models.Reservation{},
	)
	if err != nil {
		return err
	}
	return nil
}
