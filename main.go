package main

import (
	"context"

	"fmt"
	"github.com/astaxie/beego/config"
	"lotus_data_sync/syncer"
	"lotus_data_sync/utils"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"
)

var Inst = &syncer.Filscaner{}

func main() {
	err := fmt.Errorf("")
	config_file := "local.conf"

	utils.Initconf, err = config.NewConfig("ini", config_file)

	if err != nil {
		panic(err)
	}
	//webLog := utils.Initconf.String("webLog")
	//if webLog == "" {
	//	panic("web log 路径配置错误")
	//}
	//logger.InitLog(webLog)
	//rawLogger, err := zap.NewDevelopment(zap.Fields(zap.String("serive", "lotus_filscan")))

	//if err != nil {
	//	panic(err)
	//}
	
		
	

	if err := Inst.Init(context.TODO(), config_file, utils.LotusApi); err != nil {
		utils.Log.Traceln("error ", err)
		panic(err)
	}
	//}()

	Inst.Run()
	local := utils.Initconf.String("Local")

	utils.Log.Traceln("Init() ok , loacl=", local, len(local))
	localhost:=utils.Initconf.String("httpport")
	http.ListenAndServe(localhost, nil) //🔥图服务

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
