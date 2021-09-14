package main

import (
	"log"

	"github.com/go-gormigrate/gormigrate"
	"gorm.io/gorm"

	"./models"
)

func Migrate(app *App) {

	m := gormigrate.New(app.Db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "1630930445",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&models.User{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable(models.User{}.TableName())
			},
		},
		{
			ID: "1631023853",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&models.Timetable{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable(models.Timetable{}.TableName())
			},
		},
		{
			ID: "1631023855",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&models.UserTimetable{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable(models.UserTimetable{}.TableName())
			},
		},
		{
			ID: "1631098228",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&models.Event{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable(models.Event{}.TableName())
			},
		},
	})

	if err := m.Migrate(); err != nil {
		log.Fatalf("Could not migrate: %v", err)
	}
	log.Printf("Migration did run successfully")
}
