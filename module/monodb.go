package module

import (
	"context"
	"lotus_data_sync/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func MongodbConnect() {
	mongoHost := utils.Initconf.String("mongoHost")
	mongoDB := utils.Initconf.String("mongoDB")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://"+mongoHost))
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
	ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err = client.Ping(ctx, readpref.Primary())
	utils.Mdb = client.Database(mongoDB)
	CreateMsgIndex()
	CreateBlockIndex()
}

func CreateMsgIndex() {

	indexView := utils.Mdb.Collection(MsgCollection).Indexes()

	// Create two indexes: {name: 1, email: 1} and {name: 1, age: 1}
	// For the first index, specify no options. The name will be generated as
	// "name_1_email_1" by the driver.
	// For the second index, specify the Name option to explicitly set the name
	// to "nameAge".
	models := []mongo.IndexModel{
		{
			Keys: bson.D{{"cid", "text"}},
			//Options: options.Index().SetName("nameAge"),
			Options: options.Index().SetUnique(true).SetBackground(true),
		},
		{
			Keys:    bson.D{{"message.From", "text"}, {"message.To", "text"}, {"message.Method", -1}},
			Options: options.Index().SetBackground(true).SetUnique(false),
		},
		{
			Keys:    bson.D{{"msg_create", -1}},
			Options: options.Index().SetBackground(true).SetUnique(false),
		},
		{
			Keys:    bson.D{{"block_cid", "text"}},
			Options: options.Index().SetBackground(true).SetUnique(false),
		},
		{
			Keys:    bson.D{{"height", "text"}},
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

func CreateBlockIndex() {

	indexView := utils.Mdb.Collection(BlocksCollection).Indexes()
	models := []mongo.IndexModel{
		{
			Keys: bson.D{{"cid", "text"}},
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

func CreateTipsetIndex() {
	//var indexView *mongo.IndexView
	indexView := utils.Mdb.Collection(TipSetCollection).Indexes()
	models := []mongo.IndexModel{
		{
			Keys: bson.D{{"key", "text"}},
			//Options: options.Index().SetName("nameAge"),
			Options: options.Index().SetUnique(true).SetBackground(true),
		},

		{
			Keys:    bson.D{{"height", -1}},
			Options: options.Index().SetBackground(true).SetUnique(false),
		},
		{
			Keys:    bson.D{{"gmt_create", -1}},
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
