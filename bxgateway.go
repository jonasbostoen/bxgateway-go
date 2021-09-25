package bxgateway

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Bxgateway struct {
	conn *websocket.Conn
}

func NewGateway(endpoint string, authKey string) (*Bxgateway, error) {
	c, _, err := websocket.DefaultDialer.Dial(endpoint, http.Header{"Authorization": []string{authKey}})

	if err != nil {
		return nil, err
	}

	return &Bxgateway{
		conn: c,
	}, nil
}

type Filter struct {
	str   string
	valid bool
}

func NewFilter() *Filter {
	return &Filter{
		str:   "",
		valid: true,
	}
}

func (f *Filter) To(address string) *Filter {
	f.str += fmt.Sprintf("({to} == '%s')", address)
	f.valid = true
	return f
}

func (f *Filter) From(address string) *Filter {
	f.str += fmt.Sprintf("({from} == '%s')", address)
	f.valid = true
	return f
}

func (f *Filter) And() *Filter {
	f.str += " AND "
	f.valid = false
	return f
}

func (f *Filter) Or() *Filter {
	f.str += " OR "
	f.valid = false
	return f
}

func (f *Filter) Build() (string, error) {
	if !f.valid {
		return "", errors.New("invalid filter")
	}
	return f.str, nil
}

type StreamConfig struct {
	Include []string `json:"include"`
	Filter  string   `json:"filters"`
}

type subRequest struct {
	ID     int           `json:"id"`
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
}

type txRequest struct {
	ID     int         `json:"id"`
	Method string      `json:"method"`
	Params interface{} `json:"params"`
}

type StreamTopic int

const (
	NewTxs StreamTopic = iota
	PendingTxs
	NewBlocks
)

func (s StreamTopic) String() string {
	return [...]string{"newTxs", "pendingTxs", "newBlocks"}[s]
}

func (gw *Bxgateway) Subscribe(topic StreamTopic, config StreamConfig) error {
	msg := subRequest{
		ID:     1,
		Method: "subscribe",
		Params: []interface{}{
			topic.String(),
			config,
		},
	}

	json, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	if err := gw.conn.WriteMessage(websocket.TextMessage, json); err != nil {
		return err
	}

	return nil
}

func (gw *Bxgateway) SendTransaction(signedTransaction string) error {
	msg := txRequest{
		ID:     1,
		Method: "blxr_tx",
		Params: struct {
			Transaction string `json:"transaction"`
		}{
			Transaction: signedTransaction,
		},
	}

	json, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	if err := gw.conn.WriteMessage(websocket.TextMessage, json); err != nil {
		return err
	}

	return nil
}

func (gw *Bxgateway) Close() {
	gw.conn.Close()
}

type Response struct {
	Method string `json:"method"`
	Params struct {
		Subscription string `json:"subscription"`
		Result       interface{}
	} `json:"params"`
	Error struct {
		Code    int64  `json:"code"`
		Message string `json:"message"`
		Data    string `json:"data"`
	} `json:"error"`
}

func (gw *Bxgateway) Start(c chan<- string) {
	go func() {
		var res Response
		for {
			_, message, err := gw.conn.ReadMessage()
			if err != nil {
				return
			}

			if err := json.Unmarshal(message, &res); err != nil {
				log.Fatal(err)
			}

			if res.Error.Message != "" {
				j, _ := json.Marshal(res.Error)
				c <- string(j)
				close(c)
			} else {
				j, _ := json.Marshal(res.Params.Result)
				c <- string(j)
			}

		}
	}()
}

func (gw *Bxgateway) StartNoParse(c chan<- []byte) {
	go func() {
		for {
			_, message, err := gw.conn.ReadMessage()
			if err != nil {
				return
			}
			c <- message
		}
	}()
}
