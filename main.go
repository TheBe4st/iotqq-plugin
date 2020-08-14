package main

import (
	"log"
	"myiotqq-plugin/global"
	iotqq "myiotqq-plugin/model"
	"myiotqq-plugin/strategy"
	"myiotqq-plugin/task"
	"runtime"
	"strconv"
	"time"

	gosocketio "github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
)

func init() {
}

func SendJoin(c *gosocketio.Client) {
	log.Println("获取QQ号连接")
	result, err := c.Ack("GetWebConn", global.QQ, time.Second*5)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("emit", result)
	}
}
func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	c, err := gosocketio.Dial(
		gosocketio.GetUrl(global.HOST, global.PORT, false),
		transport.GetDefaultWebsocketTransport())
	if err != nil {
		log.Fatal(err)
	}

	err = c.On("OnGroupMsgs", func(h *gosocketio.Channel, args iotqq.Message) {
		defer func() {
			if err := recover(); err != nil {
				log.Fatal(err)
			}
		}()
		var mess = args.CurrentPacket.Data
		//不处理自己的消息
		if strconv.Itoa(int(mess.FromUserID)) == global.QQ {
			return
		}

		log.Println("群聊消息: ", mess.FromNickName+"<"+strconv.FormatInt(mess.FromUserID, 10)+">: "+mess.Content)
		go strategy.SelectFactory(mess.Content)(args)
	})
	if err != nil {
		log.Fatal(err)
	}
	c.On(gosocketio.OnDisconnection, func(h *gosocketio.Channel) {
		log.Fatal("Disconnected")
	})
	err = c.On(gosocketio.OnConnection, func(h *gosocketio.Channel) {
		log.Println("连接成功")
	})
	if err != nil {
		log.Fatal(err)
	}
	task.Start()
home:
	SendJoin(c)
	time.Sleep(600 * time.Second)
	goto home
}
