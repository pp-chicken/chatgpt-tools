package main

import (
    "log"
    "os"
    "outspoken-goblin/server/gpt"
    "outspoken-goblin/server/trans"
)

func main() {
    startStatus := false
    if os.Getenv("GPT") == "true" {
        go gpt.Run()
        startStatus = true
    }
    if os.Getenv("GPT_PROXY_SERVER") == "true" {
        trans.NewHttp().Run()
        startStatus = true
    }
    if startStatus {
        select {}
    } else {
        log.Println("没有启动任何服务")
    }
}
