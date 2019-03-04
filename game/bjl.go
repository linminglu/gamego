package main

import (
	"math/rand"
	"time"

	log "github.com/sirupsen/logrus"

	"local.com/abc/game/model"
	"local.com/abc/game/msg"
	"local.com/abc/game/room"
)

var(
	// 龙虎点数映射表
	bjlPoint = [64]byte{}
)
func init() {
	bjlPoint[model.AA], bjlPoint[model.BA], bjlPoint[model.CA], bjlPoint[model.DA] = 1, 1, 1, 1
	bjlPoint[model.A2], bjlPoint[model.B2], bjlPoint[model.C2], bjlPoint[model.D2] = 2, 2, 2, 2
	bjlPoint[model.A3], bjlPoint[model.B3], bjlPoint[model.C3], bjlPoint[model.D3] = 3, 3, 3, 3
	bjlPoint[model.A4], bjlPoint[model.B4], bjlPoint[model.C4], bjlPoint[model.D4] = 4, 4, 4, 4
	bjlPoint[model.A5], bjlPoint[model.B5], bjlPoint[model.C5], bjlPoint[model.D5] = 5, 5, 5, 5
	bjlPoint[model.A6], bjlPoint[model.B6], bjlPoint[model.C6], bjlPoint[model.D6] = 6, 6, 6, 6
	bjlPoint[model.A7], bjlPoint[model.B7], bjlPoint[model.C7], bjlPoint[model.D7] = 7, 7, 7, 7
	bjlPoint[model.A8], bjlPoint[model.B8], bjlPoint[model.C8], bjlPoint[model.D8] = 8, 8, 8, 8
	bjlPoint[model.A9], bjlPoint[model.B9], bjlPoint[model.C9], bjlPoint[model.D9] = 9, 9, 9, 9
	bjlPoint[model.A10], bjlPoint[model.B10], bjlPoint[model.C10], bjlPoint[model.D10] = 10, 10, 10, 10
	bjlPoint[model.AJ], bjlPoint[model.BJ], bjlPoint[model.CJ], bjlPoint[model.DJ] = 20, 20, 20, 20
	bjlPoint[model.AQ], bjlPoint[model.BQ], bjlPoint[model.CQ], bjlPoint[model.DQ] = 30, 30, 30, 30
	bjlPoint[model.AK], bjlPoint[model.BK], bjlPoint[model.CK], bjlPoint[model.DK] = 40, 40, 40, 40
}

// 百家乐
type BjlDealer struct {
	i      int32
	Poker  []byte //所有的牌
	Offset int    //牌的位置
}

var(
	// 执行步骤
	bjlSchedule = []Plan{
		{f: gameReady, d: second},
		{f: bjlOpen, d: time.Microsecond},
		{f: bjlBet, d: second},
		{f: bjlBet, d: second},
		{f: bjlBet, d: second},
		{f: bjlBet, d: second},
		{f: bjlBet, d: second},
		{f: bjlBet, d: second},
		{f: bjlBet, d: second},
		{f: bjlBet, d: second},
		{f: bjlBet, d: second},
		{f: bjlBet, d: second},
		{f: bjlClose, d: second},
		{f: bjlDeal, d: 2*second},
	}

	// 百家乐税率千分比
	bjlTaxs = []int64{50,50,50,50,50}
)

func NewBjlDealer(config *model.RoomInfo) *BjlDealer {
	d := &BjlDealer{
	}
	return d
}

// 开始
func bjlOpen (table *Table){
	// 发送开始下注消息给所有玩家
	table.State = 2
	log.Debugf("百家乐开始下注:%v", table.CurId)
}

func bjlBet(table *Table) {
	for _, role := range table.Roles {
		if role.Session == nil && (role.Id%5) == rand.Int31n(5) {
			bet := msg.BetReq{
				Item: rand.Int31n(31),
				Coin: 100 + rand.Int31n(100)*100,
			}
			if role.RobotCanBet(bet.Item) {
				role.AddBet(bet)
				//log.Debugf("机器人下注:%v,%v", bet, r)
			}
		}
	}
}

// 停止下注
func bjlClose (table *Table) {
	table.State = 3
	log.Debugf("停止下注:%v", table.CurId)
}

