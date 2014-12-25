package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWsHandler(t *testing.T) {
	r, _ := http.NewRequest("GET", "/ws?username=K&password=123", nil)
	r.Header.Add("Connection", "keep-alive")
	r.Header.Add("Upgrade", "websocket")
	r.Header.Add("Connection", "Upgrade")
	r.Header.Add("Sec-WebSocket-Version", "13")
	r.Header.Add("Sec-WebSocket-Key", "IevqsuQWkrIYE0gVrLf1pg==")
	r.Header.Add("Sec-WebSocket-Extensions", "permessage-deflate; client_max_window_bits")
	r.Header.Add("Sec-WebSocket-Protocol", "myProtocol")
	w := httptest.NewRecorder()
	wsHandle := WsHandler()
	wsHandle.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("Home page didn't return %v", http.StatusOK)
	}
}
