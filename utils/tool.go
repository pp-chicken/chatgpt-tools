package utils

import (
	"log"
	"net/http"
	"net/url"
	"os"
)

func GetTr() (tr *http.Transport) {
	proxyUrlString := os.Getenv("PROXY_URL")
	if proxyUrlString != "" {
		log.Println("代理信息", proxyUrlString)
		uri, proxyError := url.Parse(proxyUrlString)
		if proxyError != nil {
			panic(proxyError)
		}
		tr = &http.Transport{
			// 设置代理
			Proxy: http.ProxyURL(uri),
		}
	}
	return
}
