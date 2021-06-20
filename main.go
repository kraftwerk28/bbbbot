package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/joho/godotenv"
	tb "gopkg.in/tucnak/telebot.v2"
)

func setupBot() (*tb.Bot, error) {
	var poller tb.Poller
	switch envPanic("GO_ENV") {
	case "development", "production":
		poller = &tb.LongPoller{}
		// case "production":
		// 	log.Fatal("Need to make webhook poller")
		// 	// poller = &tb.WebhookPoller{}
	}

	bot, err := tb.NewBot(tb.Settings{
		Token:  envPanic("BOT_TOKEN"),
		Poller: poller,
	})
	if err != nil {
		return nil, err
	}
	return bot, nil
}

func main() {
	godotenv.Load()
	bbClient := NewBbClient()
	bot, err := setupBot()
	if err != nil {
		log.Fatal(err)
	}

	bot.Handle(tb.OnText, func(m *tb.Message) {
		if m.Chat.Type != tb.ChatPrivate {
			return
		}
		sent, err := bot.Send(
			m.Sender,
			"<u>I guess I'm alive...</u>",
			&tb.SendOptions{ReplyTo: m, ParseMode: tb.ModeHTML},
		)
		if err != nil {
			log.Println(err)
			return
		}
		go func(m *tb.Message) {
			time.Sleep(time.Second * 5)
			bot.Delete(m) // Ignore error
		}(sent)
	})

	emptyAnswer := func(q *tb.Query) error {
		return bot.Answer(q, &tb.QueryResponse{
			CacheTime:         5,
			SwitchPMText:      "No response...",
			SwitchPMParameter: "nil",
		})
	}

	bot.Handle(tb.OnQuery, func(q *tb.Query) {
		queryString := strings.TrimSpace(q.Text)
		log.Printf("Got query \"%s\"\n", queryString)
		if queryString == "" {
			emptyAnswer(q)
			return
		}

		bbResponse, err := bbClient.Run(queryString)
		if err != nil {
			log.Println(err)
			err := emptyAnswer(q)
			if err != nil {
				log.Println(err)
			}
			return
		}

		articleResult := tb.ArticleResult{Title: bbResponse.Text}
		articleResult.SetResultID("nil")
		articleResult.SetContent(&tb.InputTextMessageContent{
			Text: fmt.Sprintf(
				"<b>%s</b>%s",
				escapeHtml(bbResponse.Query),
				escapeHtml(bbResponse.Text),
			),
			ParseMode: tb.ModeHTML,
		})

		err = bot.Answer(q, &tb.QueryResponse{
			Results:   tb.Results{&articleResult},
			CacheTime: 5,
		})
		if err != nil {
			log.Println(err)
		}
	})

	if envPanic("GO_ENV") == "development" {
		bot.Raw("deleteWebhook", map[string]string{
			"drop_pending_updates": "true",
		})
	}

	log.Print("Starting bot")
	bot.Start()
}
