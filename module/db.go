package module

import (
	"lotus_data_sync/utils"
	"fmt"
	"log"

	"github.com/astaxie/beego/config"

	"time"

	"github.com/globalsign/mgo"
)

var ps = fmt.Sprintf
var globalS *mgo.Session
var DB string
var TimeNow int64 = 0
var TimeNowStr string = "2006-01-02 15:04:05"

func TimenowInit() {
	go func() {
		for {
			tim_t := time.Now()
			TimeNow = tim_t.Unix()
			TimeNowStr = tim_t.Format("2006-01-02 15:04:05")
			time.Sleep(time.Second)
		}
	}()
	return
}

func DbInit(config config.Configer) {
	host := config.String("mongoHost")
	user := config.String("mongoUser")
	pass := config.String("mongoPass")
	mongoDB := config.String("mongoDB")
	mgosession := GetGlobalSession(host, user, pass, mongoDB)
	_, err := mgosession.DatabaseNames()
	if err != nil {
		fmt.Println("err:", err)
		panic(ps("mongoInit fail:%v", err))
	}
	DB = mongoDB
	//utils.Log.Traceln("不可能啊！！")
	if err = mgosession.Ping(); err != nil {
		panic(ps("mongoInit fail:%v", err))
	}
	//mgo.SetDebug(true)
	//mgo.SetLogger(utils.MongodbLog)
	CreateBlockIndex()
	CreateBlockRewardIndex()
	CreateTipsetIndex()
	CreatePeerIndex()
	CreateMsgIndex()
	CreateBlockIndex()
	CreateAccountIndex()
	CreateTipsetRewardsIndex()
	CreateMinerIndex()
	CreateTagIndex()
	CreateGasIndex()
	CreateOutputIndex()
	CreateStatisticsPowerIndex()
	CreateHistoryBlockIndex()
	CreateHistoryMessageIndex()
	CreateMinerAddress()
	CreateBlockTimeOutIndex()
	CreateMinerDeadLineIndex()
	CreateAboutMachineEfficientIndex()
}

func GetGlobalSession(host, user, pass, mongoDB string) *mgo.Session {
	dialInfo := &mgo.DialInfo{
		Addrs:    []string{host},
		Username: user,
		Password: pass,
		Database: mongoDB,
		//Source:    "filecoin",
		//Mechanism: "SCRAM-SHA-1",
		Timeout: time.Duration(60) * time.Second,
	}
	//mongodb://<username>:<password>@<server_address>:<port>/<database_name>
	//url := fmt.Sprintf("mongodb://%s:%s@%s/%s", user, pass, host, mongoDB)
	//log.Println(url)
	//s, err := mgo.DialWithTimeout(url, time.Duration(30)*time.Second)
	s, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		log.Fatalln("create mongodb session error ", err)
	}
	globalS = s

	//mongolog =*utils.Logger//(os.Stdout,"[mongo---]",logger.Ldate|logger.Ltime|logger.Lmicroseconds|logger.Lshortfile)
	//mongolog.Output(4,"")

	mgo.SetDebug(false)
	mgo.SetLogger(utils.MongodbLog)
	return s
}

func connect(collection string) (*mgo.Session, *mgo.Collection) {
	s := globalS.Copy()
	c := s.DB(DB).C(collection)
	return s, c
}
func Close() {
	globalS.Close()
}
func Copy() (*mgo.Session, *mgo.Database) {
	s := globalS.Copy()
	db := s.DB(DB)
	return s, db
}

func Connect(collection string) (*mgo.Session, *mgo.Collection) {
	return connect(collection)
}

func Insert(collection string, docs ...interface{}) error {
	ms, c := connect(collection)
	defer ms.Close()
	return c.Insert(docs...)
}

func BulkUpsert(c *mgo.Collection, c_name string, docs []interface{}) (*mgo.BulkResult, error) {
	if c == nil {
		var s *mgo.Session
		s, c = connect(c_name)
		defer func(session *mgo.Session) {
			session.Close()
		}(s)
	}
	if c == nil {
		return nil, fmt.Errorf("connection is nil")
	}

	bulk := c.Bulk()
	bulk.Upsert(docs[:]...)
	return bulk.Run()
}

func BulkUpdate(collection string, docs []interface{}) (*mgo.BulkResult, error) {
	ms, c := connect(collection)
	defer ms.Close()
	bulk := c.Bulk()
	bulk.Update(docs[:]...)
	return bulk.Run()
}

func Upsert(collection string, selector interface{}, docs interface{}) (info *mgo.ChangeInfo, err error) {
	ms, c := connect(collection)
	defer ms.Close()
	return c.Upsert(selector, docs)
}
func UpdateAll(collection string, docs []interface{}) (*mgo.BulkResult, error) {
	ms, c := connect(collection)
	defer ms.Close()
	bulk := c.Bulk()
	bulk.UpdateAll(docs[:]...)
	return bulk.Run()
}

func IsExist(collection string, query interface{}) bool {
	ms, c := connect(collection)
	defer ms.Close()
	count, _ := c.Find(query).Count()
	return count > 0
}

func FindOne(collection string, query, selector, result interface{}) error {
	ms, c := connect(collection)
	defer ms.Close()
	return c.Find(query).Select(selector).One(result)
}

func FindAll(collection string, query, selector, result interface{}) error {
	ms, c := connect(collection)
	defer ms.Close()
	return c.Find(query).Select(selector).All(result)
}
func FindSortLimit(collection, sort string, query, selector, result interface{}, begindex, count int) error {
	ms, c := connect(collection)
	defer ms.Close()
	return c.Find(query).Select(selector).Sort(sort).Limit(count).Skip(begindex).All(result)
}

func FindSortCollationLimit(collection, sort string, query, selector, result interface{}, begindex, count int, collation *mgo.Collation) error {
	ms, c := connect(collection)
	defer ms.Close()
	utils.Log.Traceln(c.Database.Name, c.FullName, c.Name)
	return c.Find(query).Select(selector).Collation(collation).Sort(sort).Limit(count).Skip(begindex).All(result)
}

func FindAllLimit(collection string, query, selector, result interface{}, begindex, count int) error {
	ms, c := connect(collection)
	defer ms.Close()
	return c.Find(query).Select(selector).Limit(count).Skip(begindex).All(result)
}

func FindCount(collection string, query, selector interface{}) (int, error) {
	ms, c := connect(collection)
	defer ms.Close()
	return c.Find(query).Select(selector).Count()

}

func Update(collection string, query, update interface{}) error {
	ms, c := connect(collection)
	defer ms.Close()
	return c.Update(query, update)
}

func Remove(collection string, query interface{}) error {
	ms, c := connect(collection)
	defer ms.Close()
	return c.Remove(query)
}

func Distinct(collection, distinctKey string, query, result interface{}) error {
	ms, c := connect(collection)
	defer ms.Close()
	return c.Find(query).Distinct(distinctKey, result)

}
func AggregateAll(collection string, query interface{}, result interface{}) error {
	ms, c := connect(collection)
	defer ms.Close()

	return c.Pipe(query).Iter().All(result)
}
func AggregateOne(collection string, query interface{}, result interface{}) error {
	ms, c := connect(collection)
	defer ms.Close()
	return c.Pipe(query).One(result)
}

func AggregateCount(collection string, query interface{}, result interface{}) error {
	ms, c := connect(collection)
	defer ms.Close()
	return c.Pipe(query).One(result)
}
