package strategy

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"myiotqq-plugin/helper/impl"
	iotqq "myiotqq-plugin/model"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func SingExecutor(message iotqq.Message) {
	data := message.CurrentPacket.Data

	baseDir := "./resource/mengmengsing/"
	files, _ := ioutil.ReadDir(baseDir)
	l := len(files)
	rand.Seed(time.Now().Unix())
	pos := rand.Intn(l)
	fileName := baseDir + files[pos].Name()
	Sender.SendVoiceToGroupFromFile(int64(data.FromGroupID), fileName)
}
func LuotianyiExecutor(message iotqq.Message) {
	data := message.CurrentPacket.Data

	baseDir := "./resource/luotianyi/"
	files, _ := ioutil.ReadDir(baseDir)
	l := len(files)
	rand.Seed(time.Now().Unix())
	pos := rand.Intn(l)
	fileName := baseDir + files[pos].Name()
	Sender.SendVoiceToGroupFromFile(int64(data.FromGroupID), fileName)
}

func ZanExecutor(message iotqq.Message) {
	mess := message.CurrentPacket.Data
	ok := true
	for i := 0; i < len(Zanok); i++ {
		if Zanok[i] == mess.FromUserID {
			ok = false
		}
	}
	if ok {
		Sender.SendToGroup(int64(mess.FromGroupID), "正在赞，可能需要50s时间🤣", message)
		for i := 1; i <= 50; i++ {
			iotqq.Zan(strconv.Atoi(strconv.FormatInt(mess.FromUserID, 10)))
			time.Sleep(time.Second * 1)
		}
		Sender.SendToGroup(int64(mess.FromGroupID), "已经赞了50次，如果没有成功，可能是腾讯服务器限制了！", message)
		Zanok = append(Zanok, mess.FromUserID)
	} else {
		Sender.SendToGroup(int64(mess.FromGroupID), "之前已经赞了", message)
	}
	return
}

func ColorPicExecutor(message iotqq.Message) {
	defer func() {
		if err := recover(); err != nil {
			log.Fatal(err)
		}
	}()
	data := message.CurrentPacket.Data
	arr := strings.Split(data.Content, " ")
	param := url.Values{}
	param.Add("r18", "2")
	param.Add("apikey", "465493885f16a8788b10d4")
	param.Add("num", "5")
	param.Add("size1200", "true")
	if len(arr) > 1 {
		param.Add("keyword", arr[len(arr)-1])
	}
	paramEncode := param.Encode()
	log.Print("涩图接口开始调用")
	resp, err := http.Get("https://api.lolicon.app/setu?" + paramEncode)
	log.Print("涩图接口调用完成")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer resp.Body.Close()
	res, _ := ioutil.ReadAll(resp.Body)
	resMap := make(map[string]interface{})
	json.Unmarshal(res, &resMap)
	if resMap["code"].(float64) != 0 {
		return
	}
	for index, item := range resMap["data"].([]interface{}) {
		sexPicUrl := item.(map[string]interface{})["url"]
		log.Printf("正在发送第 %d 张图到群：%d", index, int64(data.FromGroupID))
		go Sender.SendPicToGroupFromUrl(int64(data.FromGroupID), sexPicUrl.(string), "")
	}
	Sender.SendToGroup(int64(data.FromGroupID), "你觉得这图怎么样", message)
}

func TestExecutor(message iotqq.Message) {
	data := message.CurrentPacket.Data
	Sender.SendToGroup(int64(data.FromGroupID), "哈哈哈哈哈哈哈 我是神经病QAQ", message)
}
func HelpExecutor(message iotqq.Message) {
	data := message.CurrentPacket.Data
	Sender.SendToGroup(int64(data.FromGroupID), "当前转换为默认聊天模式", message)
}

func MenuExecutor(message iotqq.Message) {
	data := message.CurrentPacket.Data
	Sender.SendToGroup(int64(data.FromGroupID), "你好,我是纱雾酱😊\n回复：赞我、给你点50个赞哟😘，回复：来点涩图看涩图哦", message)
}

func SwitchExecutor(message iotqq.Message) {
	data := message.CurrentPacket.Data

	if isText {
		Sender = impl.NewVoiceInstance()
		isText = !isText
		Sender.SendToGroup(int64(data.FromGroupID), "转换成功，当前为语音模式", message)
	} else {
		Sender = impl.NewTextInstance()
		Sender.SendToGroup(int64(data.FromGroupID), "转换成功，当前为文字模式", message)
	}
}

func UnSupportsExecutor(message iotqq.Message) {
	data := message.CurrentPacket.Data
	return
	baseDir := "./resource/dinggong/"
	files, _ := ioutil.ReadDir(baseDir)
	l := len(files)
	rand.Seed(time.Now().Unix())
	pos := rand.Intn(l)
	fileName := baseDir + files[pos].Name()
	if strings.Contains(data.Content, "GroupPic") {
		i := rand.Intn(100)
		if i > 17 {
			return
		}
		// 随机抽取钉宫语音包发送
		Sender.SendVoiceToGroupFromFile(int64(data.FromGroupID), fileName)
	} else {
		i := rand.Intn(100)
		if i > 17 {
			return
		}
		//青云API 兜底
		data := message.CurrentPacket.Data
		resp, err := http.Get("http://api.qingyunke.com/api.php?key=free&appid=0&msg=" + data.Content)
		if err != nil {
			log.Fatal(err)
			return
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		v := make(map[string]interface{}, 10)
		json.Unmarshal(body, &v)
		Sender.SendToGroup(int64(data.FromGroupID), v["content"].(string), message)
	}

}
