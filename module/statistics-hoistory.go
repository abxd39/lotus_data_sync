package module

import (
	"encoding/json"
	"lotus_data_sync/utils"
	"fmt"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// 记录统计过的数据 统计到 表 StatisticsPower
type BlockHistory struct {
	Miner       string `json:"miner" bson:"miner"`
	BlockCid    string `json:"block_cid" bson:"block_cid"`
	BlockReword string `json:"block_reword" bson:"block_reword"`
	Date        string `json:"date" bson:"date"`
	Created     string `json:"created" bson:"created"`
}

type MessageHistory struct {
	MessageCid string `json:"message_cid" bson:"message_cid"`
	Date       string `json:"date" bson:"date"`
	Created    string `json:"created" bson:"created"`
}

const (
	BlockHistoryCollection   = "block_history"
	MessageHistoryCollection = "message_history"
)

func CreateHistoryBlockIndex() {
	ms, c := Connect(BlockHistoryCollection)
	defer ms.Close()
	ms.SetMode(mgo.Monotonic, true)

	indexs := []mgo.Index{
		{Key: []string{"block_cid"}, Unique: false, Background: true},
	}
	for _, index := range indexs {
		if err := c.EnsureIndex(index); err != nil {
			panic(err)
		}
	}
}
func CreateHistoryMessageIndex() {
	ms, c := Connect(MessageHistoryCollection)
	defer ms.Close()
	ms.SetMode(mgo.Monotonic, true)

	indexs := []mgo.Index{
		{Key: []string{"message_cid"}, Unique: false, Background: true},
	}
	for _, index := range indexs {
		if err := c.EnsureIndex(index); err != nil {
			panic(err)
		}
	}
}

func (BlockHistory) InsertOne(b *BlockHistory) error {
	if b.BlockCid == "" {
		return fmt.Errorf("block cid 为空！！")
	}
	if b.Miner == "" {
		return fmt.Errorf("miner 为空！！")
	}
	m := bson.M{"block_cid": b.BlockCid}
	count, err := FindCount(BlockHistoryCollection, m, nil)
	if err != nil {
		utils.Log.Errorln(err)
		return err
	}
	if count != 0 { //已经累计过了
		return fmt.Errorf("block_cid=%v 已经统计过了", b.BlockCid)
	}
	tbyte, _ := json.Marshal(b)
	var p interface{}
	err = json.Unmarshal(tbyte, &p)
	if err != nil {
		return err
	}

	return Insert(BlockHistoryCollection, p)
}
func (BlockHistory) Remove(block_cid string) {
	m := bson.M{"block_cid": block_cid}
	if err := Remove(BlockHistoryCollection, m); err != nil {
		utils.Log.Errorln(err)
	}
}
func (MessageHistory) InsertOne(m *MessageHistory) error {
	if m.MessageCid == "" {
		return fmt.Errorf("block cid 为空！！")
	}
	q := bson.M{"message_cid": m.MessageCid}
	count, err := FindCount(MessageHistoryCollection, q, nil)
	if err != nil {
		utils.Log.Errorln(err)
		return err
	}
	if count != 0 { //已经累计过了
		return fmt.Errorf("message_cid=%v 已经统计过了", m.MessageCid)
	}
	tbyte, _ := json.Marshal(m)
	var p interface{}
	err = json.Unmarshal(tbyte, &p)
	if err != nil {
		return err
	}

	return Insert(MessageHistoryCollection, p)
}
