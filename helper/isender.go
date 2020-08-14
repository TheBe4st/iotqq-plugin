package helper

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"myiotqq-plugin/enum"
	"myiotqq-plugin/global"
	iotqq "myiotqq-plugin/model"
	"net/http"
)

type ISender interface {
	SendToUser(UserID int64, Content string, message iotqq.Message)
	SendToGroup(GroupID int64, Content string, message iotqq.Message)

	//SendTextToUser(UserID int64, Content string)
	//SendTextToGroup(GroupID int64, Content string)
	//
	//SendVoiceToUserFromUrl(UserID int64, url string)
	//SendVoiceToGroupFromUrl(GroupID int64, url string)
	//
	//SendVoiceToUserFromFile(UserID int64, filename string)
	SendVoiceToGroupFromFile(GroupID int64, filename string)
	//
	//SendVoiceToUserFromBase64(UserID int64, base64 string)
	//SendVoiceToGroupFromBase64(GroupID int64, base64 string)
	//
	//SendPicToUserFromUrl(UserID int64, url string)
	SendPicToGroupFromUrl(GroupID int64, url string, content string)
	//
	//SendPicToUserFromFile(UserID int64, filename string)
	//SendPicToGroupFromFile(GroupID int64, filename string)
	//
	//SendPicToUserFromBase64(UserID int64, base64 string)
	//SendPicToGroupFromBase64(GroupID int64, base64 string)
}

type Sender struct {
}

func fileToBase64(fileName string) string {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(data)
}

func (th *Sender) SendVoiceToGroupFromFile(GroupID int64, filename string) {

	tmp := make(map[string]interface{})
	tmp["toUser"] = GroupID
	tmp["sendToType"] = enum.SEND_TO_GROUP
	tmp["sendMsgType"] = enum.VOICE_MSG
	tmp["groupid"] = 0
	tmp["content"] = ""
	tmp["atUser"] = 0
	tmp["voiceUrl"] = ""
	tmp["voiceBase64Buf"] = fileToBase64(filename)
	tmp["pwd"] = "mcoo"
	tmp1, _ := json.Marshal(tmp)
	th.DoSend(global.BASE_URL, global.QQ, tmp1)
}

func (th *Sender) SendPicToGroupFromUrl(GroupID int64, url string, content string) {
	//发送图文信息
	tmp := make(map[string]interface{})
	tmp["toUser"] = GroupID
	tmp["sendToType"] = enum.SEND_TO_GROUP
	tmp["sendMsgType"] = enum.PIC_MSG
	tmp["picBase64Buf"] = ""
	tmp["fileMd5"] = ""
	tmp["picUrl"] = url
	tmp["content"] = content
	tmp["groupid"] = 0
	tmp["atUser"] = 0
	tmp["pwd"] = "mcoo"
	tmp1, _ := json.Marshal(tmp)
	th.DoSend(global.BASE_URL, global.QQ, tmp1)
}

func (th *Sender) DoSend(BaseUrl string, QQ string, body []byte) {
	resp, err := (http.Post(BaseUrl+"/v1/LuaApiCaller?funcname=SendMsg&timeout=10&qq="+QQ, "application/json", bytes.NewBuffer(body)))
	if err != nil {
		log.Fatal(err)
		return
	}
	defer resp.Body.Close()
	res, _ := ioutil.ReadAll(resp.Body)
	log.Println("LuaApiCaller接口返回：" + string(res))
}
