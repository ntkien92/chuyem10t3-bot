package main

import (
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/spf13/viper"
	"golang.org/x/exp/rand"
)

func main() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Load config error: %v", err)
	}

	for {
		runCommonBotMessage()

		sleepDuration := time.Duration(rand.Intn(15)+1) * time.Minute
		fmt.Println("Next run in:", sleepDuration)

		time.Sleep(sleepDuration)
	}
}

func runCommonBotMessage() {
	botToken := viper.GetString("botToken")
	chatID := int64(viper.GetInt("chatID"))

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	message := tgbotapi.NewMessage(chatID, "Chém đuê ae!!")
	_, err = bot.Send(message)
	if err != nil {
		log.Panic(err)
	}
}
