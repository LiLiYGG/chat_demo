package service

import (
	"chat-demo/conf"
	"chat-demo/pkg/e"
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

func (manager *ClientManger) Start() {
	for {
		fmt.Println("----监听管道通信----")
		// select语句 只能用于channel信道的IO操作，每个case都必须是一个信道。
		select {
		case conn := <-Manager.Register: //建立连接
			fmt.Printf("有新连接:%v", conn.ID)
			Manager.Clients[conn.ID] = conn //把连接放到用户管理上
			replyMsg := &ReplyMsg{
				Code:    e.WebsocketSuccess,
				Content: "已经连接到服务器了",
			}
			msg, _ := json.Marshal(replyMsg)
			_ = conn.Socket.WriteMessage(websocket.TextMessage, msg)
		// 断开连接
		case conn := <-Manager.Unregister:
			fmt.Printf("连接失败%s", conn.ID)
			if _, ok := Manager.Clients[conn.ID]; ok {
				replyMsg := &ReplyMsg{
					Code:    e.WebsocketEnd,
					Content: "连接中断",
				}
				msg, _ := json.Marshal(replyMsg)
				_ = conn.Socket.WriteMessage(websocket.TextMessage, msg)
				close(conn.Send)
				delete(Manager.Clients, conn.ID)
			}
		//广播信息
		case broadcast := <-Manager.Broadcast:
			message := broadcast.Message
			sendId := broadcast.Client.ID
			flag := false //默认对方是离线状态
			for id, conn := range Manager.Clients {
				// id不在redis缓存里
				if id != sendId {
					continue
				}
				select {
				case conn.Send <- message:
					flag = true
				default:
					close(conn.Send)
					delete(Manager.Clients, conn.ID)
				}
			}
			id := broadcast.Client.ID
			if flag {
				replyMsg := &ReplyMsg{
					Code:    e.WebsocketOnlineReply,
					Content: "对方在线应答",
				}
				msg, _ := json.Marshal(replyMsg)
				_ = broadcast.Client.Socket.WriteMessage(websocket.TextMessage, msg)
				err := InsertMsg(conf.MongoDBName, id, string(message), 1, int64(3*month)) //1为默认在线已读
				if err != nil {
					fmt.Println("InsertOne Err", err)
				}
			} else {
				log.Println("对方不在线")
				replyMsg := ReplyMsg{
					Code:    e.WebsocketOfflineReply,
					Content: "对方不在线应答",
				}
				msg, err := json.Marshal(replyMsg)
				_ = broadcast.Client.Socket.WriteMessage(websocket.TextMessage, msg)
				err = InsertMsg(conf.MongoDBName, id, string(message), 0, int64(3*month))
				if err != nil {
					fmt.Println("InsertOneMsg Err", err)
				}
			}
		}
	}
}
