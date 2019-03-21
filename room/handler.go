package room

import (
	log "github.com/sirupsen/logrus"

	//"local.com/abc/game/db"
	//"local.com/abc/game/model"
	"local.com/abc/game/protocol"
)

// 主线程调用，玩家上线
func userOnline(sess *Session, uid int32) {
	log.Debugf("userOnline:%v", sess)
	// 登录检查
	oldSess := GetUser(uid)
	if oldSess == sess {
		return
	}
	// 锁定玩家
	win, bet, round := int64(0), int64(0), int32(0)
	if oldSess != nil && oldSess.Role != nil {
		win, bet, round = sess.TotalWin, sess.TotalBet, sess.TotalRound
	}
	user, err := sess.LockRoom(uid, win, bet, round)
	if err != nil {
		// 发送错误消息
		log.Debugf("id:%v,uid:%v,kid:%v,room:%v,登录失败:%v", sess.AgentId, uid, KindId, RoomId, err.Error())
		sess.SendError(int32(protocol.MsgId_LoginRoomReq), 1000, "登录失败", err.Error())
		sess.Close()
		return
	}
	if user == nil {
		log.Debugf("id:%v,uid:%v,kid:%v,room:%v,登录失败2", sess.AgentId, uid, KindId, RoomId)
		sess.SendError(int32(protocol.MsgId_LoginRoomReq), 1000, "登录失败2", "")
		sess.Close()
		return
	}

	sess.UserId = uid
	if oldSess != nil {
		sess.Role = oldSess.Role
		oldSess.Disposed = true
		oldSess.UserId = 0
		AddUser(sess)
		hall.UserReline(oldSess, sess)
		// 发送错误消息,顶掉玩家
		oldSess.Close()
	} else {
		// 检查房间配置
		if Config.Lock > 0{
			log.Debugf("id:%v,uid:%v,kid:%v,room:%v,登录失败:%v", sess.AgentId, uid, KindId, RoomId, "房间已锁定")
			sess.SendError(int32(protocol.MsgId_LoginRoomReq), 1000, "房间已锁定", "")
			sess.Close()
			return
		}

		// 检查所带金币
		coin := user.Bag[CoinKey]
		if coin < Config.DoorMin || coin > Config.DoorMax {
			log.Debugf("id:%v,uid:%v,kid:%v,room:%v,登录失败:%v", sess.AgentId, uid, KindId, RoomId, "金币不足")
			sess.SendError(int32(protocol.MsgId_LoginRoomReq), 1000, "金币不足", "")
			sess.Close()
			return
		}
		AddUser(sess)
		hall.UserOnline(sess, user, coin)
	}
}

// 主线程调用，玩家下线
func userOffline(sess *Session) {
	log.Debugf("userOffline:%v", sess)
	hall.UserOffline(sess)
}

