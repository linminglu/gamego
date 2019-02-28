// +build !mgo

package mongodb

import (
	"time"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo/options"

	"local.com/abc/game/util"
)

var (
	returnNew = options.FindOneAndUpdate().SetReturnDocument(options.After)
	upNew1    = options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	upsert2   = options.Update().SetUpsert(true)
)

func (d *driver) GetIncrementKey(key string, n int64) int64 {
	id := bson.D{{"_id", key}}
	up := bson.D{{"$inc", bson.D{{"n", n}}}}
	doc := incrementKeyDoc{}
	if err := d.gameConf.FindOneAndUpdate(d.ctx, id, up, upNew1).Decode(&doc); err == nil {
		return doc.N
	}
	return 0
}

func (d *driver) GetConfigValue(id string) (int64, error) {
	doc := incrementKeyDoc{}
	if err := d.gameConf.FindOne(d.ctx, bson.D{{"_id", id}}).Decode(&doc); err == nil {
		return doc.N, nil
	} else {
		return 0, err
	}
}

func (d *driver) SetConfigValue(id string, v interface{}) error {
	up := bson.D{{"$set", bson.D{{"n", v}}}}
	_, err := d.gameConf.UpdateOne(d.ctx, bson.D{{"_id", id}}, up, upsert2)
	return err
}

//初始化纪元时间
var idEpoch = bson.D{{"_id", "epoch"}}

func (d *driver) initEpoch() {
	up := bson.D{{"$setOnInsert", bson.D{{"n", util.Now()}}}}
	doc := incrementKeyDoc{}
	if err := d.conf.FindOneAndUpdate(d.ctx, idEpoch, up, upNew1).Decode(&doc); err == nil {
		d.epoch = int32(doc.N)
	}
}

var idUser = bson.D{{"_id", CollUser}}

func (d *driver) initUserId() {
	up := bson.D{{"$setOnInsert", bson.D{{"n", int64(1000000)}}}}
	doc := incrementKeyDoc{}
	if err := d.conf.FindOneAndUpdate(d.ctx, idUser, up, upNew1).Decode(&doc); err == nil {
		//d.epoch = int32(doc.N)
	}
}

// 新的序列号
func (d *driver) NewSN(key interface{}, count int64) int64 {
	up := bson.D{{"$inc", bson.D{{"n", count}}}}
	doc := incrementKeyDoc{}
	if err := d.conf.FindOneAndUpdate(d.ctx, bson.D{{"_id", key}}, up, upNew1).Decode(&doc); err == nil {
		if start := (doc.N - count) + 1; start > 0 {
			//log.Debugf("NewSN:%v,%v,%v", key, count, start)
			return start
		}
	}
	//log.Debugf("NewSN error:%v,%v,%v", key, count, 0)
	return 0
}

var idNow = bson.D{{"_id", "now"}}

func (d *driver) Now() time.Time {
	var t UpTime
	if e := d.conf.FindOneAndUpdate(d.ctx, idNow, upNowChange, upNew1).Decode(&t); e == nil {
		return t.Up.Local()
	}
	return time.Now().Local()
}