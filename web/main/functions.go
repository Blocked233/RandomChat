package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type socketAndchan struct {
	conn *websocket.Conn
	ch   chan *socketAndchan
	stop chan struct{}
}

var shortwaitusers = make(chan *socketAndchan, 1000)
var midwaitusers = make(chan *socketAndchan, 1000)
var longwaitusers = make(chan *socketAndchan, 1000)
var upgrader = websocket.Upgrader{} //default
const fail = "stop"

func transMsg(local, peer *socketAndchan) {

	local.conn.WriteMessage(1, []byte("！发送消息吧！"))
	startTime := time.Now().Unix()
	for {

		//服务器收到发送端的消息 web->server
		messageType, message, err := local.conn.ReadMessage()
		if err != nil {
			local.conn.Close()
			peer.conn.WriteMessage(1, []byte(fail))
			break //主进程结束
		}

		if string(message) == "stop" {
			peer.conn.WriteMessage(messageType, message)
			stopTime := time.Now().Unix()
			spendTime := stopTime - startTime

			if isStart(local.conn) {
				if spendTime < 5 {
					randomChat(local, longwaitusers)
				} else if spendTime < 30 {
					randomChat(local, midwaitusers)
				} else {
					randomChat(local, shortwaitusers)
				}
			}
		}

		if string(message) == "stopack" {

			if isStart(local.conn) {

				randomChat(local, shortwaitusers)
			}
		}

		//服务器发送给接收端的消息 server->web
		if err := peer.conn.WriteMessage(messageType, message); err != nil {
			peer.conn.Close()
			local.conn.WriteMessage(1, []byte(fail))
			if isStart(local.conn) {
				randomChat(local, shortwaitusers)
			}
		}

	}
}

func randomChat(local *socketAndchan, users chan *socketAndchan) {

	//开始匹配
	select {
	case peer := <-users:
		log.Println("catch queue")
		peer.ch <- local

		transMsg(local, peer)
	default:
		users <- local
		log.Println("queue")
		peer := <-local.ch

		transMsg(local, peer)
	} 

}

func socketHandler(w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Error during upgradation:", err)
		return
	}

	if isStart(conn) {
		ch := make(chan *socketAndchan)
		stop := make(chan struct{}, 1)
		local := socketAndchan{
			conn,
			ch,
			stop,
		}

		randomChat(&local, shortwaitusers)

	}

}

func isStart(conn *websocket.Conn) bool {
	if _, start, err := conn.ReadMessage(); err != nil {
		return false
	} else {
		if string(start) == "start" {
			return true
		} else {
			return false
		}
	}
}
