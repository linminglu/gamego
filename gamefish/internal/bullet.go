package internal

import (
	"math"

	"local.com/abc/game/model"
	"local.com/abc/game/protocol/fish"
)

// 子弹
type Bullet struct {
	fish.Bullet   // 子弹
	Job   int32   // 区分机器人和真人
	Speed float64 // 子弹速度
	dx    float64
	dy    float64
}

func (bullet *Bullet) isRobot() bool {
	return bullet.Job == model.JobRobot
}

func (bullet *Bullet) isPlayer() bool {
	return bullet.Job == model.JobPlayer
}

func (bullet *Bullet) InitMove() {
	angle := bullet.Direction
	bullet.dx = math.Cos(angle - (math.Pi / 2))
	bullet.dy = math.Sin(angle - (math.Pi / 2))
}

func (bullet *Bullet) Move(ms float64) {
	se := float64(ms) / 1000
	x := bullet.X + (bullet.Speed * bullet.dx * se)
	y := bullet.Y + (bullet.Speed * bullet.dy * se)

	if x < 0 {
		x = 0 + (0 - x)
		bullet.dx = -bullet.dx
		bullet.Direction = -bullet.Direction
	}
	if x > DefaultWidth {
		x = DefaultWidth - (x - DefaultWidth)
		bullet.dx = -bullet.dx
		bullet.Direction = -bullet.Direction
	}
	if y < 0 {
		y = 0 + (0 - y)
		bullet.dy = -bullet.dy
		bullet.Direction = math.Pi - bullet.Direction
	}
	if y < DefaultHeight {
		y = DefaultHeight - (y - DefaultHeight)
		bullet.dy = -bullet.dy
		bullet.Direction = math.Pi - bullet.Direction
	}
	bullet.X = x
	bullet.Y = y
}

func (bullet *Bullet) HitTest(fish *Fish ) bool {
	if fish.State >= ObjectState_Dead {
		return false
	}
	if bullet.Fish != 0 && bullet.Fish != fish.Id {
		return false
	}
	for _, box := range fish.Boxes {
		dis := CalcDistance(box.X, box.Y, bullet.X, bullet.Y)
		if dis < box.Radio+20 {
			return true
		}
	}
	return false
}
