package main

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/spf13/viper"
)

func main() {
	// Cấu hình Viper để đọc file config.yaml
	viper.SetConfigName("config") // Tên file không bao gồm phần mở rộng .yaml
	viper.SetConfigType("yaml")   // Kiểu file
	viper.AddConfigPath(".")      // Đường dẫn chứa file (thư mục hiện tại)

	// Đọc file config.yaml
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Lỗi khi đọc config: %v", err)
	}

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

	log.Println("Message sent successfully!")
}
