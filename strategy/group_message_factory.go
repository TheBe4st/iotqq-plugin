package strategy

import (
	"myiotqq-plugin/helper"
	"myiotqq-plugin/helper/impl"
	iotqq "myiotqq-plugin/model"
	"strings"
)

var isText = true
var Sender helper.ISender = impl.NewTextInstance()
var Zanok []int64

type IGroupMsgExecutor func(message iotqq.Message)
type MsgFactory func(tp string) IGroupMsgExecutor

func GroupMsgFactory(tp string) IGroupMsgExecutor {
	switch tp {
	case "菜单":
		return MenuExecutor
	case "转换模式":
		return SwitchExecutor
	case "test":
		return TestExecutor
	case "来点涩图":
		return ColorPicExecutor
	case "赞我":
		return ZanExecutor
	case "聊天模式":
		return HelpExecutor
	}
	if strings.Contains(tp, "来点涩图") {
		return ColorPicExecutor
	}

	if strings.Contains(tp, "洛天依") {
		return LuotianyiExecutor
	}
	if strings.Contains(tp, "歌") {
		return SingExecutor
	}

	return UnSupportsExecutor
}
