package room

import (
	"errors"
	"reflect"

	log "github.com/sirupsen/logrus"

	"local.com/abc/game/protocol"
	"local.com/abc/game/util"
)

const (
	EventConfigChanged   = 1  //房间配置更改
	EventFrameUpdate     = 0  // 帧更新
	EventCreateTable     = 1  // 创建桌子
	EventDeleteTable     = 2  // 删除桌子
	EventUserSitdown     = 5  // 用户坐下
	EventUserStandup     = 6  // 用户起立
	EventRoomClose       = 7  //房间关闭通知
	EventUserOnEntryRoom = 10 //进入房间
	EventUserOnReset     = 11 //退出房间
	EventUserOnExitRoom  = 11 //退出房间
)

var (
	ErrorIncorrectFrameType = errors.New("incorrect frame type")
	ErrorServiceNotBind     = errors.New("service not bind")
	ErrorServiceBusy        = errors.New("service is busy")
)

var (
	signal *util.AppSignal
	Coder  protocol.Coder
)

func RegistMsg(id int32, arg interface{}) {
	t := reflect.TypeOf(arg)
	if _, ok := Coder.GetMsgId(t); ok {
		log.Fatalf("message %v is already registered", t)
	}
	Coder.SetMsgId(t, id)
}
