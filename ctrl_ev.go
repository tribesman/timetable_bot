package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"./models"
	tgBot "github.com/go-telegram-bot-api/telegram-bot-api"
)

// CtrlEv Расписание
type CtrlEv struct {
}

// Add Добавить
func (ctrl CtrlEv) Add(app *App, update *tgBot.Update) bool {
	if update.CallbackQuery != nil {
		_json := GetCallbackQueryData(update.CallbackQuery.Data)

		// Text
		text := "<b>Добавить новое событие</b>\nВведи назвине собития, остальное можно будет отредактировать"
		msg := tgBot.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, text)
		msg.ParseMode = "html"
		_, err := app.Bot.Send(msg)
		NoPanic(err)

		// MarkUp
		kbd := tgBot.NewInlineKeyboardMarkup(
			tgBot.NewInlineKeyboardRow(
				tgBot.NewInlineKeyboardButtonData("⬅️ back", createCallbackDataJson(&CallbackQueryData{Action: "tt.show", Id: _json.Id})),
			),
		)

		_kbd := tgBot.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, kbd)
		_, err = app.Bot.Send(_kbd)
		NoPanic(err)
		app.Step[app.User.ID] = "ev.create"
		app.TimePadID[app.User.ID] = _json.Id
	}

	return true
}

func (ctrl CtrlEv) Create(app *App, update *tgBot.Update) bool {
	t := time.Now()
	event := models.Event{}
	event.TimetableID = app.TimePadID[app.User.ID]
	event.Title = update.Message.Text
	event.From = fmt.Sprintf("%s:00", t.Format("15"))
	event.To = fmt.Sprintf("%s:00", t.Add(time.Hour*1).Format("15"))

	weekday := int(time.Now().Weekday())
	switch {
	case weekday == 0:
		event.Sun = 1
	case weekday == 1:
		event.Mon = 1
	case weekday == 2:
		event.Tues = 1
	case weekday == 3:
		event.Wed = 1
	case weekday == 4:
		event.Thurs = 1
	case weekday == 5:
		event.Fri = 1
	case weekday == 6:
		event.Sat = 1
	}
	fmt.Printf("%v\n", event)
	app.Db.Save(&event)
	fmt.Printf("%v\n", event)

	app.Step[app.User.ID] = ""
	app.TimePadID[app.User.ID] = 0

	kbd := tgBot.NewInlineKeyboardMarkup(
		tgBot.NewInlineKeyboardRow(
			tgBot.NewInlineKeyboardButtonData("⬅️ back", createCallbackDataJson(&CallbackQueryData{Action: "ev.update", Id: event.ID})),
		),
	)
	text := "Событие добавленно, перейти к редактированию\n"

	msg := tgBot.NewMessage(update.Message.Chat.ID, text)
	msg.ReplyMarkup = kbd
	msg.ParseMode = "markdown"
	_, err := app.Bot.Send(msg)
	NoPanic(err)
	return true
}

func (ctrl CtrlEv) Update(app *App, update *tgBot.Update) bool {
	if update.CallbackQuery != nil {
		_json := GetCallbackQueryData(update.CallbackQuery.Data)

		var event models.Event
		app.Db.First(&event, _json.Id)

		_text := fmt.Sprintf("%s\n\nДля редактирования дней недели по которым повторяется событие, нажми на нужный день.", event.Show(app.Db))
		msg := tgBot.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, _text)
		msg.ParseMode = "html"
		_, err := app.Bot.Send(msg)
		NoPanic(err)

		paramsTitle := make(map[string]string)
		paramsTitle["q"] = "1"
		paramsComment := make(map[string]string)
		paramsComment["q"] = "2"

		// MarkUp
		kbd := tgBot.NewInlineKeyboardMarkup(
			tgBot.NewInlineKeyboardRow(
				ctrl.ParamsDay("Mon", &event),
				ctrl.ParamsDay("Tues", &event),
				ctrl.ParamsDay("Wed", &event),
				ctrl.ParamsDay("Thurs", &event),
			),
			tgBot.NewInlineKeyboardRow(
				ctrl.ParamsDay("Fri", &event),
				ctrl.ParamsDay("Sat", &event),
				ctrl.ParamsDay("Sun", &event),
			),
			tgBot.NewInlineKeyboardRow(
				tgBot.NewInlineKeyboardButtonData("🕔 time", createCallbackDataJson(&CallbackQueryData{Action: "ev.sT", Id: event.ID, Params: paramsComment})),
				tgBot.NewInlineKeyboardButtonData("📝 title", createCallbackDataJson(&CallbackQueryData{Action: "ev.sA", Id: event.ID, Params: paramsTitle})),
				tgBot.NewInlineKeyboardButtonData("📝 comment", createCallbackDataJson(&CallbackQueryData{Action: "ev.sA", Id: event.ID, Params: paramsComment})),
			),
			tgBot.NewInlineKeyboardRow(
				tgBot.NewInlineKeyboardButtonData("⬅️ back", createCallbackDataJson(&CallbackQueryData{Action: "tt.show", Id: event.TimetableID})),
			),
		)

		_kbd := tgBot.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, kbd)
		_, err = app.Bot.Send(_kbd)
		NoPanic(err)
	}

	return true
}

