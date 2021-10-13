package module

import (
	"encoding/json"
	"lotus_data_sync/utils"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type MinerDeadLine struct {
	OpenTime        int64  `json:"open_time" bson:"open_time"` //转换成秒
	Miner           string `json:"miner" bson:"miner"`
	Deadline        uint64 `json:"deadline" bson:"deadline"`
	Sectors         uint64 `json:"sectors_no" bson:"sectors_no"` //
	Partitions      int    `json:"partitions" bson:"partitions"`
	Fault           uint64 `json:"fault" bson:"fault"`
	RecoverySectors uint64 `json:"recovery_sectors" bson:"recovery_sectors"` //恢复扇数
}

const (
	MinerDeadLineCollection = "miner_deadline"
)

func CreateMinerDeadLineIndex() {
	ms, c := Connect(MinerDeadLineCollection)
	defer ms.Close()
	ms.SetMode(mgo.Monotonic, true)

	indexs := []mgo.Index{
		//{Key: []string{"miner"}, Unique: true, Background: true},
		{Key: []string{"miner", "deadline"}, Unique: true, Background: true},
	}
	for _, index := range indexs {
		if err := c.EnsureIndex(index); err != nil {
			panic(err)
		}
	}
}

func (m *MinerDeadLine) Upsert() error {
	tbyte, _ := json.Marshal(m)
	var p interface{}
	err := json.Unmarshal(tbyte, &p)
	if err != nil {
		return err
	}
	selector := bson.M{"mine": m.Miner, "deadline": m.Deadline}
	info, err := Upsert(MinerDeadLineCollection, selector, p)
	if err != nil {
		utils.Log.Errorln(err)
		return err
	}
	utils.Log.Tracef("%+v", info)
	return nil
}

func (m *MinerDeadLine) IsExit() bool {
	selector := bson.M{"mine": m.Miner, "deadline": m.Deadline}
	return IsExist(MinerDeadLineCollection, selector)
}
func (m *MinerDeadLine) ReMove() error {
	selector := bson.M{"mine": m.Miner, "deadline": m.Deadline}
	return Remove(MinerDeadLineCollection, selector)
}

func (m *MinerDeadLine) FindOne(mine string, deadline int64) error {
	q := bson.M{"miner": mine, "deadline": deadline}
	if err := FindOne(MinerDeadLineCollection, q, nil, m); err != nil {
		utils.Log.Errorln(err)
		return err
	}
	return nil
}

//定时通知。如果到了15分钟内还发不完，下一个定时时间到了，在优化
func (m *MinerDeadLine) FindAll() ([]MinerDeadLine, error) {
	list := make([]MinerDeadLine, 0)
	if err := FindAll(MinerDeadLineCollection, nil, nil, &list); err != nil {
		utils.Log.Errorln(err)
		return nil, err
	}
	return list, nil
}
