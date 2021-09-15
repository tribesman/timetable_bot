package models

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Event события
type Event struct {
	ID          int            `json:"id" gorm:"type:INT UNSIGNED NOT NULL AUTO_INCREMENT; primaryKey;"`
	TimetableID int            `json:"timetable_id" gorm:"type:INT UNSIGNED NOT NULL; index; comment: Расписание"`
	Title       string         `json:"title" gorm:"type:VARCHAR(70) NOT NULL; comment: Название"`
	Comment     sql.NullString `json:"comment" gorm:"type:VARCHAR(191); comment: Описание"`
	From        string         `json:"from" gorm:"type:TIME NOT NULL; comment: Время начала"`
	To          string         `json:"to" gorm:"type:TIME NOT NULL; comment: Время Окончания"`
	Mon         int            `json:"mon" gorm:"type: TINYINT(1) DEFAULT 0; comment: Понедельник"`
	Tues        int            `json:"tues" gorm:"type: TINYINT(1) DEFAULT 0; comment: Вторник"`
	Wed         int            `json:"wed" gorm:"type: TINYINT(1) DEFAULT 0; comment: Среда"`
	Thurs       int            `json:"thurs" gorm:"type: TINYINT(1) DEFAULT 0; comment: Четверг"`
	Fri         int            `json:"fri" gorm:"type: TINYINT(1) DEFAULT 0; comment: Пятница"`
	Sat         int            `json:"sat" gorm:"type: TINYINT(1) DEFAULT 0; comment: Суббота"`
	Sun         int            `json:"sun" gorm:"type: TINYINT(1) DEFAULT 0; comment: Воскресение"`
	Day         int            `json:"day" gorm:"type: TINYINT(2) DEFAULT 0; comment: День"`
	Month       int            `json:"month" gorm:"type: TINYINT(2) DEFAULT 0; comment: Месяц"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// TableName название таблицы
func (model Event) TableName() string {
	return "events"
}

func (model *Event) BeforeSave(db *gorm.DB) (err error) {
	re, _ := regexp.Compile(`[\*\~<\'\"]`)
	model.Title = strings.Replace(model.Title, "`", "", -1)
	model.Title = re.ReplaceAllString(model.Title, "")

	model.Comment.String = strings.Replace(model.Comment.String, "`", "", -1)
	model.Comment.String = re.ReplaceAllString(model.Comment.String, "")
	return
}

func (model *Event) AfterFind(db *gorm.DB) (err error) {
	model.From = TrimSeconds(model.From)
	model.To = TrimSeconds(model.To)
	return
}

func TrimSeconds(s string) string {
	suffix := ":00"
	if strings.HasSuffix(s, suffix) {
		s = s[:len(s)-len(suffix)]
	}
	return s
}

// GetByPK Поиск по ID
func (model Event) GetByPK(db *gorm.DB, id int) Event {
	db.First(&model, id)
	return model
}

// Show Вывод события
func (model Event) Show(db *gorm.DB) string {

	var tt Timetable
	db.First(&tt, model.TimetableID)

	text := ""
	text = fmt.Sprintf("📅 %s - %s\n C %s По %s", tt.Title, model.Title, model.From, model.To)
	if model.Comment.Valid && model.Comment.String != "" {
		text = fmt.Sprintf("%s\n<code>%s</code>", text, model.Comment.String)
	}
	return text
}

// HM получаем часы и минуты
func (model Event) HM(time string) (int, int) {
	hour := 0
	minute := 0
	_time := strings.Split(time, ":")
	if len(_time) > 0 {
		hour, _ = strconv.Atoi(_time[0])
	}
	if len(_time) > 1 {
		minute, _ = strconv.Atoi(_time[1])
	}
	return hour, minute
}

// HoursDiff разница в часах
func (model Event) HoursDiff(time string, sub string) int {
	diff := 0
	timeH, timeM := model.HM(time)
	subH, subM := model.HM(sub)
	diff = timeH - subH
	if timeM-subM < 0 {
		diff--
	}
	return diff
}

// List События на выбранный день
func (model Event) List(db *gorm.DB, date *time.Time, ttId int) []Event {
	var models []Event
	weekday := int(date.Weekday())
	search := ""
	switch {
	case weekday == 0:
		search = "sun=1"
	case weekday == 1:
		search = "mon=1"
	case weekday == 2:
		search = "tues=1"
	case weekday == 3:
		search = "wed=1"
	case weekday == 4:
		search = "thurs=1"
	case weekday == 5:
		search = "fri=1"
	case weekday == 6:
		search = "sat=1"
	}
	db.Raw("SELECT id, title, comment, DATE_FORMAT(`from`,'%H:%i') as `from`, DATE_FORMAT(`to`,'%H:%i') as `to` FROM `events` WHERE `timetable_id` = ? AND "+search+" ORDER BY `events`.`from` ASC", ttId).Scan(&models)
	return models
}

// SetDay Сохранение повтора в день недели
func (model Event) SetDay(db *gorm.DB, dayOfWeek string) {
	switch {
	case dayOfWeek == "1":
		if model.Mon == 1 {
			db.Model(&model).Update("mon", "0")
		} else {
			db.Model(&model).Update("mon", "1")
		}
	case dayOfWeek == "2":
		if model.Tues == 1 {
			db.Model(&model).Update("tues", "0")
		} else {
			db.Model(&model).Update("tues", "1")
		}
	case dayOfWeek == "3":
		if model.Wed == 1 {
			db.Model(&model).Update("wed", "0")
		} else {
			db.Model(&model).Update("wed", "1")
		}
	case dayOfWeek == "4":
		if model.Thurs == 1 {
			db.Model(&model).Update("thurs", "0")
		} else {
			db.Model(&model).Update("thurs", "1")
		}
	case dayOfWeek == "5":
		if model.Fri == 1 {
			db.Model(&model).Update("fri", "0")
		} else {
			db.Model(&model).Update("fri", "1")
		}
	case dayOfWeek == "6":
		if model.Sat == 1 {
			db.Model(&model).Update("sat", "0")
		} else {
			db.Model(&model).Update("sat", "1")
		}
	default:
		if model.Sun == 1 {
			db.Model(&model).Update("sun", "0")
		} else {
			db.Model(&model).Update("sun", "1")
		}
	}
}

func (model Event) UpdateAttr(db *gorm.DB, attr string, value string) string {
	s := "Данные обновлены"
	if attr == "1" {
		db.Model(&model).Update("title", value)
	}
	if attr == "2" {
		if value == "clear" {
			value = ""
		}
		db.Model(&model).Update("comment", value)
	}
	return s
}
