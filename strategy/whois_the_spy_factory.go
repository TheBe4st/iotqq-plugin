package strategy

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	iotqq "myiotqq-plugin/model"
	"strings"
	"time"
)

type Key struct {
	Key1 string
	Key2 string
}

var keys = []Key{
	{"灌汤包","小笼包"},
	{"酷我音乐","酷狗音乐"},
	{"薰衣草","满天星bai"},
	{"富二代","高富帅"},
	{"生活费","零花钱"},
	{"麦克风","扩音器"},
	{"郭德纲","周立波"},
	{"图书馆","图书店"},
	{"男朋友","前男友"},
	{"洗衣粉","皂角粉"},
	{"牛肉干","猪肉脯"},
	{"泡泡糖","棒棒糖"},
	{"小沈阳","宋小宝"},
	{"张韶涵","王心凌"},
	{"刘诗诗","刘亦菲"},
	{"甄嬛传","红楼梦"},
	{"甄子丹","李连杰"},
	{"包青天","狄仁杰"},
	{"大白兔","金丝猴"},
	{"果粒橙","鲜橙多"},
	{"沐浴露","沐浴盐"},
	{"洗发露","护发素"},
	{"自行车","电动车"},
	{"班主任","辅导员"},
	{"过山车","碰碰车"},
	{"铁观音","碧螺春"},
	{"丑小鸭","灰姑娘"},
	{"十面埋伏","四面楚歌"},
	{"成吉思汗","努尔哈赤"},
	{"谢娜张杰","邓超孙俪"},
	{"福尔摩斯","工藤新一"},
	{"贵妃醉酒","黛玉葬花"},
	{"流星花园","花样男子"},
	{"神雕侠侣","天龙八部"},
	{"天天向上","非诚勿扰"},
	{"勇往直前","全力以赴"},
	{"鱼香肉丝","四喜丸子"},
	{"语无伦次","词不达意"},
	{"鼠目寸光","井底之蛙"},
	{"近视眼镜","隐形眼镜"},
	{"美人心计","倾世皇妃"},
	{"夏家三千金","爱情睡醒了"},
	{"降龙十八掌","九阴白骨爪"},
	{"红烧牛肉面","香辣牛肉面"},
	{"江南style","最炫民族风"},
	{"脚踏车","自行车"},
	{"口香糖","木糖醇"},
	{"老佛爷","老天爷"},
	{"金丝猴","大白兔(奶糖)"},
	{"近视眼镜","隐形眼镜"},
	{"两小无猜","青梅竹马"},
	{"龙凤呈祥","鸳鸯戏水"},
	{"麻婆豆腐","皮蛋豆腐"},
	{"江南style","最炫民族风"},
	{"降龙十八掌","九阴白骨爪"},
	{"福尔摩斯-工藤新","福尔摩斯-柯南"},
	{"梁山伯与祝英台","罗密欧与朱丽叶"},
}

const VOTE_INVALID = -1

const GAME_CREATE = 0
const GAME_RUNNING = 1
const GAME_VOTEING = 2
const GAME_END = -1

const PLAYER_CREATE = 0
const PLAYER_SPEAKING = 1


const SPEAK_TIME = 20 *2 * time.Second
const VOTE_TIME = 15 * 2 * time.Second

const MAX_PLAYER = 8
const MIN_PLAYER = 6

var games map[int64]*Game

type Players []*Player
type Game struct {
	Players    Players
	Spy        Players
	Creator    string
	Wheel      int
	SpyWord    string
	NormalWord string
	GameStatus int
	VoteChan chan int // 投票结束 通知游戏继续进行~
}
type Player struct {
	QQ        int64
	NickName  string
	IsSpy     bool // 这货是不是个卧底
	IsOut     bool // 这货是不是出局了
	VoteEntry [100]*Vote
	SpeakChannel chan time.Time
	PlayerStatus int
	VoteCount [100]int  //每轮得票数
}
type Vote struct {
	ToUserQQ   *Player
	FromUserQQ *Player
}

