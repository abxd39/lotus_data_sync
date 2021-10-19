package module

import (
	"context"
	"github.com/filecoin-project/go-state-types/abi"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"lotus_data_sync/utils"
	"time"
)

type BlockMsg struct {
	Msg      []*MessageInfo `json:"msg,omitempty"`
	Height   int64          `json:"height,omitempty"`
	BlockCid string         `json:"block_cid,omitempty"`
	Crated   int64          `json:"crated,omitempty"`
}

type MessageInfo struct {
	Cid     string `json:"cid,omitempty"`
	Version uint64 `json:"version,omitempty"`

	To   string `json:"to,omitempty"`
	From string `json:"from,omitempty"`

	Nonce uint64 `json:"nonce,omitempty"`

	Value int64 `json:"value,omitempty"`

	GasLimit   int64 `json:"gas_limit,omitempty"`
	GasFeeCap  int64 `json:"gas_fee_cap,omitempty"`
	GasPremium int64 `json:"gas_premium,omitempty"`

	Method abi.MethodNum `json:"method,omitempty"`
	Params []byte        `json:"params,omitempty"`
}

const (
	BlockMsgCollection = "BlockMsg"
)

func CreateBlockMsgIndex() {

	indexView := utils.Mdb.Collection(BlockMsgCollection).Indexes()

	// Create two indexes: {name: 1, email: 1} and {name: 1, age: 1}
	// For the first index, specify no options. The name will be generated as
	// "name_1_email_1" by the driver.
	// For the second index, specify the Name option to explicitly set the name
	// to "nameAge".
	models := []mongo.IndexModel{
		{
			Keys: bson.D{{"msg.cid", 1}},
			//Options: options.Index().SetName("nameAge"),
			Options: options.Index().SetUnique(true).SetBackground(true),
		},
		{
			Keys:    bson.D{{"msg.From", 1}, {"msg.To", 1}},
			Options: options.Index().SetBackground(true).SetUnique(false),
		},
		{
			Keys:    bson.D{{"created", -1}},
			Options: options.Index().SetBackground(true).SetUnique(false),
		},
		{
			Keys:    bson.D{{"block_cid", 1}},
			Options: options.Index().SetBackground(true).SetUnique(false),
		},
		{
			Keys:    bson.D{{"height", -1}},
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

func (b *BlockMsg) InsertMany(param BlockMsg) error {
	doc := utils.ToInterface(param)
	res, err := utils.Mdb.Collection(BlockMsgCollection).InsertOne(context.TODO(), doc)
	if err != nil {
	//	utils.Log.Errorln(err)
		//utils.Log.Errorf("block_cid=%s height=%d",param.BlockCid,param.Height)
		return err
	}
	utils.Log.Tracef("%+v", res)
	return nil
}
