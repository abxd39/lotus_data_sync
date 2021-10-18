package utils

import (
	bc "github.com/astaxie/beego/config"
	"github.com/filecoin-project/lotus/api"
	"github.com/go-redis/redis"
	"xorm.io/xorm"
		"xorm.io/xorm/log"
		"go.mongodb.org/mongo-driver/mongo"

)

//var beegoConf bc.Configer

//func GetConfiger() bc.Configer {
//	if beegoConf == nil {
//		var err error
//		if beegoConf, err = bc.NewConfig("ini", "conf/app.conf"); err != nil {
//			panic(err)
//		}
//	}
//	return beegoConf
//}

var LotusApi api.FullNode
var Rdb0 *redis.Client
var Rdb1 *redis.Client

var NotifyChan chan string

var Initconf bc.Configer

//https://www.wenjiangs.com/doc/ci3qi5ax
var DB *xorm.Engine

 var LoggerXorm *log.SimpleLogger 

 var Mdb *mongo.Database

 