func init() {
	games = make(map[int64]*Game)
}

var _ MsgFactory = WhoIsTheSpyMsgFactory

func WhoIsTheSpyMsgFactory(tp string) IGroupMsgExecutor {
	defer func() {
		if err := recover() ; err != nil {
			log.Fatal(err)
		}
	}()

	if strings.Contains(tp,"投票") && strings.Contains(tp,"UserID") {
		return DoVote
	}
	switch tp {
	case "谁是卧底":
		return CreateGame
	case "创建房间":
		return CreateGame
	case "开始游戏":
		return BeginGame
	case "玩家列表":
		return PlayerList
	case "加入游戏":
		return JoinGame
	case "过":
		return PassSpeak
	case "开始投票":
		return BeginVote
	case "结束投票":
		return EndVote
	case "结束游戏":
		return EndGame
	case "私聊测试":
		return PrivateTest
	case "帮助":
		return PrintHelper
	default:
		return UnSupportOperator
	}
}

func BeginGame(args iotqq.Message) {
	game := getCurrentGame(args)
	if game == nil {
		CreateGame(args)
		return
	}
	if game.GameStatus >= GAME_RUNNING {
		return
	}
	if len(game.Players) < MIN_PLAYER {
		Sender.SendToGroup(args.GetGroupId(),"玩家数量太少，纱雾无法开始游戏QAQ",args)
		return
	}
	game.GameStatus = GAME_RUNNING

	// 从词库抓一个词
	rand.Seed(time.Now().UnixNano())
	{
		a := rand.Intn(len(keys))
		b :=rand.Intn(10)
		if b > 5 {
			game.NormalWord = keys[a].Key1
			game.SpyWord = keys[a].Key2
		} else {
			game.NormalWord = keys[a].Key2
			game.SpyWord = keys[a].Key1
		}
	}
	// 随机挑选 spyNum 个幸运儿当作卧底
	spyNum := int(math.Floor(float64(len(game.Players) / MIN_PLAYER)))
	for len(game.Spy) < spyNum {
		r := rand.Intn(len(game.Players))
		luckDog := game.Players[r]
		isRept := false
		for _,item := range game.Spy {
			if luckDog.QQ == item.QQ {
				isRept = true
			}
		}
		if !isRept {
			game.Spy = append(game.Spy, game.Players[r])
			game.Players[r].IsSpy = true
		}
	}

	// 挨个私聊发给他们吧。。
	for _, item := range game.Players {
		word := ""
		if item.IsSpy {
			word = game.SpyWord
		} else {
			word = game.NormalWord
		}
		Sender.SendToUser(item.QQ, fmt.Sprintf("你的关键词是：%s ，不要告诉别人呀~", word), args)
		time.Sleep(1*time.Second)
	}

	Sender.SendToGroup(args.GetGroupId(), fmt.Sprintf("游戏开始~请查看私聊消息获取各自的词，如果没收到消息，请先添加我为好友哦"), args)

	time.Sleep(time.Second)
	// 启动一个线程去定时轮次
	go wheel(args)
}

func wheel(args iotqq.Message) {
	for true {
		game := getCurrentGame(args)
		if game == nil || game.GameStatus == GAME_END {
			return
		}
		for _, item := range game.Players {
			game := getCurrentGame(args)
			if game == nil || game.GameStatus == GAME_END {
				return
			}
			if item.IsOut {
				continue
			}
			Sender.SendToGroup(args.GetGroupId(), fmt.Sprintf("第 %d 轮，每个玩家发言20秒，请 %s 进行演讲，演讲完毕请输入“过”",game.Wheel, item.NickName), args)
			item.PlayerStatus = PLAYER_SPEAKING
			// 超过SPEAK_TIME秒 去把阻塞当前玩家的值给读出来
			// 准备工作都做完了，把自己卡住等 player慢慢BB
			c := time.After(SPEAK_TIME)
			go func() {
				<- c
				//  SPEAK_TIME 之后 如果状态还是讲话 就写一个东东进去
				if item.PlayerStatus == PLAYER_SPEAKING {
					item.PlayerStatus = PLAYER_CREATE
					item.SpeakChannel <- time.Now()
				}
			}()
			<- item.SpeakChannel
		}

		// 所有发言完毕 开始投票~
		BeginVote(args)
		<- game.VoteChan
		EndVote(args)
		if game.IsDone() {
			EndGame(args)
			return
		}
		game.Wheel += 1
	}
}

