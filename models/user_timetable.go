package models

import "time"

type UserTimetable struct {
	ID          int       `json:"id" gorm:"type:INT UNSIGNED NOT NULL AUTO_INCREMENT; primaryKey;"`
	UserId      int       `json:"user_id" gorm:"type:INT UNSIGNED NOT NULL; index; comment: Пользователь"`
	TimetableID int       `json:"timetable_id" gorm:"type:INT UNSIGNED NOT NULL; index; comment: Расписание"`
	Admin       int       `json:"admin" gorm:"type:tinyint(1) DEFAULT 0; comment: Админ?"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName Название таблицы
func (model UserTimetable) TableName() string {
	return "user_timetables"
}
