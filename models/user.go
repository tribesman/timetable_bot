package models

import (
	"database/sql"
	"errors"
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID         int            `json:"id" gorm:"type:INT UNSIGNED NOT NULL AUTO_INCREMENT; primaryKey;"`
	TelegramId int            `json:"telegram_id" gorm:"type:INT UNSIGNED; uniqueIndex; comment: Chat_id"`
	Username   sql.NullString `json:"username" gorm:"type: VARCHAR(70) DEFAULT NULL; comment: Имя пользователя"`
	Active     bool           `json:"active" gorm:"type: TINYINT(1); comment: Вывод?"`
	TempPwd    sql.NullString `json:"temp_pwd" gorm:"type: VARCHAR(32) DEFAULT NULL; comment: Временный пароль"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

func (model User) TableName() string {
	return "users"
}

// GetOrInsert Поиск или создание юзера по chat_id
func (model User) GetOrInsert(db *gorm.DB, chatId int, username string) User {
	result := db.First(&model, "telegram_id = ?", chatId)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		model.TelegramId = chatId
		model.Active = true
		model.Username = sql.NullString{String: username, Valid: true}
		db.Create(&model)
	} else {
		// Проверка актуальности юзернайма
		if model.Username.String != username {
			model.Username.String = username
			db.Save(&model)
		}
	}
	return model
}

// GetTimetables Получаем расписания для пользователя
func (model User) GetTimetables(db *gorm.DB) []Timetable {
	var models []Timetable
	utt := "`" + UserTimetable{}.TableName() + "`"
	tt := "`" + Timetable{}.TableName() + "`"
	db.Raw("SELECT "+tt+".* FROM "+tt+" LEFT JOIN "+utt+" ON "+tt+".id = "+utt+".timetable_id WHERE "+utt+".user_id = ?", model.ID).Scan(&models)
	return models
}

// IsAdmin Получаем расписания для пользователя
func (model User) IsAdmin(db *gorm.DB, tpId int) bool {
	var _model UserTimetable
	db.Raw("SELECT * FROM `"+UserTimetable{}.TableName()+"` WHERE `user_id` = ? AND `timetable_id` = ?", model.ID, tpId).Scan(&_model)
	return _model.Admin == 1
}
