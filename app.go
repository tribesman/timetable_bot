package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"./models"
	tgBot "github.com/go-telegram-bot-api/telegram-bot-api"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// App Переменные приложения
type App struct {
	Cfg           Config            `json:"cfg"`
	Db            *gorm.DB          `json:"db"`
	Step          map[int]string    `json:"step"`
	TimePadID     map[int]int       `json:"time_pad_id"`
	EventID       map[int]int       `json:"event_id"`
	Query         map[int]string    `json:"query"`
	User          models.User       `json:"user"`
	CallaBackData map[string]string `json:"callback_data"`
	Bot           tgBot.BotAPI
}

// Config Настройки
type Config struct {
	TelegramApiKey string `json:"telegramApiKey"`
	BotDebug       bool   `json:"bot_debug"`
	Root           string `json:"root"`
	DbUser         string `json:"db_user"`
	DbPassword     string `json:"db_password"`
	DbDSN          string `json:"db_dsn"`
	BotUrl         string `json:"bot_url"`
}

// Start Заводим проект
func Start(app *App) {
	app.Cfg = LoadConfig()
	app.Db = Connect(&app.Cfg)
	Migrate(app)
	SetCallBackData(app)
}

// Connect Подключение к БД
func Connect(Cfg *Config) *gorm.DB {
	dsn := Cfg.DbUser + ":" + Cfg.DbPassword + "@" + Cfg.DbDSN
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	return db
}

// LoadConfig загрузка конфига
func LoadConfig() Config {

	fmt.Printf(
		"Start by \t\t: %s/%s\n",
		runtime.GOOS,
		runtime.GOARCH,
	)

	// Текущаяя директория
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	root := filepath.Dir(ex)
	fmt.Println("Current dir\t\t:", root)

	// Загрузка конфига
	var cfg = Config{}
	file, _ := ioutil.ReadFile(root + "/config.json")
	err = json.Unmarshal(file, &cfg) //([]byte(file), &cfg)
	if err != nil {
		cfg.TelegramApiKey = "BotKey"
		cfg.BotDebug = true
		cfg.BotUrl = "https://t.me/BotName"

		//[protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
		cfg.DbDSN = "unix(/Applications/MAMP/tmp/mysql/mysql.sock)/DB_NAME?charset=utf8mb4"
		cfg.DbUser = "root"
		cfg.DbPassword = "root"

		_cfg, _ := json.MarshalIndent(cfg, "", " ")
		_ = ioutil.WriteFile(root+"/config.json", _cfg, 0644)
	}
	cfg.Root = root

	jsonData, _ := json.Marshal(cfg)
	fmt.Println("Current settings\t:", string(jsonData))

	return cfg
}

func SetCallBackData(app *App) {
	app.CallaBackData["index"] = createCallbackDataJson(&CallbackQueryData{Action: "tt.index"})
	app.CallaBackData["tt.add"] = createCallbackDataJson(&CallbackQueryData{Action: "tt.add"})

	/*
		var call back QueryData CallbackQueryData
		var callbackQueryDataJson []byte

		callbackQueryData = CallbackQueryData{Action: "tt.index"}
		callbackQueryDataJson, _ = json.Marshal(callbackQueryData)
		app.CallaBackData["index"] = string(callbackQueryDataJson)


		callbackQueryData = CallbackQueryData{Action: "tt.add"}
		callbackQueryDataJson, _ = json.Marshal(callbackQueryData)
		app.CallaBackData["tt.add"] = string(callbackQueryDataJson)*/

}

func createCallbackDataJson(callbackQueryData *CallbackQueryData) string {
	var callbackQueryDataJson []byte
	callbackQueryDataJson, _ = json.Marshal(callbackQueryData)
	s := string(callbackQueryDataJson)
	if len(s) > 62 {
		log.Panicf("CallbackData is too long\n[%d] %s", len(s), s)
	}
	return s
}

func ChunkInlineKeyboardButton(items []tgBot.InlineKeyboardButton, chunkSize int) (chunks [][]tgBot.InlineKeyboardButton) {
	for chunkSize < len(items) {
		items, chunks = items[chunkSize:], append(chunks, items[0:chunkSize:chunkSize])
	}
	return append(chunks, items)
}

func GetCallbackQueryData(data string) CallbackQueryData {
	var _json CallbackQueryData
	_ = json.Unmarshal([]byte(data), &_json)
	return _json
}

func Panic(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func NoPanic(err error) {
	if err != nil {
		log.Println(err)
	}
}
