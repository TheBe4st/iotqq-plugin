package iotqq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"myiotqq-plugin/global"
	"net/http"
)

type QQinfo struct {
	Code    int    `json:"code"`
	Data    Data1  `json:"data"`
	Default int    `json:"default"`
	Message string `json:"message"`
	Subcode int    `json:"subcode"`
}
type Data1 struct {
	AvatarURL     string `json:"avatarUrl"`
	Bitmap        string `json:"bitmap"`
	Commfrd       int    `json:"commfrd"`
	Friendship    int    `json:"friendship"`
	Greenvip      int    `json:"greenvip"`
	IntimacyScore int    `json:"intimacyScore"`
	IsFriend      int    `json:"isFriend"`
	Logolabel     string `json:"logolabel"`
	Nickname      string `json:"nickname"`
	Qqvip         int    `json:"qqvip"`
	Qzone         int    `json:"qzone"`
	Realname      string `json:"realname"`
	Redvip        int    `json:"redvip"`
	Smartname     string `json:"smartname"`
	Uin           int    `json:"uin"`
}
type QQ struct {
	Cont int
}
type PSkey struct {
	Connect     string `json:"connect"`
	Docs        string `json:"docs"`
	Docx        string `json:"docx"`
	Game        string `json:"game"`
	Gamecenter  string `json:"gamecenter"`
	Imgcache    string `json:"imgcache"`
	MTencentCom string `json:"m.tencent.com"`
	Mail        string `json:"mail"`
	Mma         string `json:"mma"`
	Now         string `json:"now"`
	Office      string `json:"office"`
	Openmobile  string `json:"openmobile"`
	Qqweb       string `json:"qqweb"`
	Qun         string `json:"qun"`
	Qzone       string `json:"qzone"`
	QzoneCom    string `json:"qzone.com"`
	TenpayCom   string `json:"tenpay.com"`
	Ti          string `json:"ti"`
	Vip         string `json:"vip"`
	Weishi      string `json:"weishi"`
}
type Cook struct {
	ClientKey string `json:"ClientKey"`
	Cookies   string `json:"Cookies"`
	Gtk       string `json:"Gtk"`
	Gtk32     string `json:"Gtk32"`
	PSkey     PSkey  `json:"PSkey"`
	Skey      string `json:"Skey"`
}
type Data2 struct {
	Date   string `json:"date"`
	City   string `json:"city"`
	Adcode string `json:"adcode"`
	Min    string `json:"min"`
	Max    string `json:"max"`
	Type   string `json:"type"`
	Air    string `json:"air"`
	Wind   string `json:"wind"`
}
type Weather struct {
	Code int   `json:"code"`
	Data Data2 `json:"data"`
}
type CurrentPacket struct {
	Data      Data   `json:"Data"`
	WebConnID string `json:"WebConnId"`
}
type Data struct {
	Content       string      `json:"Content"`
	FromGroupID   int64       `json:"FromGroupId"`
	FromGroupName string      `json:"FromGroupName"`
	FromNickName  string      `json:"FromNickName"`
	FromUserID    int64       `json:"FromUserId"`
	MsgRandom     int         `json:"MsgRandom"`
	MsgSeq        int         `json:"MsgSeq"`
	MsgTime       int         `json:"MsgTime"`
	MsgType       string      `json:"MsgType"`
	RedBaginfo    interface{} `json:"RedBaginfo"`
}
type Message struct {
	CurrentPacket CurrentPacket `json:"CurrentPacket"`
	CurrentQQ     int64         `json:"CurrentQQ"`
}
type Channel struct {
	Channel string `json:"channel"`
}

type AtInfo struct {
	Content string
	UserID []int64
}

func (th Message) GetGroupId() int64 {
	return th.CurrentPacket.Data.FromGroupID
}
func (th Message) GetGroupName() string {
	return th.CurrentPacket.Data.FromGroupName
}
func (th Message) GetSendUserId() int64 {
	return th.CurrentPacket.Data.FromUserID
}
func (th Message) GetSendUserNickName() string {
	return th.CurrentPacket.Data.FromNickName
}
func (th Message) GetAtInfo() (error,AtInfo) {
	atInfo := AtInfo{}
	if err := json.Unmarshal([]byte(th.CurrentPacket.Data.Content),&atInfo); err != nil {
		log.Fatal(err)
		return err, atInfo
	}
	return nil,atInfo
}
func Zan(qq1 int, err error) {
	//名片点赞
	tmp := make(map[string]interface{})
	tmp["UserID"] = qq1
	tmp["pwd"] = "mcoo"
	tmp1, _ := json.Marshal(tmp)
	fmt.Println(string(tmp1))
	resp, err := http.Post(global.BASE_URL+"/v1/LuaApiCaller?funcname=QQZan&timeout=10&qq="+global.QQ, "application/json", bytes.NewBuffer(tmp1))
	if err != nil {
		log.Fatal(err)
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
}
