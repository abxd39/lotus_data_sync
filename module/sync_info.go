package module

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"lotus_data_sync/utils"
	"time"
)

type SyncInfo struct {
	BlockCid string `json:"block_cid,omitempty"`
	Height   int64  `json:"height,omitempty"`
	Created  int64  `json:"created,omitempty"`
}

const (
	SyncInfoCollection = "sync_info"
)

func CreateSyncInfoIndex() {
	indexView := utils.Mdb.Collection(SyncInfoCollection).Indexes()

	// Create two indexes: {name: 1, email: 1} and {name: 1, age: 1}
	// For the first index, specify no options. The name will be generated as
	// "name_1_email_1" by the driver.
	// For the second index, specify the Name option to explicitly set the name
	// to "nameAge".
	models := []mongo.IndexModel{
		{
			Keys: bson.D{{"block_cid", 1}},
			//Options: options.Index().SetName("nameAge"),
			Options: options.Index().SetUnique(true).SetBackground(true),
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

func (s *SyncInfo) InsertOne(param SyncInfo) error {
	doc := utils.ToInterface(param)
	res, err := utils.Mdb.Collection(SyncInfoCollection).InsertOne(context.TODO(), doc)
	if err != nil {
		//utils.Log.Errorln(err,"		block_cid=",param.BlockCid,"	height=",param.Heights)
		return err
	}
	utils.Log.Tracef("%+v", res)
	return nil
}

func (s *SyncInfo) ExistOfHeight(height int) bool {
	var id primitive.ObjectID
	opts := options.FindOne().SetSort(bson.D{{"height", 1}})
	var result bson.M
	err := utils.Mdb.Collection(SyncInfoCollection).FindOne(
		context.TODO(),
		bson.D{{"height", height}},
		opts,
	).Decode(&result)
	if err != nil {
		// ErrNoDocuments means that the filter did not match any documents in
		// the collection.
		if err == mongo.ErrNoDocuments {
			return false
		}
		utils.Log.Errorln(err)
		return false
	}
	//utils.Log.Errorln(result)
	utils.Log.Traceln(id.String())
	return true

}