// 发牌结算
func bjlDeal(table *Table){
	table.State = 4
	log.Debugf("发牌结算:%v", table.CurId)
	// 发牌结算
	table.dealer.Deal(table)
}

func(this *BjlDealer) Schedule()[]Plan{
	return bjlSchedule
}

func getBjlPoint(a []byte)byte {
	return ((bjlPoint[a[0]]) + (bjlPoint[a[1]]) + (bjlPoint[a[2]])) % 10
}

func (this *BjlDealer) Deal(table *Table) {
	// 检查剩余牌数量
	offset := this.Offset
	if offset >= len(this.Poker)*2/3 {
		log.Debugf("重新洗牌:%v", this.i)
		this.Poker = model.NewPoker(8, false, true)
		offset = 0
	}
	// 闲庄先各取2张牌
	a := []byte{this.Poker[offset], this.Poker[offset+2], 0}
	b := []byte{this.Poker[offset+1], this.Poker[offset+3], 0}
	offset += 4

	pointA := getBjlPoint(a)
	pointB := getBjlPoint(b)

	// 检查是否补牌
	if pointA >= 8 || pointB >= 8 || (pointA >=6 && pointB >= 6){
		//任何一家拿到9点（天生赢家），牌局就算结束，不再补牌
	} else {
		//闲家0-5必须博牌
		addA := false
		if pointA < 6 {
			addA = true
			a[2] = this.Poker[offset]
			offset++
		}

		addB := false
		//庄家0-2必须博牌
		if pointB < 3 {
			addB = true
		} else {
			aa := byte(255)
			if addA {
				aa = bjlPoint[a[2]]
			}
			switch pointB {
			case 3:
				// 如果闲家补得第三张牌（非三张牌点数相加，下同）是8点，不须补牌，其他则需补牌
				addB = aa != 8
			case 4:
				// 如果闲家补得第三张牌是0,1,8,9点，不须补牌，其他则需补牌
				addB = (aa != 0) && (aa != 1) && (aa != 8) && (aa != 9)
			case 5:
				// 如果闲家补得第三张牌是0,1,2,3,8,9点，不须补牌，其他则需补牌
				addB = (aa != 0) && (aa != 1) && (aa != 2) && (aa != 3) && (aa != 8) && (aa != 9)
			case 6:
				// 如果闲家需补牌（即前提是闲家为1至5点）而补得第三张牌是6或7点，补一张牌，其他则不需补牌
				addB = addA && (aa == 6 || aa == 7)
			}
		}

		if addB {
			b[2] = this.Poker[offset]
			offset++
		}
	}
	this.Offset = offset

	note := model.PokerArrayString(a) + "|" + model.PokerArrayString(b)
	round := table.round
	round.Odds = bjlPk(a, b)
	round.Poker = []byte{a[0], a[1], a[2], b[0], b[1], b[2]}
	round.Note = note
	// log.Debugf("发牌:%v,%v", note, round.Odds)

	for _, role := range table.Roles {
		if flow := role.Balance(bjlTaxs); flow != nil {
			room.WriteCoin(flow)
			if role.Session != nil {
				log.Debugf("结算:%v", flow)
			}
		}
	}
	// 结算结果发给玩家
	table.LastId = table.CurId
	round.End = room.Now()
	room.SaveLog(round)
}

func bjlPk(a []byte, b []byte) (odds []int32) {
	pa := getBjlPoint(a)
	pb := getBjlPoint(b)
	// 可下注的选项数量(0:闲赢,1:庄赢,2:和,3:闲对,4:庄对)
	odds = []int32{lostRadix, lostRadix, lostRadix, lostRadix, lostRadix}
	if pa > pb {
		// 闲赢
		odds[0] = 1*radix + radix
	} else if pa < pb {
		// 庄赢
		odds[1] = 1*radix + radix
	} else {
		// 和
		odds[2] = 8*radix + radix
	}
	//闲对
	if (bjlPoint[a[0]]) == (bjlPoint[a[1]]) {
		odds[3] = 11*radix + radix
	}
	//庄对
	if (bjlPoint[b[0]]) == (bjlPoint[b[1]]) {
		odds[4] = 11*radix + radix
	}
	return odds
}

func (this *BjlDealer) BetItem() int{
	// 可下注的选项数量(0:闲赢,1:庄赢,2:和,3:闲对,4:庄对)
	return 5
}