func PassSpeak(args iotqq.Message) {
    game :=	getCurrentGame(args)
    if game == nil {
    	return
	}
	 player := game.FindPlayerByQQ(args.GetSendUserId())
	 if player == nil || player.PlayerStatus != PLAYER_SPEAKING {
		 return
	 }
	player.PlayerStatus = PLAYER_CREATE
	player.SpeakChannel <- time.Now()
}

func CreateGame(args iotqq.Message) {
	PrintHelper(args)
	time.Sleep(2 * time.Second)
	groupId := args.GetGroupId()
	if games[groupId] != nil {
		// 生成消息 告诉用户这个房间的玩家列表
		Sender.SendToGroup(groupId, fmt.Sprintf("该群已经有了一个房间，请等待游戏结束或发送“结束游戏”来结束本场游戏~ %s", genPlayerList(getCurrentGame(args))), args)
		return
	}
	game := &Game{
		Players:    make(Players, 0),
		Wheel:      1,
		GameStatus: GAME_CREATE,
		Creator:    args.GetSendUserNickName(),
		Spy:        make(Players, 0),
		SpyWord:    "暂无",
		NormalWord: "暂无",
		VoteChan: make(chan int),
	}
	games[groupId] = game
	Sender.SendToGroup(groupId, fmt.Sprintf("%s 成功创建房间 %s ，如果60S没有足够玩家加入则房间将被释放", args.GetSendUserNickName(), args.GetGroupName()), args)
	c := time.After(60 * time.Second)
	go func() {
		<-c
		// 60秒后 看看是开始游戏还是解散房间
		if len(game.Players) < MIN_PLAYER {
			//解散房间
			EndGame(args)
		} else {
			// 开始游戏
			if game.GameStatus != GAME_RUNNING {
				BeginGame(args)
			}
		}
	}()
}

func JoinGame(args iotqq.Message) {
	game := getCurrentGame(args)
	if game == nil {
		// 如果加入游戏的时候当前房间还没创建一个游戏 创建并加入
		CreateGame(args)
		JoinGame(args)
	} else {
		if len(game.Players) >= MAX_PLAYER {
			Sender.SendToGroup(args.GetGroupId(), fmt.Sprintf("房间已经满了哦，下次快点来吧~"), args)
			return
		}

		if game.GameStatus >= GAME_RUNNING {
			Sender.SendToGroup(args.GetGroupId(), fmt.Sprintf("游戏已经开始啦 不能加入哦"), args)
			return
		}

		// 先看看这个B是不是加过房间了
		for _, player := range game.Players {
			if player.QQ == args.GetSendUserId() {
				Sender.SendToGroup(args.GetGroupId(), fmt.Sprintf("你已经在房间里啦~"), args)
				return
			}
		}

		game.Players = append(game.Players, &Player{
			QQ:        args.GetSendUserId(),
			NickName:  args.GetSendUserNickName(),
			IsSpy:     false,
			IsOut:     false,
			SpeakChannel: make(chan time.Time),
		})
		Sender.SendToGroup(args.GetGroupId(), fmt.Sprintf("%s 成功加入房间，正在等待玩家加入,当前玩家数量：%d %s", args.GetSendUserNickName(), len(game.Players), genPlayerList(game)), args)
	}
}

