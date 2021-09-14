package models

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Timetable Расписание
type Timetable struct {
	ID        int       `json:"id" gorm:"type:INT UNSIGNED NOT NULL AUTO_INCREMENT; primaryKey;"`
	Title     string    `json:"title" gorm:"type:VARCHAR(70) NOT NULL; comment: Название"`
	UUID      string    `json:"uuid" gorm:"type:VARCHAR(36) NOT NULL; index; comment: uuid"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName Название таблицы
func (model Timetable) TableName() string {
	return "timetables"
}

func (model *Timetable) BeforeSave(db *gorm.DB) (err error) {
	re, _ := regexp.Compile(`[\*\~<\'\"]`)
	model.Title = strings.Replace(model.Title, "`", "", -1)
	model.Title = re.ReplaceAllString(model.Title, "")
	return
}
func (model *Timetable) BeforeCreate(db *gorm.DB) (err error) {
	model.UUID = uuid.New().String()
	return
}

// GetByPK Поиск по ID
func (model Timetable) GetByPK(db *gorm.DB, id int) Timetable {
	db.First(&model, id)
	return model
}

// GetLink deep link
func (model Timetable) GetLink() string {
	return fmt.Sprintf("tt:%d:%s", model.ID, model.UUID)
}

func (model Timetable) Name() string {
	return fmt.Sprintf("📅 %s", model.Title)
}

func (model Timetable) UpdateAttr(db *gorm.DB, attr string, value string) string {
	s := "Данные обновлены"
	if attr == "1" {
		db.Model(&model).Update("title", value)
	}
	return s
}
