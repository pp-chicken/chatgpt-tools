package gpt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	gogpt "github.com/sashabaranov/go-gpt3"
	"log"
	"net/http"
	"outspoken-goblin/utils"
	"strconv"
	"strings"
	"time"
)

const (
	Time1 = 5 * time.Minute
	Time2 = -2 * time.Hour
)

type Gpt struct {
	c       *gogpt.Client
	rdb     *redis.Client
	ctx     context.Context
	lockMap map[int64]chan bool
}

func NewGpt(redisAddr string, redisPass string, openAiOAuthToken string) *Gpt {
	if redisAddr == "" {
		redisAddr = "127.0.0.1:6379"
	}

	gptHttpClient := &http.Client{
		Timeout:   time.Minute,
		Transport: utils.GetTr(),
	}

	if openAiOAuthToken == "" {
		panic("open ai oauth token 不能为空")
	}
	gptConfig := gogpt.DefaultConfig(openAiOAuthToken)
	gptConfig.HTTPClient = gptHttpClient

	return &Gpt{
		c: gogpt.NewClientWithConfig(gptConfig),
		rdb: redis.NewClient(&redis.Options{
			Addr:     redisAddr,
			Password: redisPass,
			DB:       0,
		}),
		lockMap: make(map[int64]chan bool),
		ctx:     context.Background(),
	}
}

func (gpt *Gpt) Chat(sessionId int64, msg string, fn func(sessionId int64, respMessage string)) {
	redisCacheKey := fmt.Sprintf("gpt3-chat-session-%d", sessionId)
	gpt.getLock(sessionId) <- true
	defer func() {
		<-gpt.getLock(sessionId)
	}()
	now := time.Now()
	var messageCache []gogpt.ChatCompletionMessage
	if result, err := gpt.rdb.TTL(gpt.ctx, redisCacheKey).Result(); err == nil && result >= -1 {
		gpt.rdb.Expire(gpt.ctx, redisCacheKey, Time1)
	}
	if result, err := gpt.rdb.ZRangeByScore(gpt.ctx, redisCacheKey, &redis.ZRangeBy{
		Min: strconv.FormatInt(now.Add(Time2).UnixNano(), 10),
		Max: strconv.FormatInt(now.UnixNano(), 10),
	}).Result(); err == nil {
		for _, v := range result {
			newMessage := &gogpt.ChatCompletionMessage{}
			if err := json.NewDecoder(strings.NewReader(v)).Decode(newMessage); err != nil {
				continue
			}
			messageCache = append(messageCache, *newMessage)
		}
	}
	newMessage := gogpt.ChatCompletionMessage{
		Role:    "user",
		Content: msg,
	}
	messageCache = append(messageCache, newMessage)
	req := gogpt.ChatCompletionRequest{
		Model:            gogpt.GPT3Dot5Turbo,
		Messages:         messageCache,
		Temperature:      0.5, //MaxTokens:        4000,
		TopP:             0.3,
		FrequencyPenalty: 0.5,
		PresencePenalty:  0.0,
		Stream:           false,
		User:             strconv.FormatInt(sessionId, 10),
	}
	resp, err := gpt.c.CreateChatCompletion(gpt.ctx, req)
	if err != nil {
		log.Println("请求open ai 服务器异常", err)
		fn(sessionId, err.Error())
		return
	}
	jsonCacheBuffer := &bytes.Buffer{}
	if err := json.NewEncoder(jsonCacheBuffer).Encode(newMessage); err != nil {
		log.Println("缓存会话json序列化出错", err)
		fn(sessionId, "缓存会话json序列化出错"+err.Error())
		return
	}
	gpt.rdb.ZAdd(gpt.ctx, redisCacheKey, redis.Z{
		Score:  float64(now.UnixNano()),
		Member: jsonCacheBuffer.String(),
	})
	gpt.rdb.Expire(gpt.ctx, redisCacheKey, Time1)
	// Now that we know we've gotten a new message, we can construct a
	// reply! We'll take the Chat ID and Text from the incoming message
	// and use it to create a new message.
	for _, respMsg := range resp.Choices {
		newCacheBuffer := &bytes.Buffer{}
		if err := json.NewEncoder(newCacheBuffer).Encode(respMsg.Message); err != nil {
			log.Println("缓存会话json序列化出错", err)
			return
		}
		gpt.rdb.ZAdd(gpt.ctx, redisCacheKey, redis.Z{
			Score:  float64(time.Now().UnixNano()),
			Member: newCacheBuffer.String(),
		})
		fn(sessionId, respMsg.Message.Content)
	}
}

func (gpt *Gpt) ClearSession(sessionId int64) (bool, error) {
	redisCacheKey := fmt.Sprintf("gpt3-chat-session-%d", sessionId)
	_, err := gpt.rdb.Del(gpt.ctx, redisCacheKey).Result()
	return err == nil, err
}

func (gpt *Gpt) getLock(key int64) chan bool {
	if lock, ex := gpt.lockMap[key]; ex {
		return lock
	}
	lock := make(chan bool, 1)
	gpt.lockMap[key] = lock
	return lock
}
