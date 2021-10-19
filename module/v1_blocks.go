package module

import (
	"context"
	"lotus_data_sync/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func CreateBlockIndex() {
	indexView := utils.Mdb.Collection(BlocksCollection).Indexes()
	models := []mongo.IndexModel{
		{
			Keys: bson.D{{"cid", 1}},
			//Options: options.Index().SetName("nameAge"),
			Options: options.Index().SetUnique(true).SetBackground(true),
		},

		{
			Keys:    bson.D{{"block_header.Height", -1}},
			Options: options.Index().SetBackground(true).SetUnique(false),
		},
		{
			Keys:    bson.D{{"block_header.Timestamp", -1}},
			Options: options.Index().SetBackground(true).SetUnique(false),
		},
	}
	opts := options.CreateIndexes().SetMaxTime(2 * time.Second)
	names, err := indexView.CreateMany(context.TODO(), models, opts)
	if err != nil {
		panic(err)
	}

	utils.Log.Tracef("created indexes %v\n", names)

}
func (f *FilscanBlock) InsertMany(param *FilscanBlock) error {
	doc := utils.ToInterface(param)
	res, err := utils.Mdb.Collection(BlocksCollection).InsertOne(context.TODO(), doc)
	if err != nil {
		//utils.Log.Errorln(err)
		//utils.Log.Errorf("block_cid=%s height=%d",param.BlockCid,param.Height)
		return err
	}
	utils.Log.Tracef("%+v", res)
	return nil
}
