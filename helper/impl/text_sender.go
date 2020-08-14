package impl

import (
	"encoding/json"
	"fmt"
	"log"
	"myiotqq-plugin/enum"
	"myiotqq-plugin/global"
	"myiotqq-plugin/helper"
	iotqq "myiotqq-plugin/model"
)

type TextSender struct {
	helper.Sender
}

var textInstance helper.ISender = &TextSender{}


func (th *TextSender) makeBody(UserId int64,GroupID int64,SendTo int,Content string) []byte {
	//发送语音信息
	body := make(map[string]interface{})
	body["toUser"] = UserId
	body["sendToType"] = SendTo
	body["sendMsgType"] = enum.TEXT_MSG
	body["content"] = Content
	body["groupid"] = GroupID
	body["atUser"] = 0
	body["pwd"] = "mcoo"
	ret, _ := json.Marshal(body)
	return ret
}

func NewTextInstance() helper.ISender {
	return textInstance
}

func (th *TextSender) SendToUser(UserID int64, Content string,message iotqq.Message) {
	body := th.makeBody(UserID,0,enum.SEND_TO_FRIEND,Content)
	th.DoSend(global.BASE_URL,global.QQ,body)
}

func (th *TextSender) SendToGroup(GroupID int64, Content string,message iotqq.Message) {
	log.Print(fmt.Sprintf("发送: %s 到群: %d",Content,GroupID))
	body := th.makeBody(GroupID,0,enum.SEND_TO_GROUP,Content)
	th.DoSend(global.BASE_URL,global.QQ,body)
}
