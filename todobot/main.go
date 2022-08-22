package main

import (
	"database/sql"
	"log"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/mattn/go-sqlite3"
)

// Статусы, обозначающие поведение бота
const (
	WAITING_FOR_COMMAND int = 0
	WAITING_FOR_TASK    int = 1
	WAITING_FOR_NUMBER  int = 2
)

func main() {

	bot_status := WAITING_FOR_COMMAND //Статус - Бот ожидает команды

	db, err := sql.Open("sqlite3", "todo.db")
	if err != nil {
		log.Panic(err)
	}

	//env := os.Getenv("BOT_KEY")
	//print(env)
	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_KEY")) //Назначаем боту токен
	if err != nil {
		log.Panic(err)
	}
	log.Printf("Hello, Bot %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0) //Взял с туториала, полагаю, что мы здесь задаем, как часто бот должен проверять сообщения
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message != nil { //Если сообщение существует
			switch bot_status { //Проверяем статусы бота и задаем новые
			case WAITING_FOR_COMMAND: //
				bot_message := tgbotapi.NewMessage(update.Message.Chat.ID, "")
				switch update.Message.Command() { //Проверяем, какую команду отправил пользователь

				case "add":
					bot_message.Text = "Что вам нужно сделать?"
					bot_status = WAITING_FOR_TASK
					bot.Send(bot_message)

				case "list":
					bot_message.Text = "Ваш список дел:"
					command, err := db.Prepare("SELECT * FROM TASK WHERE USERNAME=?")

					if err != nil {
						log.Panic(err)
					}
					rows, err := command.Query(update.Message.From.UserName)
					var id int
					var username string
					var task string

					bot.Send(bot_message)
					for rows.Next() {
						rows.Scan(&id, &username, &task)
						log.Print()
						task := tgbotapi.NewMessage(update.Message.Chat.ID, "Номер: \n"+strconv.Itoa(id)+"\nЗадача:\n"+task)
						bot.Send(task)
					}
					bot_message.Text = ""
					bot.Send(bot_message)

				case "complete":
					bot_message.Text = "Введите номер выполненной задачи"
					bot_status = WAITING_FOR_NUMBER
					bot.Send(bot_message)
				default:
					bot_message.Text = "Я не знаю такой команды"
					bot.Send(bot_message)
				}

			case WAITING_FOR_TASK:

				log.Printf("Это дело хочет добавить пользователь %s", update.Message.Text)
				command, err := db.Prepare("INSERT INTO TASK(USERNAME, TEXT) values(?,?)")

				if err != nil {
					log.Panic(err)
				}

				command.Exec(update.Message.From.UserName, update.Message.Text)
				bot_message := tgbotapi.NewMessage(update.Message.Chat.ID, "Добавление выполнено успешно")
				bot_message.ReplyToMessageID = update.Message.MessageID
				bot.Send(bot_message)
				bot_status = WAITING_FOR_COMMAND

			case WAITING_FOR_NUMBER: //Ждем номер той задачи, которую хочет завершить/удалить пользователь
				log.Printf("Задача под этим номером удаляется пользователем %d", update.Message.Text)
				number, err := strconv.Atoi(update.Message.Text)
				if err != nil {
					log.Panic(err)
				}
				command, err := db.Prepare("DELETE FROM TASK where ID=?")
				command.Exec(number)
				bot_message := tgbotapi.NewMessage(update.Message.Chat.ID, "Удаление выполнено успешно")
				bot_message.ReplyToMessageID = update.Message.MessageID
				bot_status = WAITING_FOR_COMMAND
				bot.Send(bot_message)
				bot_status = WAITING_FOR_COMMAND

			}

		}

	}

}
