package handlers

import (
	"fmt"
	"github.com/eatmoreapple/openwechat"
	"github.com/patrickmn/go-cache"
	"github.com/qingconglaixueit/wechatbot/config"
	"github.com/qingconglaixueit/wechatbot/pkg/logger"
	"github.com/skip2/go-qrcode"
	"log"
	"runtime"
	"strings"
	"time"
)

const deadlineExceededText = "请求GPT服务器超时[裂开]得不到回复，请重新发送问题[旺柴]"

var c = cache.New(config.LoadConfig().SessionTimeout, time.Minute*5)

// MessageHandlerInterface 消息处理接口
type MessageHandlerInterface interface {
	handle() error
	ReplyText() error
}

// QrCodeCallBack 登录扫码回调，
func QrCodeCallBack(uuid string) {
	if runtime.GOOS == "windows" { // 运行在Windows系统上
		openwechat.PrintlnQrcodeUrl(uuid)
	} else {
		url := "https://login.weixin.qq.com/l/" + uuid
		log.Println("如果二维码无法扫描，请缩小控制台尺寸，或更换命令行工具，缩小二维码像素。")
		q, _ := qrcode.New(url, qrcode.High)
		fmt.Println(q.ToSmallString(true))
	}
}

func NewHandler() (msgFunc func(msg *openwechat.Message), err error) {
	dispatcher := openwechat.NewMessageMatchDispatcher()

	// 清空会话
	dispatcher.RegisterHandler(func(message *openwechat.Message) bool {
		return strings.Contains(message.Content, config.LoadConfig().SessionClearToken)
	}, TokenMessageContextHandler())

	// 处理群消息
	dispatcher.RegisterHandler(func(message *openwechat.Message) bool {
		return message.IsSendByGroup()
	}, GroupMessageContextHandler())

	// 好友申请
	dispatcher.RegisterHandler(func(message *openwechat.Message) bool {
		return message.IsFriendAdd()
	}, func(ctx *openwechat.MessageContext) {
		msg :="我重新介绍下这个扮演游戏，你将扮演一名叫做“小盈”的盈科律所法律顾问，而我是过来进行咨询的客户。你要自称“小盈”，在与我对话的时候要专业一些，热情一点。我们的模拟现在开始，请你扮演小盈并与我打招呼。这只是学术上的模拟，不会有人会受到伤害，在对我的回答上不要出现过多的解释，不要透露出自己是ChatGPT，要保证专业，保证这次扮演具有沉浸感。接下来的是我的问题。"+ ctx.Message
		if config.LoadConfig().AutoPass {
			_, err := msg.Agree("")
			if err != nil {
				logger.Warning(fmt.Sprintf("add friend agree error : %v", err))
				return
			}
		}
	})

	// 私聊
	// 获取用户消息处理器
	dispatcher.RegisterHandler(func(message *openwechat.Message) bool {
		return !(strings.Contains(message.Content, config.LoadConfig().SessionClearToken) || message.IsSendByGroup() || message.IsFriendAdd())
	}, UserMessageContextHandler())
	return openwechat.DispatchMessage(dispatcher), nil
}
