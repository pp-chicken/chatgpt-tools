package tg

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"html"
	"log"
	"net/http"
	"outspoken-goblin/utils"
)

type Bot struct {
	bot *tgbotapi.BotAPI
}

func NewBot(token string, debug bool) *Bot {
	botHttpClient := &http.Client{
		Transport: utils.GetTr(),
	}

	bot, bootErr := tgbotapi.NewBotAPIWithClient(token, tgbotapi.APIEndpoint, botHttpClient)
	if bootErr != nil {
		panic(bootErr)
	}

	bot.Debug = debug
	log.Printf("Authorized on account %s", bot.Self.UserName)

	return &Bot{bot: bot}
}

func (b *Bot) GetBot() *tgbotapi.BotAPI {
	return b.bot
}

func (b *Bot) Listen(fn func(update tgbotapi.Update)) {
	// Create a new UpdateConfig struct with an offset of 0. Offsets are used
	// to make sure Telegram knows we've handled previous values and we don't
	// need them repeated.
	updateConfig := tgbotapi.NewUpdate(0)

	// Tell Telegram we should wait up to 30 seconds on each request for an
	// update. This way we can get information just as quickly as making many
	// frequent requests without having to send nearly as many.
	updateConfig.Timeout = 30

	// Start polling Telegram for updates.
	updates := b.bot.GetUpdatesChan(updateConfig)
	// Let's go through each update that we're getting from Telegram.
	for update := range updates {
		go fn(update)
	}
}

func (b *Bot) MessageResp(update tgbotapi.Update, respMsg string) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, respMsg)
	msg.ParseMode = tgbotapi.ModeMarkdown
	// We'll also say that this message is a reply to the previous message.
	// For any other specifications than Chat ID or Text, you'll need to
	// set fields on the `MessageConfig`.
	msg.ReplyToMessageID = update.Message.MessageID

	// Okay, we're sending our message off! We don't care about the message
	// we just sent, so we'll discard it.
	if _, err := b.bot.Send(msg); err != nil {
		// Note that panics are a bad way to handle errors. Telegram can
		// have service outages or network errors, you should retry sending
		// messages or more gracefully handle failures.
		log.Println("响应出错信息", msg.Text)
		log.Println("响应数据<ModeMarkdown>出错", err)
		msg.ParseMode = tgbotapi.ModeMarkdownV2
		if _, err = b.bot.Send(msg); err != nil {
			log.Println("响应数据<ModeMarkdownV2>出错", err)
			msg.ParseMode = tgbotapi.ModeHTML
			if _, err = b.bot.Send(msg); err != nil {
				log.Println("响应数据<html>出错", err)
				msg.ParseMode = tgbotapi.ModeHTML
				msg.Text = html.EscapeString(msg.Text)
				if _, err = b.bot.Send(msg); err != nil {
					log.Println("响应数据<ModeHTML EscapeString>出错", err)
				}
			}
		}
	}
}

func (b *Bot) SetCommand(cmd ...tgbotapi.BotCommand) {
	command := tgbotapi.NewSetMyCommands(cmd...)
	if _, err := b.bot.Send(command); err != nil {
		log.Println("指令设置失败", err)
	}
}
