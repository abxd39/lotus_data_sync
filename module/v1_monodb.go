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
	if err != nil {
		panic(err)
	}
	ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		panic(err)
	}
	utils.Mdb = client.Database(mongoDB)
	CreateBlockMsgIndex()
	CreateSyncInfoIndex()
	CreateBlockIndex()
	go CheckConnect(context.TODO())
}

//mongodb 健康检测
func CheckConnect(ctx context.Context) {
	ping := time.NewTicker(time.Second * 10)
	notify := make(chan int)
	for {
		select {
		case _, ok := <-notify:
			{
				if !ok {
					//通道没有关闭
					MongodbConnect() //重新连接
				}
			}
		case <-ping.C:
			{
				//检测连接的健康状况
				ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
				if err := utils.Mdb.Client().Ping(ctx, readpref.Primary()); err != nil {
					utils.Log.Errorln(err)
					notify <- 1
				}
			
			}
		case <-ctx.Done():
			{
				ping.Stop()
				close(notify)
				return
			}
		}
	}

}

func CreateMsgIndex() {

	indexView := utils.Mdb.Collection(BlockMsgCollection).Indexes()

	// Create two indexes: {name: 1, email: 1} and {name: 1, age: 1}
	// For the first index, specify no options. The name will be generated as
	// "name_1_email_1" by the driver.
	// For the second index, specify the Name option to explicitly set the name
	// to "nameAge".
	models := []mongo.IndexModel{
		{
			Keys: bson.D{{"cid", 1}},
			//Options: options.Index().SetName("nameAge"),
			Options: options.Index().SetUnique(true).SetBackground(true),
		},
		{
			Keys:    bson.D{{"message.From", 1}, {"message.To", 1}, {"message.Method", -1}},
			Options: options.Index().SetBackground(true).SetUnique(false),
		},
		{
			Keys:    bson.D{{"msg_create", -1}},
			Options: options.Index().SetBackground(true).SetUnique(false),
		},
		{
			Keys:    bson.D{{"block_cid", 1}},
			Options: options.Index().SetBackground(true).SetUnique(false),
		},
		{
			Keys:    bson.D{{"height", 1}},
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
