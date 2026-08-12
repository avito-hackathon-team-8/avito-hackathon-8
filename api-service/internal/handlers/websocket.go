package handlers

import (
	"time"

	"github.com/gorilla/websocket"
)

const (
	websocketWriteWait  = 10 * time.Second
	websocketPongWait   = 60 * time.Second
	websocketPingPeriod = (websocketPongWait * 9) / 10
	websocketMaxMessage = 4 << 10
)

func configureWebSocket(connection *websocket.Conn) {
	connection.SetReadLimit(websocketMaxMessage)
	_ = connection.SetReadDeadline(time.Now().Add(websocketPongWait))
	connection.SetPongHandler(func(string) error {
		return connection.SetReadDeadline(time.Now().Add(websocketPongWait))
	})
}

func writeWebSocketJSON(connection *websocket.Conn, value any) error {
	if err := connection.SetWriteDeadline(time.Now().Add(websocketWriteWait)); err != nil {
		return err
	}

	return connection.WriteJSON(value)
}

func writeWebSocketPing(connection *websocket.Conn) error {
	return connection.WriteControl(
		websocket.PingMessage,
		nil,
		time.Now().Add(websocketWriteWait),
	)
}

func closeWebSocket(connection *websocket.Conn) {
	_ = connection.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		time.Now().Add(websocketWriteWait),
	)
	_ = connection.Close()
}
