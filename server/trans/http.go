package trans

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"outspoken-goblin/utils"
)

const GptUrl = "https://api.openai.com/v1/chat/completions"

type Http struct {
	tr  *http.Transport
	gin *gin.Engine
}

func NewHttp() *Http {
	r := gin.Default()
	h := &Http{tr: utils.GetTr(), gin: r}
	h.bind()
	return h
}

func (h *Http) Run() {
	_ = h.gin.Run()
}

func (h *Http) bind() {
	h.gin.POST("/v1/chat/completions", func(c *gin.Context) {
		// 创建请求对象
		if req, err := http.NewRequest("POST", GptUrl, c.Request.Body); err != nil {
			c.String(200, "转发请求客户端错误：%s", err)
		} else {
			req.Header = c.Request.Header
			// 发送 POST 请求
			client := &http.Client{}
			if h.tr != nil {
				client.Transport = h.tr
			}
			if resp, err := (client).Do(req); err != nil {
				c.String(200, "转发错误：%s", err)
			} else {
				headers := c.Writer.Header()
				for k, v := range resp.Header {
					headers[k] = v
				}
				c.DataFromReader(200, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
			}
		}
		return
	})
}
