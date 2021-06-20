package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/joho/godotenv"
	tb "gopkg.in/tucnak/telebot.v2"
)

var isDev = envPanic("GO_ENV") == "development"

func setupBot() (*tb.Bot, error) {
	var poller tb.Poller
	switch envPanic("GO_ENV") {
	case "development":
		poller = &tb.LongPoller{Timeout: 2}
	case "production":
		poller = &tb.LongPoller{Timeout: 30}
	}
	return tb.NewBot(tb.Settings{
		Token:  envPanic("BOT_TOKEN"),
		Poller: poller,
	})
}

func main() {
	if isDev {
		godotenv.Load()
	}
	bbClient := NewBbClient()
	bot, err := setupBot()
	if err != nil {
		log.Fatal(err)
	}

	bot.Handle("/start", func(m *tb.Message) {
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
			SwitchPMText:      "No response...",
			SwitchPMParameter: "nil",
		})
	}

	bot.Handle(tb.OnQuery, func(q *tb.Query) {
		queryString := strings.TrimSpace(q.Text)
		if queryString == "" {
			emptyAnswer(q)
			return
		}

		bbResponse, err := bbClient.Run(queryString)
		if err != nil {
			log.Println(err)
			emptyAnswer(q)
			return
		}

		if bbResponse.BadQuery > 0 || bbResponse.Error > 0 {
			bot.Answer(q, &tb.QueryResponse{
				SwitchPMText:      "Bad query or something",
				SwitchPMParameter: "nil",
			})
			return
		}

		articleResult := tb.ArticleResult{
			Title:       bbResponse.Query,
			Description: bbResponse.Text,
		}
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
			Results: tb.Results{&articleResult},
		})
		if err != nil {
			log.Println(err)
		}
	})

	if isDev {
		bot.Raw("deleteWebhook", map[string]string{
			"drop_pending_updates": "true",
		})
	}

	log.Print("Starting bot")
	bot.Start()
}
