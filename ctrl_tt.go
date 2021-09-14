package main

import (
	"fmt"
	"strconv"
	"time"

	"./models"
	tgBot "github.com/go-telegram-bot-api/telegram-bot-api"
)

// CtrlTt Расписание
type CtrlTt struct {
}

func (ctrl CtrlTt) Index(app *App, update *tgBot.Update) bool {

	timePads := app.User.GetTimetables(app.Db)
	kbd := tgBot.InlineKeyboardMarkup{}
	for _, tp := range timePads {
		var row []tgBot.InlineKeyboardButton
		btn := tgBot.NewInlineKeyboardButtonData(tp.Title, createCallbackDataJson(&CallbackQueryData{Action: "tt.show", Id: tp.ID}))
		row = append(row, btn)
		kbd.InlineKeyboard = append(kbd.InlineKeyboard, row)
	}
	addBtn := tgBot.NewInlineKeyboardButtonData("🗓 Add", app.CallaBackData["tt.add"])
	kbd.InlineKeyboard = append(kbd.InlineKeyboard, []tgBot.InlineKeyboardButton{addBtn})

	text := "👋 Привет, я бот для удобного хранения расписания. Выбери расписание или создай новое *новое*.\n"
	if update.CallbackQuery != nil {
		// Text
		msg := tgBot.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, text)
		msg.ParseMode = "markdown"

		// MarkUp
		app.Bot.Send(msg)
		_kbd := tgBot.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, kbd)
		app.Bot.Send(_kbd)
	}
	if update.Message != nil {
		msg := tgBot.NewMessage(update.Message.Chat.ID, text)
		msg.ReplyMarkup = kbd
		msg.ParseMode = "markdown"
		app.Bot.Send(msg)
	}
	return true
}

func (ctrl CtrlTt) Add(app *App, update *tgBot.Update) bool {
	if update.CallbackQuery != nil {
		// Text
		msg := tgBot.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "Добавить новое расписание\n Введи название расписания (только буквы и числа)")
		msg.ParseMode = "html"
		app.Bot.Send(msg)

		// MarkUp
		kbd := tgBot.NewInlineKeyboardMarkup(
			tgBot.NewInlineKeyboardRow(
				tgBot.NewInlineKeyboardButtonData("⬅️ back", app.CallaBackData["index"]),
			),
		)

		_kbd := tgBot.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, kbd)
		app.Bot.Send(_kbd)
		app.Step[app.User.ID] = "tt.create"
	}

	return true
}

// Create a new
func (ctrl CtrlTt) Create(app *App, update *tgBot.Update) bool {
	title := update.Message.Text

	model := models.Timetable{Title: title}
	app.Db.Save(&model)

	userTimePad := models.UserTimetable{UserId: app.User.ID, TimetableID: model.ID, Admin: 1}
	app.Db.Save(&userTimePad)

	ctrl.Index(app, update)
	app.Step[app.User.ID] = ""
	return true
}

func (ctrl CtrlTt) Show(app *App, update *tgBot.Update) bool {
	if update.CallbackQuery != nil {
		_json := GetCallbackQueryData(update.CallbackQuery.Data)

		model := models.Timetable{}.GetByPK(app.Db, _json.Id)
		isAdmin := app.User.IsAdmin(app.Db, model.ID)

		// Текущий день
		_day, _ := _json.Params["day"]
		day, err := strconv.Atoi(_day)
		if err != nil {
			day = 0
		}
		next := make(map[string]string)
		next["day"] = fmt.Sprint(day + 1)
		prev := make(map[string]string)
		prev["day"] = fmt.Sprint(day - 1)

		date := time.Now()
		date = date.AddDate(0, 0, day)

		// Заголовок
		_text := fmt.Sprintf("%s", date.Format("02 January 2006 Monday"))
		if day == -1 {
			_text = fmt.Sprintf("%s ⬅️", _text)
		}
		if day == 0 {
			_text = fmt.Sprintf("%s ⏺", _text)
		}
		if day == 1 {
			_text = fmt.Sprintf("%s ➡️", _text)
		}
		_text = fmt.Sprintf("📅 %s - %s\n-----\n", model.Title, _text)

		events := models.Event.List(models.Event{}, app.Db, &date, model.ID)
		if len(events) < 1 {
			_text = fmt.Sprintf("%s - Пока нет ни каких дел\n", _text)
		} else {
			for i, event := range events {
				_text = fmt.Sprintf("%s - <b>#%d</b> %s-%s %s\n", _text, i+1, event.From, event.To, event.Title)
				if event.Comment.Valid && event.Comment.String != "" {
					_text = fmt.Sprintf("%s - <code>%s</code>\n", _text, event.Comment.String)
				}
			}
		}

		// Ссылка
		if isAdmin {
			_text = fmt.Sprintf("%s-----\n", _text)

		}

		msg := tgBot.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, _text)
		msg.ParseMode = "html"
		app.Bot.Send(msg)

		kbd := tgBot.InlineKeyboardMarkup{}
		nav := tgBot.NewInlineKeyboardRow(
			tgBot.NewInlineKeyboardButtonData("⬅️", createCallbackDataJson(&CallbackQueryData{Action: "tt.show", Id: model.ID, Params: prev})),
			tgBot.NewInlineKeyboardButtonData("⏺", createCallbackDataJson(&CallbackQueryData{Action: "tt.show", Id: model.ID})),
			tgBot.NewInlineKeyboardButtonData("➡️", createCallbackDataJson(&CallbackQueryData{Action: "tt.show", Id: model.ID, Params: next})),
		)
		kbd.InlineKeyboard = append(kbd.InlineKeyboard, nav)
		var kbdRows []tgBot.InlineKeyboardButton
		for i, event := range events {
			btnText := fmt.Sprintf("📝 #%d", i+1)
			btn := tgBot.NewInlineKeyboardButtonData(btnText, createCallbackDataJson(&CallbackQueryData{Action: "ev.update", Id: event.ID}))
			kbdRows = append(kbdRows, btn)
		}

		_kbdRows := ChunkInlineKeyboardButton(kbdRows, 3)
		for _, kbdRow := range _kbdRows {
			if len(kbdRow) > 0 {
				kbd.InlineKeyboard = append(kbd.InlineKeyboard, kbdRow)
			}

		}
		actions := tgBot.NewInlineKeyboardRow(
			tgBot.NewInlineKeyboardButtonData("⤴️", app.CallaBackData["index"]),
			tgBot.NewInlineKeyboardButtonData("🆕 new", createCallbackDataJson(&CallbackQueryData{Action: "ev.add", Id: model.ID})),
			tgBot.NewInlineKeyboardButtonData("📝 edit", createCallbackDataJson(&CallbackQueryData{Action: "tt.update", Id: model.ID})),
		)
		kbd.InlineKeyboard = append(kbd.InlineKeyboard, actions)

		_kbd := tgBot.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, kbd)
		app.Bot.Send(_kbd)
	}

	return true
}

