package service

import (
	"chat-demo/cache"
	"chat-demo/conf"
	"chat-demo/pkg/e"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	// Gorilla的工作是转换原始HTTP连接进入一个有状态的websocket连接
	"github.com/gorilla/websocket"
)

const month = 60 * 60 * 24 * 30

// 定义发送消息的结构体
type SendMsg struct {
	Type    int    `json:"type"`
	Content string `json:"content"`
}

// 定义回复消息的结构体
type ReplyMsg struct {
	From    string `json:"from"`
	Code    int    `json:"code"`
	Content string `json:"content"`
}

// 定义用户结构体
type Client struct {
	ID     string
	SendID string
	Socket *websocket.Conn
	Send   chan []byte
}

// 定义广播结构体（广播内容和源用户）
type Broadcast struct {
	Client  *Client
	Message []byte
	Type    int
}

// 用户管理
type ClientManger struct {
	Clients    map[string]*Client
	Broadcast  chan *Broadcast
	Reply      chan *Client
	Register   chan *Client
	Unregister chan *Client
}

// 信息转JSON（发送者、接收者、内容）
type Message struct {
	Sender    string `json:"sender,omitempty"`
	Recipient string `json:"recipient,omitempty"`
	Content   string `json:"content,omitempty"`
}

var Manager = ClientManger{
	//参与连接的用户，设置最大连接数
	Clients:    make(map[string]*Client),
	Broadcast:  make(chan *Broadcast),
	Reply:      make(chan *Client),
	Register:   make(chan *Client),
	Unregister: make(chan *Client),
}

func CreateID(uid, toUid string) string {
	return uid + "->" + toUid //1->2
}

// 处理websocket逻辑
func Handler(c *gin.Context) {
	uid := c.Query("id")
	toUid := c.Query("toUid")
	// 升级websocket
	conn, err := (&websocket.Upgrader{
		// CheckOrigin是用于拦截或放行跨域请求。函数返回值为bool类型，即true放行，false拦截。
		// 如果请求不是跨域请求可以不赋值，这里是跨域请求并且为了方便直接返回true
		CheckOrigin: func(r *http.Request) bool {
			return true
		}}).Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		// NotFound 回复请求，并显示HTTP 404 未找到错误。
		http.NotFound(c.Writer, c.Request)
		return
	}
	//创建一个用户实例
	client := &Client{
		ID:     CreateID(uid, toUid), //1->2 发送方
		SendID: CreateID(toUid, uid), //2->1 接收方
		Socket: conn,
		Send:   make(chan []byte),
	}
	// 用户注册到用户管理上
	Manager.Register <- client
	go client.Read()
	go client.Write()
}

func (c *Client) Read() {
	defer func() {
		Manager.Unregister <- c
		_ = c.Socket.Close()
	}()
	for {
		// 发送方->接收方：ping；接收方->发送方：pong；
		c.Socket.PongHandler()
		SendMsg := new(SendMsg)
		// 传入字符串用c.Socket.ReadMessage()
		// c.Socket.ReadMessage()
		// 传入JSON格式用c.Socket.ReadJSON()
		err := c.Socket.ReadJSON(&SendMsg)
		if err != nil {
			fmt.Println("数据格式不正确", err)
			Manager.Unregister <- c
			_ = c.Socket.Close()
			break
		}
		if SendMsg.Type == 1 { //发送消息 1->2
			r1, _ := cache.RedisClient.Get(c.ID).Result()     //1->2
			r2, _ := cache.RedisClient.Get(c.SendID).Result() //2->1
			if r1 > "3" && r2 == "" {
				//1给2发消息，超过3条没有回复或2没看到，就停止1发送
				replyMsg := ReplyMsg{
					Code:    e.WebsocketLimit,
					Content: "达到限制",
				}
				msg, _ := json.Marshal(replyMsg) //序列化 data []byte
				_ = c.Socket.WriteMessage(websocket.TextMessage, msg)
				continue
			} else {
				// 将ID存储到redis里
				cache.RedisClient.Incr(c.ID)
				// 设置过期时间为3个月
				_, _ = cache.RedisClient.Expire(c.ID, time.Hour*24*30*30*3).Result()
			}
			Manager.Broadcast <- &Broadcast{
				Client:  c,
				Message: []byte(SendMsg.Content), //发送过来的消息
			}
		} else if SendMsg.Type == 2 {
			// 获取历史消息
			// strconv.Atoi()用于将字符串类型转换为int类型
			timeT, err := strconv.Atoi(SendMsg.Content) //string to int
			if err != nil {
				timeT = 999999
			}
			results, _ := FindMany(conf.MongoDBName, c.SendID, c.ID, int64(timeT), 10) //获取十条历史消息
			if len(results) > 10 {
				results = results[:10]
			} else if len(results) == 0 {
				replyMsg := ReplyMsg{
					Code:    e.WebsocketEnd,
					Content: "到底了",
				}
				msg, _ := json.Marshal(replyMsg)
				_ = c.Socket.WriteMessage(websocket.TextMessage, msg)
				continue
			}
			for _, result := range results {
				replyMsg := ReplyMsg{
					From:    result.From, //消息发送者
					Content: result.Msg,
				}
				msg, _ := json.Marshal(replyMsg)
				_ = c.Socket.WriteMessage(websocket.TextMessage, msg)
			}
		} else if SendMsg.Type == 3 {
			results, err := FirsFindtMsg(conf.MongoDBName, c.SendID, c.ID)
			if err != nil {
				log.Println(err)
			}
			for _, result := range results {
				replyMsg := ReplyMsg{
					From:    result.From,
					Content: fmt.Sprintf("%s", result.Msg),
				}
				msg, _ := json.Marshal(replyMsg)
				_ = c.Socket.WriteMessage(websocket.TextMessage, msg)
			}
		}
	}
}
func (c *Client) Write() {
	defer func() {
		_ = c.Socket.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				_ = c.Socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			replyMsg := ReplyMsg{
				Code:    e.WebsocketSuccessMessage,
				Content: fmt.Sprintf("%s", string(message)),
			}
			msg, _ := json.Marshal(replyMsg)
			_ = c.Socket.WriteMessage(websocket.TextMessage, msg)
		}
	}
}
