package module

import (
	"encoding/json"
	"lotus_data_sync/utils"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"time"
)

type Gas struct {
	Miner               string `bson:"miner" json:"miner"`
	CreateTotalGas      string `bson:"create_gas" json:"create_total_gas"`                 //生产Gas 总消耗
	IncreasePower       string `bson:"increase_power" json:"increase_power"`               //单日封装量
	IncreasePowerOffset string `bson:"increase_power_offset" json:"increase_power_offset"` //单日算力增量
	PledgeGas           string `bson:"pledge_gas"json:"pledge_gas"`                        //扇区质押
	WinGas              string `bson:"win_gas" json:"win_gas"`                             //维护gas 消耗
	Date                string `bson:"date" json:"date"`                                   //日期
	Created             int64  `bson:"created" json:"created"`
	Updated             int64  `bson:"updated" json:"updated"`
}

const (
	GasCollection = "gas"
)

func CreateGasIndex() {
	ms, c := Connect(GasCollection)
	defer ms.Close()
	ms.SetMode(mgo.Monotonic, true)

	indexs := []mgo.Index{
		{Key: []string{"miner"}, Unique: false, Background: true},
		{Key: []string{"date"}, Unique: false, Background: true},
	}
	for _, index := range indexs {
		if err := c.EnsureIndex(index); err != nil {
			panic(err)
		}
	}
}

func UpsertGas(gas *Gas) (err error) {
	//gas.Date = time.Now().Format(utils.TimeDate)
	gas.Date = "2021-08-02"
	gas.Updated = time.Now().Unix()
	gas.Created = gas.Updated
	tbyte, _ := json.Marshal(gas)
	var p interface{}
	err = json.Unmarshal(tbyte, &p)
	if err != nil {
		utils.Log.Errorln(err)
		return err
	}
	selector := bson.M{"miner": gas.Miner, "date": gas.Date}
	_, err = Upsert(GasCollection, selector, p)
	return
}