func (ctrl CtrlTt) Update(app *App, update *tgBot.Update) bool {
	if update.CallbackQuery != nil {
		_json := GetCallbackQueryData(update.CallbackQuery.Data)
		model := models.Timetable{}.GetByPK(app.Db, _json.Id)

		_text := model.Name() + " - редактирование"
		//_text = fmt.Sprintf("%s\nПоделится расписанием: <code>%s?start=%s</code>", _text, app.Cfg.BotUrl, model.GetLink())
		_text = fmt.Sprintf("%s\nПоделится расписанием: <code>%s</code>", _text, model.GetLink())

		msg := tgBot.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, _text)
		msg.ParseMode = "html"
		app.Bot.Send(msg)

		paramsTitle := make(map[string]string)
		paramsTitle["q"] = "1"

		kbd := tgBot.InlineKeyboardMarkup{}
		actions := tgBot.NewInlineKeyboardRow(
			tgBot.NewInlineKeyboardButtonData("⤴️", createCallbackDataJson(&CallbackQueryData{Action: "tt.show", Id: model.ID})),
			tgBot.NewInlineKeyboardButtonData("📝 title", createCallbackDataJson(&CallbackQueryData{Action: "tt.sA", Id: model.ID, Params: paramsTitle})),
			tgBot.NewInlineKeyboardButtonData("🗑 delete", createCallbackDataJson(&CallbackQueryData{Action: "tt.rm", Id: model.ID})),
		)
		kbd.InlineKeyboard = append(kbd.InlineKeyboard, actions)

		_kbd := tgBot.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, kbd)
		app.Bot.Send(_kbd)
	}
	return true
}

func (ctrl CtrlTt) SetAttr(app *App, update *tgBot.Update) bool {
	if update.CallbackQuery != nil {
		_json := GetCallbackQueryData(update.CallbackQuery.Data)

		app.Step[app.User.ID] = "tt.saveAttr"
		app.TimePadID[app.User.ID] = _json.Id
		app.Query[app.User.ID] = _json.Params["q"]

		_text := ""
		if _json.Params["q"] == "1" { // Title
			_text = "Введи название"
		}
		msg := tgBot.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, _text)
		app.Bot.Send(msg)

		// MarkUp
		kbd := tgBot.NewInlineKeyboardMarkup(
			tgBot.NewInlineKeyboardRow(
				tgBot.NewInlineKeyboardButtonData("⬅️ back", createCallbackDataJson(&CallbackQueryData{Action: "tt.update", Id: _json.Id})),
			),
		)
		_kbd := tgBot.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, kbd)
		app.Bot.Send(_kbd)

	}
	return true
}

func (ctrl CtrlTt) SaveAttr(app *App, update *tgBot.Update) bool {
	id := app.TimePadID[app.User.ID]
	query := app.Query[app.User.ID]

	model := models.Timetable{}.GetByPK(app.Db, id)
	_text := model.UpdateAttr(app.Db, query, update.Message.Text)

	kbd := tgBot.NewInlineKeyboardMarkup(
		tgBot.NewInlineKeyboardRow(
			tgBot.NewInlineKeyboardButtonData("👌 ok", createCallbackDataJson(&CallbackQueryData{Action: "tt.update", Id: id})),
		),
	)
	msg := tgBot.NewMessage(update.Message.Chat.ID, _text)
	msg.ReplyMarkup = kbd
	app.Bot.Send(msg)
	return true
}