func (ctrl CtrlEv) SetDay(app *App, update *tgBot.Update) bool {
	if update.CallbackQuery != nil {
		_json := GetCallbackQueryData(update.CallbackQuery.Data)

		event := models.Event{}.GetByPK(app.Db, _json.Id)
		event.SetDay(app.Db, _json.Params["d"])

		ctrl.Update(app, update)
	}
	return true
}

func (ctrl CtrlEv) SetTime(app *App, update *tgBot.Update) bool {
	if update.CallbackQuery != nil {
		_json := GetCallbackQueryData(update.CallbackQuery.Data)

		app.Step[app.User.ID] = "ev.saveTime"
		app.EventID[app.User.ID] = _json.Id

		event := models.Event{}.GetByPK(app.Db, _json.Id)
		_text := event.Show(app.Db) + "\n\nВведи время в формате ЧЧ:ММ-ЧЧ:ММ"
		msg := tgBot.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, _text)
		_, err := app.Bot.Send(msg)
		NoPanic(err)

		// MarkUp
		kbd := tgBot.NewInlineKeyboardMarkup(
			tgBot.NewInlineKeyboardRow(
				tgBot.NewInlineKeyboardButtonData("⬅️ back", createCallbackDataJson(&CallbackQueryData{Action: "ev.update", Id: _json.Id})),
			),
		)
		_kbd := tgBot.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, kbd)
		_, err = app.Bot.Send(_kbd)
		NoPanic(err)
	}
	return true
}
func (ctrl CtrlEv) SaveTime(app *App, update *tgBot.Update) bool {
	id := app.EventID[app.User.ID]
	event := models.Event{}.GetByPK(app.Db, id)
	from := ""
	to := ""
	update.Message.Text = strings.Replace(update.Message.Text, " ", "", -1)
	pattern := regexp.MustCompile(`^(?P<From>([0|1]?[0-9]|[2][0-3]):([0-5][0-9]))-(?P<To>([0|1]?[0-9]|[2][0-3]):([0-5][0-9]))$`)
	res := pattern.FindStringSubmatch(update.Message.Text)
	names := pattern.SubexpNames()
	for i := range res {
		if names[i] == "From" {
			from = res[i]
		}
		if names[i] == "To" {
			to = res[i]
		}
	}
	_text := ""
	btn := ""
	if from == "" || to == "" {
		_text = fmt.Sprintf("Не верный формат данных, попробуйте еще раз")
		btn = "⬅️ back"
	} else {
		_text = fmt.Sprintf("Данные сохранены. С: %s - по: %s", from, to)
		event.From = from
		event.To = to
		app.Db.Save(&event)
		btn = "👌 ok"
	}

	kbd := tgBot.NewInlineKeyboardMarkup(
		tgBot.NewInlineKeyboardRow(
			tgBot.NewInlineKeyboardButtonData(btn, createCallbackDataJson(&CallbackQueryData{Action: "ev.update", Id: id})),
		),
	)
	msg := tgBot.NewMessage(update.Message.Chat.ID, _text)
	msg.ReplyMarkup = kbd
	_, err := app.Bot.Send(msg)
	NoPanic(err)

	return true
}

