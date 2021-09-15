package main

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"log"
	"reflect"
	"strings"
	"time"

	"./models"
	tgBot "github.com/go-telegram-bot-api/telegram-bot-api"
)

type stubMapping map[string]interface{}

type CallbackQueryData struct {
	Action string            `json:"a"`
	Id     int               `json:"id"`
	Params map[string]string `json:"p"`
}

var StubStorage = stubMapping{}

func main() {
	app := App{
		Step:          make(map[int]string),
		TimePadID:     make(map[int]int),
		EventID:       make(map[int]int),
		Query:         make(map[int]string),
		CallaBackData: make(map[string]string),
	}

	Start(&app)

	bot, err := tgBot.NewBotAPI(app.Cfg.TelegramApiKey)
	Panic(err)
	bot.Debug = app.Cfg.BotDebug
	log.Printf("Authorized on account %s", bot.Self.UserName)
	app.Bot = *bot

	StubStorage = map[string]interface{}{
		"tt.index":    CtrlTt.Index,
		"tt.week":     CtrlTt.Week,
		"tt.add":      CtrlTt.Add,
		"tt.create":   CtrlTt.Create,
		"tt.show":     CtrlTt.Show,
		"tt.update":   CtrlTt.Update,
		"tt.sA":       CtrlTt.SetAttr,
		"tt.saveAttr": CtrlTt.SaveAttr,
		"tt.rm":       CtrlTt.Rm,
		"tt.del":      CtrlTt.Delete,
		"ev.add":      CtrlEv.Add,
		"ev.create":   CtrlEv.Create,
		"ev.update":   CtrlEv.Update,
		"ev.sD":       CtrlEv.SetDay,
		"ev.sA":       CtrlEv.SetAttr,
		"ev.saveAttr": CtrlEv.SaveAttr,
		"ev.sT":       CtrlEv.SetTime,
		"ev.saveTime": CtrlEv.SaveTime,
	}

	u := tgBot.NewUpdate(0)
	u.Timeout = 60
	updates, _ := bot.GetUpdatesChan(u)

	for update := range updates {
		fullTimeStart := time.Now()

		if update.CallbackQuery == nil && update.Message == nil {
			continue
		}
		var from *tgBot.User
		if update.CallbackQuery != nil {
			from = update.CallbackQuery.From
		}
		if update.Message != nil {
			from = update.Message.From
		}
		app.User = models.User{}.GetOrInsert(app.Db, from.ID, from.UserName)

		// Обработка CallbackQuery
		if update.CallbackQuery != nil {
			log.Printf("%v\n", update.CallbackQuery.Data)
			callbackJson := GetCallbackQueryData(update.CallbackQuery.Data)

			if callbackJson.Action == "" {
				continue
			}

			if CallHelper(callbackJson.Action, &app, &update) == true {
				continue
			}
		}

		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		if app.Step[app.User.ID] != "" {
			if CallHelper(app.Step[app.User.ID], &app, &update) == true {
				continue
			}
		}

		if update.Message.IsCommand() { // https://go-telegram-bot-api.dev/examples/command-handling.html
		}
		if strings.HasPrefix(update.Message.Text, "tt:") {
			//_text := ""
			text := update.Message.Text[len("tt:"):len(update.Message.Text)]
			params := strings.Split(text, ":")
			if len(params) > 1 {
				var model models.Timetable
				result := app.Db.First(&model, "id = ? AND uuid = ?", params[0], params[1])
				fmt.Println(model)
				if result.Error == nil {
					var utt models.UserTimetable
					_result := app.Db.First(&utt, "user_id = ? AND timetable_id = ?", app.User.ID, model.ID)
					if errors.Is(_result.Error, gorm.ErrRecordNotFound) {
						utt.UserId = app.User.ID
						utt.TimetableID = model.ID
						app.Db.Create(&utt)
					}
				}
				_ = CtrlTt{}.Index(&app, &update)
				continue
			}

		}
		_ = CtrlTt{}.Index(&app, &update)

		fullTimeEnd := time.Now()
		log.Printf("Request time: %v\n", fullTimeEnd.Sub(fullTimeStart))
	}

}

func CallHelper(action string, app *App, update *tgBot.Update) bool {
	_continue := false
	var err error

	// TimePad
	if strings.HasPrefix(action, "tt.") {
		_, err = Call(action, CtrlTt{}, app, update)
		_continue = true
	}

	// Event
	if strings.HasPrefix(action, "ev.") {
		_, err = Call(action, CtrlEv{}, app, update)
		_continue = true
	}
	Panic(err)

	return _continue
}

func Call(funcName string, params ...interface{}) (result interface{}, err error) {
	f := reflect.ValueOf(StubStorage[funcName])
	if len(params) != f.Type().NumIn() {
		err = fmt.Errorf("the number of params is out of index (%d)", f.Type().NumIn())
		return
	}
	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}
	//var res []reflect.Value
	res := f.Call(in)
	result = res[0].Interface()
	return
}

// DayOfWeekNumber Номер дня недели строкой
func DayOfWeekNumber(dayOfWeek string) string {
	switch {
	case dayOfWeek == "Mon":
		return "1"
	case dayOfWeek == "Tues":
		return "2"
	case dayOfWeek == "Wed":
		return "3"
	case dayOfWeek == "Thurs":
		return "4"
	case dayOfWeek == "Fri":
		return "5"
	case dayOfWeek == "Sat":
		return "6"
	default:
		return "0"
	}
}

/*
// Вывод структуры телеграм сообщения
func showUpdateStruct(update *tgBot.Update, show bool) {
	if show {
		empJSON, err := json.MarshalIndent(update, "", "  ")
		if err != nil {
			log.Fatalf(err.Error())
		}
		fmt.Printf("%s\n", string(empJSON))
	}
}
*/