func EndGame(args iotqq.Message) {
	// 结束游戏的时候 并不代表我结束当前机器人游戏模式的状态 解除游戏模式的状态应在 group_msg_selector中把current_factory切换成其他的MsgFactory
	// 此处仅仅释放掉当前的游戏房间 并打印游戏结果
	groupId := args.GetGroupId()
	game := getCurrentGame(args)
	if game == nil {
		return //说明游戏已经结束或者压根没开始
	}
	game.GameStatus = GAME_END
	games[groupId] = nil
	Sender.SendToGroup(groupId, fmt.Sprintf("游戏已结束，卧底关键词：%s，平民关键词：%s，玩家身份信息：%s", game.SpyWord, game.NormalWord, genPlayerListWithSpy(game, true)), args)
}

func PlayerList(args iotqq.Message) {
	game := getCurrentGame(args)
	Sender.SendToGroup(args.GetGroupId(), genPlayerList(game), args)
}

func PrivateTest(args iotqq.Message) {
	game := getCurrentGame(args)
	if game == nil {
		return
	}
	for _, player := range game.Players {
		Sender.SendToUser(player.QQ, "为了保证游戏顺利进行，这是一条测试私聊的消息", args)
	}
}
func BeginVote(args iotqq.Message) {
	game := getCurrentGame(args)
	if game == nil || game.GameStatus == GAME_VOTEING || game.GameStatus == GAME_END {
		return
	}
	game.GameStatus = GAME_VOTEING
	Sender.SendToGroup(args.GetGroupId(),fmt.Sprintf("第 %d 轮，进入投票阶段~，使用 投票@xxx来进行投票",game.Wheel),args)
	c := time.After(VOTE_TIME)
	go func() {
		//VOTE_TIME 秒后告诉wheel 投票结束 进行下一阶段
		<- c
		game.VoteChan <- 1
	}()
}

func DoVote(args iotqq.Message) {
	log.Print("进入投票分支")
    game :=	getCurrentGame(args)
    if game == nil {
		return
	}
	//判断 游戏的状态
	if game.GameStatus != GAME_VOTEING {
		log.Print(fmt.Sprintf("投票发起人：%s不是游戏状态不能投票！",args.GetSendUserNickName()))
		return
	}
	err,atInfo := args.GetAtInfo()
	if err != nil {
		return
	}
	from := game.FindPlayerByQQ(args.GetSendUserId())
	to := game.FindPlayerByQQ(atInfo.UserID[0])
	// 校验一下 双方必须都是玩家才阔以
	if from == nil || to == nil{
		log.Print("无效的投票，双方不是玩家！")
		return
	}
	// 生成一条投票纪录 放进 投票者的 vote里
	vote := &Vote{
		ToUserQQ:   to,
		FromUserQQ: from,
	}
	from.VoteEntry[game.Wheel] = vote
	Sender.SendToGroup(args.GetGroupId(),fmt.Sprintf("%s 投票给了 %s ~",args.GetSendUserNickName(),vote.ToUserQQ.NickName),args)
}

func (th *Game) IsDone() bool {

	if th.GameStatus == GAME_END {
		return true
	}
	// 检查用户的投票情况 返回投票
	ret := true
	for _, player := range th.Players {
		if !player.IsOut && !player.IsSpy{
			ret = false
			break
		}
	}
	return ret
}

