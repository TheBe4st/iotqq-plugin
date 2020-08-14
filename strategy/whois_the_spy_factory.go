package strategy

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	iotqq "myiotqq-plugin/model"
	"strings"
	"time"
)

const VOTE_INVALID = -1

const GAME_CREATE = 0
const GAME_RUNNING = 1
const GAME_VOTEING = 2
const GAME_END = -1

const PLAYER_CREATE = 0
const PLAYER_SPEAKING = 1


const SPEAK_TIME = 20 * time.Second
const VOTE_TIME = 15 * time.Second

const MAX_PLAYER = 8
const MIN_PLAYER = 2

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
	VoteEntry []*Vote
	SpeakChannel chan time.Time
	PlayerStatus int
}
type Vote struct {
	ToUserQQ   int64
	FromUserQQ int64
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
	if game.GameStatus >= GAME_RUNNING {
		return
	}
	if len(game.Players) < MIN_PLAYER {
		Sender.SendToGroup(args.GetGroupId(),"玩家数量太少，纱雾无法开始游戏QAQ",args)
		return
	}
	game.GameStatus = GAME_RUNNING

	// 从词库抓一个词
	game.NormalWord = "爬山"
	game.SpyWord = "攀岩"

	// 随机挑选 spyNum 个幸运儿当作卧底
	spyNum := int(math.Floor(float64(len(game.Players) / MIN_PLAYER)))
	rand.Seed(time.Now().UnixNano())
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
	game := getCurrentGame(args)
	for true {
		if game.GameStatus == GAME_END {
			return
		}
		for _, item := range game.Players {
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
		CheckGame(args)
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
			VoteEntry: make([]*Vote, 0),
			SpeakChannel: make(chan time.Time),
		})
		Sender.SendToGroup(args.GetGroupId(), fmt.Sprintf("%s 成功加入房间，正在等待玩家加入,当前玩家数量：%d %s", args.GetSendUserNickName(), len(game.Players), genPlayerList(game)), args)
	}
}

func CheckGame(args iotqq.Message) {
	game := getCurrentGame(args)
	// 检查用户的投票情况 返回投票
	print(game)
}

func EndGame(args iotqq.Message) {
	// 结束游戏的时候 并不代表我结束当前机器人游戏模式的状态 解除游戏模式的状态应在 group_msg_selector中把current_factory切换成其他的MsgFactory
	// 此处仅仅释放掉当前的游戏房间 并打印游戏结果
	groupId := args.GetGroupId()
	game := getCurrentGame(args)
	games[groupId] = nil
	if game == nil {
		return //说明游戏已经结束或者压根没开始
	}
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
	if game.GameStatus == GAME_VOTEING {
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
	message := args.CurrentPacket.Data.Content
	//判断 游戏的状态
	atInfo := iotqq.AtInfo{}
	if err := json.Unmarshal([]byte(message),&atInfo); err != nil {
		log.Fatal(err)
		Sender.SendToGroup(args.GetGroupId(),fmt.Sprintf("因为内部错误， %s 的投票失败了",args.GetSendUserNickName()),args)
		return
	}

	// 生成一条投票纪录 放进 投票者的 vote里
	vote := &Vote{
		ToUserQQ:   atInfo.UserID[0],
		FromUserQQ: args.GetSendUserId(),
	}
	from := game.FindPlayerByQQ(vote.FromUserQQ)
	to := game.FindPlayerByQQ(vote.ToUserQQ)

	// 校验一下 双方必须都是玩家才阔以
	if from == nil {
		Sender.SendToGroup(args.GetGroupId(),fmt.Sprintf("%s 不是游戏的参与者哦 ~",from.NickName),args)
		return
	}
	if to == nil {
		Sender.SendToGroup(args.GetGroupId(),fmt.Sprintf("%s 不是游戏的参与者哦 ~",to.NickName),args)
		return
	}
	player := game.FindPlayerByQQ(vote.FromUserQQ)
	player.VoteEntry = append(player.VoteEntry,vote)
	Sender.SendToGroup(args.GetGroupId(),fmt.Sprintf("%s 投票给了 %s ~",args.GetSendUserNickName(),game.FindPlayerByQQ(vote.ToUserQQ).NickName),args)
}

func EndVote(args iotqq.Message) {
	Sender.SendToGroup(args.GetGroupId(),fmt.Sprintf("投票结束~正在进行整理得票情况..."),args)
	game := getCurrentGame(args)
	game.GameStatus = GAME_RUNNING
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
			ret += fmt.Sprintf("%d、 %s 身份： %s\n", i, players[i].NickName, strIsSpy)
		} else {
			ret += fmt.Sprintf("%d、 %s\n", i, players[i].NickName)
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