func (ctrl CtrlEv) SetAttr(app *App, update *tgBot.Update) bool {
	if update.CallbackQuery != nil {
		_json := GetCallbackQueryData(update.CallbackQuery.Data)
		event := models.Event{}.GetByPK(app.Db, _json.Id)

		app.Step[app.User.ID] = "ev.saveAttr"
		app.EventID[app.User.ID] = _json.Id
		app.Query[app.User.ID] = _json.Params["q"]

		_text := event.Show(app.Db)
		if _json.Params["q"] == "1" { // Title
			_text = fmt.Sprintf("%s\n\nВведи новое название", _text)
		}
		if _json.Params["q"] == "2" { // Comment
			_text = fmt.Sprintf("%s\n\nВведи новый комментарий или <code>clear</code> что бы удалить старый ", _text)
		}
		msg := tgBot.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, _text)
		msg.ParseMode = "html"
		_, err := app.Bot.Send(msg)
		NoPanic(err)

		// MarkUp
		kbd := tgBot.NewInlineKeyboardMarkup(
			tgBot.NewInlineKeyboardRow(
				tgBot.NewInlineKeyboardButtonData("⬅️ back", createCallbackDataJson(&CallbackQueryData{Action: "ev.update", Id: _json.Id})),
			),
		)
		_kbd := tgBot.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, kbd)
		_, err = app.Bot.Send(_kbd)
		NoPanic(err)

	}
	return true
}

func (ctrl CtrlEv) SaveAttr(app *App, update *tgBot.Update) bool {
	id := app.EventID[app.User.ID]
	query := app.Query[app.User.ID]

	event := models.Event{}.GetByPK(app.Db, id)
	_text := event.Show(app.Db) + "\n\n" + event.UpdateAttr(app.Db, query, update.Message.Text)

	app.Step[app.User.ID] = ""
	kbd := tgBot.NewInlineKeyboardMarkup(
		tgBot.NewInlineKeyboardRow(
			tgBot.NewInlineKeyboardButtonData("👌 ok", createCallbackDataJson(&CallbackQueryData{Action: "ev.update", Id: id})),
		),
	)
	msg := tgBot.NewMessage(update.Message.Chat.ID, _text)
	msg.ParseMode = "html"
	msg.ReplyMarkup = kbd
	_, err := app.Bot.Send(msg)
	NoPanic(err)
	return true
}

// ParamsDay Кнопки для включения повтора по дням недели
func (ctrl CtrlEv) ParamsDay(dayOfWeek string, event *models.Event) tgBot.InlineKeyboardButton {
	params := make(map[string]string)
	params["d"] = DayOfWeekNumber(dayOfWeek)
	data := createCallbackDataJson(&CallbackQueryData{Action: "ev.sD", Id: event.ID, Params: params})
	day := dayOfWeek
	switch {
	case dayOfWeek == "Mon":
		if event.Mon == 1 {
			day = "✅ " + day
		} else {
			day = "☑️ " + day
		}
	case dayOfWeek == "Tues":
		if event.Tues == 1 {
			day = "✅ " + day
		} else {
			day = "☑️ " + day
		}
	case dayOfWeek == "Wed":
		if event.Wed == 1 {
			day = "✅ " + day
		} else {
			day = "☑️ " + day
		}
	case dayOfWeek == "Thurs":
		if event.Thurs == 1 {
			day = "✅ " + day
		} else {
			day = "☑️ " + day
		}
	case dayOfWeek == "Fri":
		if event.Fri == 1 {
			day = "✅ " + day
		} else {
			day = "☑️ " + day
		}
	case dayOfWeek == "Sat":
		if event.Sat == 1 {
			day = "✅ " + day
		} else {
			day = "☑️ " + day
		}
	default:
		if event.Sun == 1 {
			day = "✅ " + day
		} else {
			day = "☑️ " + day
		}
	}
	return tgBot.NewInlineKeyboardButtonData(day, data)
}