func EndVote(args iotqq.Message) {
	Sender.SendToGroup(args.GetGroupId(),fmt.Sprintf("投票结束~正在进行整理得票情况..."),args)
	game := getCurrentGame(args)
	if game.GameStatus == GAME_END {
		return
	}
	game.GameStatus = GAME_RUNNING

	count := make(map[*Player][]*Player)
	// 把该死的玩家给弄死
	for _, player := range game.Players {
		vote := player.VoteEntry[game.Wheel]
		if vote == nil {
			continue
		}
		if count[vote.ToUserQQ] == nil {
			count[vote.ToUserQQ] = make([]*Player,0)
		}
		count[vote.ToUserQQ] = append(count[vote.ToUserQQ],vote.FromUserQQ)
	}

	// 计算玩家们的本轮得票数
	for k, v := range count {
		k.VoteCount[game.Wheel] = len(v)
	}

	// 找出最大的得票数
	var max *Player = nil
	for _, player := range game.Players {
		if player.VoteCount[game.Wheel] > 0 {
			if max == nil {
				max = player
			} else if player.VoteCount[game.Wheel] > max.VoteCount[game.Wheel] {
				max = player
			}
		}
	}

	//生成投票结果文本
	str := "投票结果：\n"
	for to, from := range count {
		for _, item := range from {
			str += item.NickName + ","
		}
		str += "投给了 "
		str += to.NickName
		str += "\n"
	}
	Sender.SendToGroup(args.GetGroupId(),str,args)
	time.Sleep(time.Second)
	if max == nil {
		Sender.SendToGroup(args.GetGroupId(),"无人投票，无人出局",args)
	} else {
		// 找找有没有平票的 qwq
		for _, player := range game.Players {

			//跳过自己本身
			if player == max {
				continue
			}

			if player.VoteCount[game.Wheel] == max.VoteCount[game.Wheel] {
				// 出现平票 各自安好
				Sender.SendToGroup(args.GetGroupId(),fmt.Sprintf("%s 和 %s 平票！无人出局",player.NickName,max.NickName),args)
				return
			}
		}
		// 木的平票的 死了
		max.IsOut = true
		// 把死了的身份爆出来
		temp := ""
		if max.IsSpy {
			temp = "卧底"
		} else {
			temp = "平民"
		}
		Sender.SendToGroup(args.GetGroupId(),fmt.Sprintf("%s 死了，他的身份是 %s",max.NickName,temp),args)
	}
	time.Sleep(time.Second)
	Sender.SendToGroup(args.GetGroupId(),genLivePlayer(game),args)
}
func UnSupportOperator(args iotqq.Message) {
}

func getCurrentGame(args iotqq.Message) *Game {
	groupId := args.GetGroupId()
	return games[groupId]
}

func genPlayerList(game *Game) string {
	return genPlayerListWithSpy(game, false)
}

func genPlayerListWithSpy(game *Game, withSpy bool) string {
	if game == nil {
		return ""
	}
	ret := "当前玩家列表:\n"
	players := game.Players
	for i := 0; i < len(players); i++ {
		if withSpy {
			strIsSpy := ""
			if players[i].IsSpy {
				strIsSpy = "卧底"
			} else {
				strIsSpy = "平民"
			}

			strIsOut := ""
			if players[i].IsOut {
				strIsOut = "否"
			} else {
				strIsOut = "是"
			}
			ret += fmt.Sprintf("%d、 %s 身份： %s 存活：%s\n", i + 1, players[i].NickName, strIsSpy,strIsOut)
		} else {
			ret += fmt.Sprintf("%d、 %s\n", i + 1, players[i].NickName)
		}
	}
	return ret
}

func genLivePlayer(game *Game) string{
	if game == nil {
		return ""
	}
	ret := "场上剩余:\n"
	players := game.Players
	for i := 0; i < len(players); i++ {
		if !players[i].IsOut {
			ret	+= fmt.Sprintf("%d、 %s\n", i + 1, players[i].NickName)
		}
	}
	return ret
}

func PrintHelper(args iotqq.Message) {
	helpMsg := `当前机器人已切换到《谁是卧底》游戏模式，其他功能将暂不可用，如需恢复请使用“聊天模式”指令恢复到聊天模式
创建房间	创建一个房间，60秒内没有足够玩家加入或没有强制开始则自动释放房间
开始游戏	人数不足时使用此指令强制开始游戏
加入游戏	已开局的游戏不能加入，如果没有房间会自动创建房间
开始投票	使用此指令跳过等待时间，直接进入投票阶段
投票@	输入“投票”二字并艾特你要摸的人
结束游戏	强制结束本局游戏并揭晓结果~
过		发言完了就说个过
玩家列表	查看加入游戏的玩家`
	Sender.SendToGroup(args.GetGroupId(), fmt.Sprintf(helpMsg), args)
}

func (game *Game) FindPlayerByQQ(qq int64) *Player {
	for _ ,item := range game.Players {
		//让我看看哪个小可爱是自己
		if qq == item.QQ {
			return item
		}
	}
	return nil
}

