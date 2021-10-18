package main

import (
	"context"

	"fmt"
	"lotus_data_sync/module"
	"lotus_data_sync/syncer"
	"lotus_data_sync/utils"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/astaxie/beego/config"
)

var Inst = &syncer.Filscaner{}

func main() {
	err := fmt.Errorf("")
	config_file := "local.conf"

	utils.Initconf, err = config.NewConfig("ini", config_file)

	if err != nil {
		panic(err)
	}
	//日志
	utils.SetupLogger()
	//初始化mongodb
	//module.MongodbInit(utils.Initconf)
	syncer.MessagMap=make(map[int]map[string]*module.MessageInfo, 0)
	module.MongodbConnect()
	//初始化lotus
	syncer.LotusInit()
	//初始化实力
	syncer.NewInstance(context.TODO(),utils.LotusApi)
	//初始化缓存
	if err := Inst.Init(context.TODO(), utils.LotusApi); err != nil {
		utils.Log.Traceln("error ", err)
		panic(err)
	}
	

	Inst.Run()
	local := utils.Initconf.String("Local")

	utils.Log.Traceln("Init() ok , loacl=", local, len(local))
	localhost := utils.Initconf.String("httpport")
	utils.Log.Traceln(fmt.Sprintf("server will listen %s", localhost))
	if err:=http.ListenAndServe(localhost, nil) ;err!=nil{//🔥图服务
		fmt.Println(err)
	}
	fmt.Printf("server will listen %s 已经退出", localhost)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer func() {
		if err = utils.Mdb.Client().Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
}

func gracefullShutdown(server *http.Server, quit <-chan os.Signal, done chan<- bool) {
	<-quit
	utils.Log.Traceln("Server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	server.SetKeepAlivesEnabled(false)
	if err := server.Shutdown(ctx); err != nil {
		utils.Log.Errorf("Could not gracefully shutdown the server: %v\n", err)
	}
	close(done)
}
