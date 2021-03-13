package allowance_socket

import (
	"errors"
	"net/http"
	"time"

	"github.com/fiatjaf/go-lnurl"
	"github.com/gorilla/websocket"
)

type Session struct {
	WS    *websocket.Conn
	ended chan bool
}

func (s *Session) Close() {
	s.ended <- true
	s.WS.Close()
}

func Connect(url string, msat int64, k1 string) (session *Session, err error) {
	dialer := websocket.Dialer{
		Proxy: http.ProxyFromEnvironment,
	}

	ws, _, err := dialer.Dial(url, nil)
	if err != nil {
		return
	}

	session = &Session{
		WS:    ws,
		ended: make(chan bool, 1),
	}

	// send first message
	err = ws.WriteJSON(lnurl.AllowanceRequest{k1, msat, "allowanceRequest"})
	if err != nil {
		return
	}

	// receive response
	ws.SetReadDeadline(time.Now().Add(10 * time.Second))
	var success lnurl.AllowanceSuccess
	err = ws.ReadJSON(&success)
	if err == nil {
		if success.Reason != "" {
			err = errors.New(success.Reason)
		}
		if success.Type != "allowanceSuccess" {
			err = errors.New("didn't get allowanceSuccess from service")
		}
	}
	if err != nil {
		return
	}

	go func() {
		defer ws.Close()
	}()

	return
}
