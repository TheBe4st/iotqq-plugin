package impl

import (
	"encoding/json"
	"myiotqq-plugin/enum"
	"myiotqq-plugin/global"
	"myiotqq-plugin/helper"
	iotqq "myiotqq-plugin/model"
	"net/url"
)

type VoiceSender struct {
	helper.Sender
}

var voiceInstance helper.ISender = &VoiceSender{}

func (th *VoiceSender) makeBody(UserId int64,GroupID int64,SendTo int,Content string) []byte {
	//发送语音信息
	body := make(map[string]interface{})
	body["toUser"] = UserId
	body["sendToType"] = SendTo
	body["sendMsgType"] = enum.VOICE_MSG
	body["content"] = ""
	body["voiceUrl"] = "https://dds.dui.ai/runtime/v1/synthesize?voiceId=qianranfa&speed=1.0&volume=100&audioType=wav&text=" + url.PathEscape(Content)
	body["groupid"] = GroupID
	body["atUser"] = 0
	body["voiceBase64Buf"] = ""
	body["pwd"] = "mcoo"
	ret, _ := json.Marshal(body)
	return ret
}

func NewVoiceInstance () helper.ISender {
	return voiceInstance
}


func (th *VoiceSender) SendToUser(UserID int64, Content string,message iotqq.Message) {
	body := th.makeBody(UserID,0,enum.SEND_TO_FRIEND,Content)
	th.DoSend(global.BASE_URL,global.QQ,body)
}

func (th *VoiceSender) SendToGroup(GroupID int64, Content string,message iotqq.Message) {
	body := th.makeBody(GroupID,0,enum.SEND_TO_GROUP,Content)
	th.DoSend(global.BASE_URL,global.QQ,body)
}
