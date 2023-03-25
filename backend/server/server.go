package server

import (
	"net/http"
)

func HandleConnections(w http.ResponseWriter, r *http.Request) {
	// TODO: 实现直播系统服务逻辑
	w.Write([]byte("Live streaming server is under construction."))
}
