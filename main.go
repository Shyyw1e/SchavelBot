package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not loaded, using environment variables")
	}

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("BOT_TOKEN is empty")
	}

	appointmentContact := os.Getenv("CONTACT")
	channelLink := os.Getenv("TELEGRAM_CHANNEL")

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Bot authorized as @%s", bot.Self.UserName)

	updatesConfig := tgbotapi.NewUpdate(0)
	updatesConfig.Timeout = 60

	updates, err := bot.GetUpdatesChan(updatesConfig)
	if err != nil {
		log.Fatal(err)
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID

		if update.Message.IsCommand() && update.Message.Command() == "start" {
			sendWelcome(bot, chatID, channelLink)
			continue
		}

		switch update.Message.Text {
		case "Записаться на сеанс":
			sendAppointmentContact(bot, chatID, appointmentContact)
		case "Рассчитать стоимость":
			sendPriceList(bot, chatID)
		case "Выбрать свободные эскизы":
			sendSketches(bot, chatID)
		}
	}
}

func sendWelcome(bot *tgbotapi.BotAPI, chatID int64, channelLink string) {
	msg := tgbotapi.NewMessage(
		chatID,
		"Привет! Добро пожаловать в «Щавелевый суп».\n\n"+
			"Здесь ты можешь:\n"+
			"- записаться на сеанс\n"+
			"- рассчитать стоимость\n"+
			"- выбрать свободные эскизы\n\n"+
			"Чтобы начать, выбери нужный пункт в меню ниже.",
	)

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Записаться на сеанс"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Рассчитать стоимость"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Выбрать свободные эскизы"),
		),
	)
	keyboard.ResizeKeyboard = true
	msg.ReplyMarkup = keyboard

	if _, err := bot.Send(msg); err != nil {
		log.Println("failed to send welcome message:", err)
	}

	sendChannelButton(bot, chatID, channelLink)
}

func sendAppointmentContact(bot *tgbotapi.BotAPI, chatID int64, appointmentContact string) {
	msg := tgbotapi.NewMessage(
		chatID,
		"Для записи на сеанс напиши сюда: "+appointmentContact,
	)

	if _, err := bot.Send(msg); err != nil {
		log.Println("failed to send appointment contact:", err)
	}
}

func sendPriceList(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(
		chatID,
		"До 5 см (5к-10к)\n"+
			"До 20 см (10к-20к)\n"+
			"От 20 см и крупные проекты (20к-….)",
	)

	if _, err := bot.Send(msg); err != nil {
		log.Println("failed to send price list:", err)
	}
}

func sendSketches(bot *tgbotapi.BotAPI, chatID int64) {
	sketches := []string{
		"assets/sketch1.jpg",
		"assets/sketch2.jpg",
		"assets/sketch3.jpg",
	}

	if err := sendPhotoAlbum(bot, chatID, sketches, "Вот часть моих эскизов"); err != nil {
		log.Println("failed to send sketches album:", err)
	}
}

func sendPhotoAlbum(bot *tgbotapi.BotAPI, chatID int64, paths []string, caption string) error {
	if len(paths) < 2 || len(paths) > 10 {
		return fmt.Errorf("telegram media group requires 2-10 photos, got %d", len(paths))
	}

	bodyReader, bodyWriter := io.Pipe()
	multipartWriter := multipart.NewWriter(bodyWriter)
	errChan := make(chan error, 1)

	go func() {
		defer bodyWriter.Close()
		defer multipartWriter.Close()

		if err := multipartWriter.WriteField("chat_id", strconv.FormatInt(chatID, 10)); err != nil {
			errChan <- err
			return
		}

		media := make([]map[string]string, 0, len(paths))
		for i := range paths {
			fieldName := fmt.Sprintf("photo%d", i)
			item := map[string]string{
				"type":  "photo",
				"media": "attach://" + fieldName,
			}
			if i == 0 && caption != "" {
				item["caption"] = caption
			}
			media = append(media, item)
		}

		mediaJSON, err := json.Marshal(media)
		if err != nil {
			errChan <- err
			return
		}

		if err := multipartWriter.WriteField("media", string(mediaJSON)); err != nil {
			errChan <- err
			return
		}

		for i, path := range paths {
			fieldName := fmt.Sprintf("photo%d", i)

			file, err := os.Open(path)
			if err != nil {
				errChan <- err
				return
			}

			part, err := multipartWriter.CreateFormFile(fieldName, filepath.Base(path))
			if err != nil {
				file.Close()
				errChan <- err
				return
			}

			if _, err := io.Copy(part, file); err != nil {
				file.Close()
				errChan <- err
				return
			}

			if err := file.Close(); err != nil {
				errChan <- err
				return
			}
		}

		errChan <- nil
	}()

	endpoint := fmt.Sprintf(tgbotapi.APIEndpoint, bot.Token, "sendMediaGroup")
	req, err := http.NewRequest(http.MethodPost, endpoint, bodyReader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())

	resp, err := bot.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var apiResp tgbotapi.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return err
	}

	if err := <-errChan; err != nil {
		return err
	}

	if !apiResp.Ok {
		return fmt.Errorf("telegram api error: %s", apiResp.Description)
	}

	return nil
}

func sendChannelButton(bot *tgbotapi.BotAPI, chatID int64, channelLink string) {
	msg := tgbotapi.NewMessage(chatID, "Наш Telegram-канал:")

	button := tgbotapi.NewInlineKeyboardButtonURL(
		"Перейти в канал",
		channelLink,
	)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(button),
	)

	msg.ReplyMarkup = keyboard

	if _, err := bot.Send(msg); err != nil {
		log.Println("failed to send channel button:", err)
	}
}
