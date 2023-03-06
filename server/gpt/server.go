package gpt

import (
	tgBotApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"outspoken-goblin/service/gpt"
	"outspoken-goblin/service/tg"
)

var (
	redisAddr string
	redisPass string
	//gptSystemSetting []string = []string{"你是一个记忆只有5分钟的机器人", "如果连续对话超过2小时就会忘记之前的一些对话"}
)

func init() {
	redisAddr = os.Getenv("REDIS_ADDR")
}

func Run() {
	log.Println("开始运行 GPT 服务")
	bot := tg.NewBot(os.Getenv("TELEGRAM_APITOKEN"), false)
	gpt3 := gpt.NewGpt(redisAddr, redisPass, os.Getenv("OPENAI_OAUTH_TOKEN"))
	bot.SetCommand(tgBotApi.BotCommand{
		Command:     "help",
		Description: "说明",
	}, tgBotApi.BotCommand{
		Command:     "clear",
		Description: "清除对话记录",
	})

	bot.Listen(func(update tgBotApi.Update) {
		if update.Message == nil {
			return
		}

		if update.Message.IsCommand() {
			log.Println("这是一个命令", update.Message.Command())
			switch update.Message.Command() {
			case "clear":
				if ok, err := gpt3.ClearSession(update.Message.Chat.ID); ok {
					bot.MessageResp(update, "清除对话记录成功")
				} else {
					log.Println("清除对话记录失败", err)
					if err != nil {
						bot.MessageResp(update, "清除对话记录失败: "+err.Error())
						return
					}
					bot.MessageResp(update, "清除对话记录失败: <nil>")
				}
			case "help":
				bot.MessageResp(update, "这是一个机器人，你可以和他聊天，他会记住你的对话，如果你连续对话超过2小时，他会忘记之前的对话，如果你超过5分钟没有对话，他会忘记全部会话，如果你想清除对话记录，可以发送 /clear 命令")
			}
			return
		}

		if update.Message.Text != "" {
			if update.Message.Chat.IsGroup() || update.Message.Chat.IsSuperGroup() {
				if update.Message.Entities != nil {
					for _, v := range update.Message.Entities {
						switch v.Type {
						case "mention":
							if bot.GetBot().IsMessageToMe(*update.Message) {
								gpt3.Chat(update.Message.Chat.ID, update.Message.Text, func(sessionId int64, respMessage string) {
									bot.MessageResp(update, respMessage)
								})
								return
							}
						case "text_mention":
							if v.User.ID == bot.GetBot().Self.ID {
								gpt3.Chat(update.Message.Chat.ID, update.Message.Text, func(sessionId int64, respMessage string) {
									bot.MessageResp(update, respMessage)
								})
								return
							}
						}
					}
				}
				return
			}
			gpt3.Chat(update.Message.Chat.ID, update.Message.Text, func(sessionId int64, respMessage string) {
				bot.MessageResp(update, respMessage)
			})
		}
	})
}
