package main

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gocolly/colly/v2"
	"github.com/spf13/viper"
	"golang.org/x/exp/rand"
	_ "modernc.org/sqlite"
)

func main() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Load config error: %v", err)
	}

	db, err := sql.Open("sqlite", "vnexpress.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTable := `
	CREATE TABLE IF NOT EXISTS articles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT,
		url TEXT UNIQUE,
		source TEXT,
		published_at TEXT,
		posted_at TEXT
	);
	`
	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatal(err)
	}

	for {
		execGetVNExpress(db)
		execArticles(db)

		sleepDuration := time.Duration(rand.Intn(2)+1) * time.Minute
		// sleepDuration := time.Second * 5
		fmt.Println("Next run in:", sleepDuration)

		time.Sleep(sleepDuration)
	}
}

func runCommonBotMessage(messageString string) {
	botToken := viper.GetString("botToken")
	chatID := int64(viper.GetInt("chatID"))

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	message := tgbotapi.NewMessage(chatID, messageString)
	_, err = bot.Send(message)
	if err != nil {
		log.Panic(err)
	}
}

func extractDate(rawText string) string {
	re := regexp.MustCompile(`(\d{1,2}/\d{1,2}/\d{4})`)
	match := re.FindString(rawText)

	if match != "" {
		parts := strings.Split(match, "/")
		if len(parts) == 3 {
			return fmt.Sprintf("%s-%s-%s", parts[2], parts[1], parts[0]) // YYYY-MM-DD
		}
	}

	return time.Now().Format("2006-01-02")
}

type Article struct {
	ID          int64
	Title       string
	URL         string
	Source      string
	PublishedAt string
	PostedAt    *string
}

func getArticles(db *sql.DB) ([]Article, error) {
	query := "SELECT id, title, url, source, published_at, posted_at FROM articles WHERE posted_at is NULL ORDER BY id DESC LIMIT 1"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []Article
	for rows.Next() {
		var article Article
		if err := rows.Scan(&article.ID, &article.Title, &article.URL, &article.Source, &article.PublishedAt, &article.PostedAt); err != nil {
			return nil, err
		}
		articles = append(articles, article)
	}

	return articles, nil
}

func updateArticle(db *sql.DB, id int64) error {
	query := `UPDATE articles SET posted_at = CURRENT_TIMESTAMP WHERE id = ?`
	result, err := db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		fmt.Printf("No record updated, URL not found: %d", id)
	}

	return nil
}

func deleteArticles(db *sql.DB) error {
	query := `UPDATE articles SET posted_at = CURRENT_TIMESTAMP WHERE source in ('codeaholicguy', 'toidicodedao')`
	result, err := db.Exec(query)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		fmt.Printf("No record updated, URL not found")
	}

	return nil
}

func execGetVNExpress(db *sql.DB) {
	c := colly.NewCollector()

	c.OnHTML("h3.title-news a", func(e *colly.HTMLElement) {
		title := e.Text
		link := e.Request.AbsoluteURL(e.Attr("href"))
		dateStr := extractDate(e.ChildText("span.date"))

		if title != "" && len(title) != 3 {
			_, err := db.Exec("INSERT OR IGNORE INTO articles (title, url, source, published_at) VALUES (?, ?, ?, ?)", title, link, "vnexpress", dateStr)

			if err != nil {
				log.Println("Error:", err)
			}
		}

	})

	c.OnScraped(func(r *colly.Response) {
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Println("Lỗi:", err)
	})

	links := []string{
		"https://vnexpress.net/",
		"https://vnexpress.net/suc-khoe",
		"https://vnexpress.net/du-lich",
		"https://vnexpress.net/oto-xe-may",
		"https://vnexpress.net/cong-nghe",
		"https://vnexpress.net/doi-song",
		"https://vnexpress.net/bat-dong-san",
		"https://vnexpress.net/khoa-hoc",
	}
	for _, link := range links {
		err := c.Visit(link)
		if err != nil {
			log.Fatal(err)
		}
	}

}

func execArticles(db *sql.DB) {
	arts, err := getArticles(db)

	if err != nil {
		log.Println("Error: ", err)
	}

	if len(arts) > 0 {
		art := arts[0]
		err := updateArticle(db, art.ID)
		if err != nil {
			log.Println("Error: ", err)
		}

		fmt.Println("-------------Data--------------")
		fmt.Println("A:", art.ID)
		fmt.Println("-------------End--------------")

		message := fmt.Sprintf("%s\n%s", art.Title, art.URL)
		runCommonBotMessage(message)
	}
}

func execGetCodeaholicguy(db *sql.DB) {
	c := colly.NewCollector()

	c.OnHTML("h1.entry-title a", func(e *colly.HTMLElement) {
		title := e.Text
		link := e.Request.AbsoluteURL(e.Attr("href"))
		dateStr := extractDate(e.ChildText("span.date"))

		if title != "" && len(title) != 3 {
			_, err := db.Exec("INSERT OR IGNORE INTO articles (title, url, source, published_at) VALUES (?, ?, ?, ?)", title, link, "codeaholicguy", dateStr)

			if err != nil {
				log.Println("Error:", err)
			}
		}

	})

	c.OnScraped(func(r *colly.Response) {
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Println("Lỗi:", err)
	})

	err := c.Visit("https://codeaholicguy.com/")
	if err != nil {
		log.Fatal(err)
	}

}

func execGetToidicodedao(db *sql.DB) {
	c := colly.NewCollector()

	c.OnHTML("h1.entry-title a", func(e *colly.HTMLElement) {
		title := e.Text
		link := e.Request.AbsoluteURL(e.Attr("href"))
		dateStr := extractDate(e.ChildText("span.date"))

		if title != "" && len(title) != 3 {
			_, err := db.Exec("INSERT OR IGNORE INTO articles (title, url, source, published_at) VALUES (?, ?, ?, ?)", title, link, "toidicodedao", dateStr)

			if err != nil {
				log.Println("Error:", err)
			}
		}

	})

	c.OnScraped(func(r *colly.Response) {
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Println("Lỗi:", err)
	})

	err := c.Visit("https://toidicodedao.com/")
	if err != nil {
		log.Fatal(err)
	}
}

func execVOZF17(db *sql.DB) {
	c := colly.NewCollector()

	c.OnHTML("div.structItem-title a", func(e *colly.HTMLElement) {
		title := e.Text
		link := e.Request.AbsoluteURL(e.Attr("href"))
		dateStr := extractDate(e.ChildText("span.date"))

		if title != "" && len(title) != 3 {
			_, err := db.Exec("INSERT OR IGNORE INTO articles (title, url, source, published_at) VALUES (?, ?, ?, ?)", title, link, "voz", dateStr)

			if err != nil {
				log.Println("Error:", err)
			}
		}

	})

	c.OnScraped(func(r *colly.Response) {
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Println("Error:", err)
	})

	err := c.Visit("https://voz.vn/f/chuyen-tro-linh-tinh%E2%84%A2.17/")
	if err != nil {
		log.Fatal(err)
	}
}
