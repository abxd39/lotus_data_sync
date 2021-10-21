package module

import (
	"context"
	"lotus_data_sync/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DayMiner struct {
	Reward    float64 `json:"reward"`
	Blocks    int64   `json:"blocks"`
	Timestamp int64   `json:"timestamp"` // 日期 一天一条
	GasUsage  float64 `json:"gas_usage"`
}

type Miner struct {
	List         []DayMiner `json:"list"`
	PeerId       string     `json:"peer_id"`
	Tag          string     `json:"tag"`
	SectorSize   int32      `json:"sector_size"`
	Addr         string     `json:"addr"`
	TotalRewards float64    `json:"total_rewards"`
	TotalBlock   int64      `json:"total_block"`
	TotalGas     float64    `json:"total_gas"`
	Timestamp    int64      `json:"timestamp"`
	Height       int64      `json:"height"`
	Owner        string     `json:"owner"`
	Worker       string     `json:"worker"`
}

type MinerCache struct {
	Owner        string   `json:"owner"`
	Worker       string   `json:"worker"`
	List         DayMiner `json:"list"`
	PeerId       string   `json:"peer_id"`
	Tag          string   `json:"tag"`
	SectorSize   string   `json:"sector_size"`
	Addr         string   `json:"addr"`
	TotalRewards float64  `json:"total_rewards"`
	TotalBlock   int64    `json:"total_block"`
	TotalGas     int64    `json:"total_gas"`
	Timestamp    int64    `json:"timestamp"`
	Height       int64    `json:"height"`
}

const (
	MinerCollection = "miner"
)

func CreateMinerIndex() {
	indexView := utils.Mdb.Collection(MinerCollection).Indexes()

	// Create two indexes: {name: 1, email: 1} and {name: 1, age: 1}
	// For the first index, specify no options. The name will be generated as
	// "name_1_email_1" by the driver.
	// For the second index, specify the Name option to explicitly set the name
	// to "nameAge".
	models := []mongo.IndexModel{
		{
			Keys: bson.D{{"peer_id", 1}, {"addr", 1}},
			//Options: options.Index().SetName("nameAge"),
			Options: options.Index().SetUnique(true).SetBackground(true),
		},
		{
			Keys:    bson.D{{"total_rewards", -1}, {"total_block", -1}, {"total_gas", -1}},
			Options: options.Index().SetBackground(true).SetUnique(false),
		},
		{
			Keys:    bson.D{{"timestamp", -1}},
			Options: options.Index().SetBackground(true).SetUnique(false),
		},
	}

	// Specify the MaxTime option to limit the amount of time the operation can
	// run on the server
	opts := options.CreateIndexes().SetMaxTime(2 * time.Second)
	names, err := indexView.CreateMany(context.TODO(), models, opts)
	if err != nil {
		panic(err)
	}

	utils.Log.Tracef("created indexes %v\n", names)
}

func (m *Miner) Upsert(param MinerCache) error {
	// utils.Mdb.Collection(MinerCollection).BulkWrite()
	// today := time.Now().Format("01-01-1970")
	filter := bson.D{{"addr", param.Addr}}
	update := bson.D{{"$inc", bson.D{{"total_rewards", param.TotalRewards}, {"total_block", param.TotalBlock}, {"total_gas", param.TotalGas}}}, {"$set", bson.D{{"timestamp", param.Timestamp},
		{"height", param.Height}, {"sector_size", param.SectorSize}, {"tag", param.Tag}, {"peer_id", param.PeerId}, {"owner", param.Owner}, {"worker", param.Worker}}}, {"$push", bson.D{{"list", param.List}}}}
	options := options.Update().SetUpsert(true)

	result, err := utils.Mdb.Collection(MinerCollection).UpdateOne(context.TODO(), filter, update, options)
	if err != nil {
		utils.Log.Errorln(err)
		return err
	}

	if result.MatchedCount != 0 {
		utils.Log.Traceln("matched and replaced an existing document")
	}

	return nil
}
