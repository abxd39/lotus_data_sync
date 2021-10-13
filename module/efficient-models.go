package module

import (
	"encoding/json"
	"lotus_data_sync/utils"
	"fmt"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

//单台效率
type AboutMachineEfficient struct {
	Addr        string `bson:"addr,omitempty"`
	Cid         string `bson:"cid,omitempty"`
	Method      int    `bson:"method,omitempty"`
	Timestamp   int64  `bson:"timestamp,omitempty"`
	SectorCount int    `bson:"sector_count,omitempty"`
}

const (
	AboutMachineEfficientCollection = "efficient_New"
)

func CreateAboutMachineEfficientIndex() {
	ms, c := Connect(AboutMachineEfficientCollection)
	defer ms.Close()
	ms.SetMode(mgo.Monotonic, true)

	indexs := []mgo.Index{
		{Key: []string{"cid"}, Unique: true, Background: true},
		{Key: []string{"timestamp"}, Unique: false, Background: true},
		{Key: []string{"addr"}, Unique: false, Background: true},
	}
	for _, index := range indexs {
		if err := c.EnsureIndex(index); err != nil {
			panic(err)
		}
	}
}

func GetAboutMachineEfficientCount(param AboutMachineEfficient) (int, int) {
	pb, _ := json.Marshal(&param)

	end := param.Timestamp
	start := end - 6*3600
	q := bson.M{}
	q["timestamp"] = bson.M{"$gte": start, "$lt": end}
	if param.Method != 0 {
		q["method"] = param.Method
	}
	if param.Addr != "" {
		q["addr"] = param.Addr
	}

	res := make([]AboutMachineEfficient, 0)

	err := FindAll(AboutMachineEfficientCollection, q, nil, &res)
	if err != nil {
		utils.Log.Errorln(err)
		return 0, 0
	}
	totalCount := 0
	AggregateCount := 0

	for _, v := range res {
		totalCount += 1
		AggregateCount += v.SectorCount
	}
	str := fmt.Sprintf("totalCount=%d  AggregateCount=%d", totalCount, AggregateCount)
	utils.Log.Traceln(string(pb), str)
	return totalCount, AggregateCount

}

func AboutMachineEfficientUpsert(param AboutMachineEfficient) error {
	if _, err := Upsert(AboutMachineEfficientCollection, bson.M{"cid": param.Cid}, &param); err != nil {
		utils.Log.Errorln(err)
		return err
	}
	return nil
}

//这个函数有问题
func GetAboutMachineEfficientCountNew(param AboutMachineEfficient) (int, int) {
	end := param.Timestamp
	start := end - 6*3600
	qPipe := []bson.M{
		{"$match": bson.M{"timestamp": bson.M{"$gte": start, "$lt": end}}},
		{"$group": bson.M{"_id": "", "count": bson.M{"$sum": 1}, "aggregate_count": bson.M{"$sum": "$sector_count"}}},
		{"$project": bson.M{"_id": 0, "count": 1, "aggregate_count": 1}},
	}
	if param.Method != 0 {
		qPipe = append(qPipe, bson.M{"$match": bson.M{"method": param.Method}})
	}
	if param.Addr != "" {
		qPipe = append(qPipe, bson.M{"$match": bson.M{"addr": param.Addr}})
	}
	type result struct {
		Count          int `bson:"count"`
		AggregateCount int `bson:"aggregate_count"`
	}
	var res result

	err := AggregateOne(AboutMachineEfficientCollection, qPipe, &res)
	if err != nil {
		utils.Log.Errorln(err)
		return 0, 0
	}
	utils.Log.Traceln("聚合结果为", res.Count, "   ", res.AggregateCount)
	return 0, 0
}
