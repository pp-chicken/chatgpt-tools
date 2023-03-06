package server

import (
    "outspoken-goblin/server/trans"
    "testing"
)

func TestHttpServer(t *testing.T) {
    h := trans.NewHttp()
    h.Run()
    select {}
}
