package strategy

var current MsgFactory = GroupMsgFactory

func SelectFactory(tp string) IGroupMsgExecutor {
	switch tp {
	case "谁是卧底":
		current = WhoIsTheSpyMsgFactory
	case "退出游戏":
		current = GroupMsgFactory
	case "聊天模式":
		current = GroupMsgFactory
	}
	return current(tp)
}
