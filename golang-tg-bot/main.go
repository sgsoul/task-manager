package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/ClickHouse/clickhouse-go"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {
	// Настройки для подключения к ClickHouse
	connect, err := sql.Open("clickhouse", "tcp://clickhouse:9000?username=default&password=")
	if err != nil {
		log.Fatalf("Ошибка при открытии соединения с ClickHouse: %v", err)
	}
	defer connect.Close()

	// Создание базы данных "todo", если она не существует
	if _, err := connect.Exec("CREATE DATABASE IF NOT EXISTS todo"); err != nil {
		log.Fatalf("Ошибка при создании базы данных: %v", err)
	}

	// Создание таблицы "todo", если она не существует
	if _, err := connect.Exec(`
		CREATE TABLE IF NOT EXISTS todo.todo (
			id UUID DEFAULT generateUUIDv4(),
			text String,
			status String DEFAULT 'active'
		) engine = MergeTree() order by id
	`); err != nil {
		log.Fatalf("Ошибка при создании таблицы: %v", err)
	}

	APP_TOKEN, _ := os.LookupEnv("APP_TOKEN")

	// Инициализация бота Telegram
	bot, err := tgbotapi.NewBotAPI(APP_TOKEN)
	if err != nil {
		log.Fatalf("Ошибка при инициализации бота: %v", err)
	}

	log.Printf("Authorized!")

	// Инициализация обработчика команд бота
	bot.Debug = true

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatalf("Ошибка при получении обновлений: %v", err)
	}

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		switch update.Message.Command() {
		case "all":
			// Получаем список всех задач
			rows, err := connect.Query("SELECT text, status FROM todo.todo")
			if err != nil {
				log.Printf("Ошибка при получении списка задач: %v", err)
				continue
			}
			defer rows.Close()

			var activeTasks, completedTasks string

			// Проходим по каждой задаче и добавляем ее в соответствующую группу
			for rows.Next() {
				var text, status string
				if err := rows.Scan(&text, &status); err != nil {
					log.Printf("Ошибка при сканировании результата: %v", err)
					continue
				}

				if status == "active" {
					activeTasks += fmt.Sprintf("- %s\n", text)
				} else {
					completedTasks += fmt.Sprintf("- %s\n", text)
				}
			}

			// Формируем сообщение с задачами
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Нужно сделать:\n%s\n\nВыполненные задачи:\n%s", activeTasks, completedTasks))
			_, err = bot.Send(msg)
			if err != nil {
				log.Printf("Ошибка при отправке сообщения: %v", err)
			}

		case "add":
			text := update.Message.CommandArguments()

			// Проверяем, что текст задачи не пустой
			if text == "" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Пожалуйста, укажите текст задачи.")
				_, err = bot.Send(msg)
				if err != nil {
					log.Printf("Ошибка при отправке сообщения: %v", err)
				}
				return
			}

			// Добавляем задачу в базу данных с указанием статуса "active"
			tx, err := connect.Begin()
			if err != nil {
				log.Printf("Ошибка при начале транзакции: %v", err)
				return
			}
			defer tx.Rollback() // Откатываем транзакцию в случае ошибки

			_, err = tx.Exec("INSERT INTO todo.todo (text) VALUES (?)", text)
			if err != nil {
				log.Printf("Ошибка при добавлении задачи: %v", err)
				return
			}

			err = tx.Commit() // Фиксируем изменения
			if err != nil {
				log.Printf("Ошибка при фиксации транзакции: %v", err)
				return
			}

			log.Printf("Добавлена задача: %s", text)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Добавлена задача: %s", text))
			_, err = bot.Send(msg)
			if err != nil {
				log.Printf("Ошибка при отправке сообщения: %v", err)
			}

		case "done":
			text := update.Message.CommandArguments()
			_, err := connect.Exec("ALTER TABLE todo.todo UPDATE status = 'complete' WHERE text = ?", text)
			if err != nil {
				log.Printf("Ошибка при выполнении задачи: %v", err)
				continue
			}
			log.Printf("Выполнена задача: %s", text)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Выполнена задача: %s", text))
			_, err = bot.Send(msg)
			if err != nil {
				log.Printf("Ошибка при отправке сообщения: %v", err)
			}
		}
	}
}
