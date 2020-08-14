package main

import (
	"encoding/json"
	iotqq "myiotqq-plugin/model"
)

func main() {
	str := "{\"Content\":\"投票@纱雾酱\",\"UserID\":[154755584]}"
	data := &iotqq.AtInfo{
		Content: "",
		UserID:  nil,
	}
	json.Unmarshal([]byte(str),data)
	print(data.Content)
}
