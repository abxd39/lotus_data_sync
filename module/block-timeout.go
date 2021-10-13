package module

import (
	"encoding/json"
	"lotus_data_sync/utils"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type BlockTimeOut struct {
	Timestamp uint64 `json:"timestamp" bson:"timestamp"`
	Height    int64  `json:"height" bson:"height"`
	BlockCid  string `json:"block_cid" bson:"block_cid" `
	Miner     string `json:"miner" bson:"miner"`
	Tag       string `json:"tag" bson:"tag"`
	Created   int64  `json:"created" bson:"created"`
}

const BlockTimeoutCollection = "block_timeout"

func CreateBlockTimeOutIndex() {
	ms, c := Connect(BlockTimeoutCollection)
	defer ms.Close()
	ms.SetMode(mgo.Monotonic, true)

	indexs := []mgo.Index{
		{Key: []string{"timestamp"}, Unique: false, Background: true},
		{Key: []string{"block_cid"}, Unique: true, Background: true},
	}
	for _, index := range indexs {
		if err := c.EnsureIndex(index); err != nil {
			panic(err)
		}
	}
	param := BlockTimeOut{}
	param.BlockCid = utils.BlockTimeTag
	tbyte, _ := json.Marshal(param)
	var p interface{}
	err := json.Unmarshal(tbyte, &p)
	if err != nil {
		utils.Log.Errorln(err)
		return
	}
	Insert(BlockTimeoutCollection, p)
}

func BlockTimeOutUpsert(param BlockTimeOut) {
	tbyte, _ := json.Marshal(param)
	var p interface{}
	err := json.Unmarshal(tbyte, &p)
	if err != nil {
		utils.Log.Errorln(err)
		return
	}
	selector := bson.M{"block_cid": param.BlockCid}
	if _, err = Upsert(BlockTimeoutCollection, selector, p); err != nil {
		utils.Log.Errorln(err)
	}
	return
}

func GetHeight() uint64 {
	result := &BlockTimeOut{}
	query := bson.M{"block_cid": utils.BlockTimeTag}
	if err := FindOne(BlockTimeoutCollection, query, nil, result); err != nil {
		utils.Log.Errorln(err)
	}
	return uint64(result.Created)
}
