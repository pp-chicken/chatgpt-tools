package server

import (
	"net/http"
	"net/url"
	"outspoken-goblin/server/trans"
	"testing"
)

func TestHttpServer(t *testing.T) {
	uri, _ := url.Parse("http://127.0.0.1:7890")
	h := trans.NewHttp(&http.Transport{
		Proxy: http.ProxyURL(uri),
	})
	h.Run()
	select {}
}
