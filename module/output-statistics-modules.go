package module

import (
	"encoding/json"
	"lotus_data_sync/utils"
	"fmt"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/shopspring/decimal"
	"time"
)

type MinerSet struct {
	Miner   string `json:"miner" bson:"miner"`
	Created string `json:"created" bson:"created"`
}

//一天一个矿工一条记录
type OutputStatistics struct {
	Data        string `json:"data" bson:"data"`
	Created     int64  `json:"created" bson:"created"`
	Updated     int64  `json:"updated" bson:"updated"`
	Miner       string `json:"miner" json:"miner"`
	WinCount    int64  `json:"win_count" bson:"win_count"`       //出块份数
	BlockReward string `json:"block_reward" bson:"block_reward"` //出块奖励
	PowerOffset string `json:"power_offset" json:"power_offset"` //算力增量
	Pledge      string `json:"pledge" bson:"pledge"`
}

//一个矿工一条记录
type StatisticsPower struct {
	Miner            string `json:"miner" json:"miner"`
	WinCountTotal    int64  `json:"win_count_total" json:"win_count_total"`       //累计出块份数
	BlockRewardTotal string `json:"block_reward_total" json:"block_reward_total"` //累计出块奖励
	//PledgeTotal      string `json:"pledge_total" bson:"pledge_total"`             //累计质押
	MinerPower    Claim  `json:"miner_power" bson:"miner_power"`
	TotalPower    Claim  `json:"total_power" bson:"total_power"`
	Created       int64  `json:"created" bson:"created"`
	Updated       int64  `json:"updated" bson:"updated"`
	Balance       string `json:"balance" bson:"balance"`               //账户余额
	Available     string `json:"available" bson:"available"`           //账户可用余额
	SectorsPledge string `json:"sectors_pledge" json:"sectors_pledge"` //扇区质押
	LockedFunds   string `json:"locked_funds" json:"locked_funds"`     //锁仓奖励
}

type Claim struct {
	// Sum of raw byte power for a miner's sectors.
	RawBytePower int64 `json:"raw_byte_power" bson:"raw_byte_power"`
	// Sum of quality adjusted power for a miner's sectors.
	QualityAdjPower int64 `json:"quality_adj_power" bson:"quality_adj_power"`
}

const (
	MinerAddressCollection    = "miner_set"
	OutputCollection          = "output_statistics"
	StatisticsPowerCollection = "statistics_power"
)

func CreateStatisticsPowerIndex() {
	ms, c := Connect(StatisticsPowerCollection)
	defer ms.Close()
	ms.SetMode(mgo.Monotonic, true)

	indexs := []mgo.Index{
		{Key: []string{"miner"}, Unique: false, Background: true},
		//{Key: []string{"date"}, Unique: false, Background: true},
	}
	for _, index := range indexs {
		if err := c.EnsureIndex(index); err != nil {
			panic(err)
		}
	}
}

func CreateOutputIndex() {
	ms, c := Connect(OutputCollection)
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
func CreateMinerAddress() {
	ms, c := Connect(MinerAddressCollection)
	defer ms.Close()
	ms.SetMode(mgo.Monotonic, true)

	indexs := []mgo.Index{
		{Key: []string{"miner"}, Unique: true, Background: true},
	}
	for _, index := range indexs {
		if err := c.EnsureIndex(index); err != nil {
			panic(err)
		}
	}
}

func UpsetMinerAddress(miner MinerSet) error {
	tbyte, _ := json.Marshal(miner)
	var p interface{}
	err := json.Unmarshal(tbyte, &p)
	if err != nil {
		utils.Log.Errorln(err)
		return err
	}
	selector := bson.M{"miner": miner}
	if _, err = Upsert(StatisticsPowerCollection, selector, p); err != nil {
		utils.Log.Errorln(err)
		return err
	}
	return nil
}

func UpsertStatistics(power *StatisticsPower) (err error) {
	tbyte, _ := json.Marshal(power)
	var p interface{}
	err = json.Unmarshal(tbyte, &p)
	if err != nil {
		utils.Log.Errorln(err)
		return err
	}
	selector := bson.M{"miner": power.Miner}
	if _, err = Upsert(StatisticsPowerCollection, selector, p); err != nil {
		utils.Log.Errorln(err)
		return err
	}
	return
}

func UpdateStatistics(miner, reward string, wincount int64) error {
	var rewardTotal string
	//TODO 把遇到的所有的矿工地址都保留下来
	result := new(StatisticsPower)
	if err := GetStatisticsByMiner(miner, result); err != nil { //没找到对应的矿工 直接返回。
		err = fmt.Errorf("miner =%v %s", miner, err.Error())
		//utils.Log.Errorln(err)
		var mineset MinerSet
		mineset.Miner = miner
		mineset.Created = time.Now().Format(utils.TimeString)
		MinerSetUpsert(&mineset)
		return err
	}
	if result.BlockRewardTotal == "" {
		result.BlockRewardTotal = "0"
	}
	dec, err := decimal.NewFromString(result.BlockRewardTotal)
	if err != nil {
		utils.Log.Errorln(err)
		return err
	}
	if reward == "" {
		reward = "0"
	}
	dec1, err := decimal.NewFromString(reward)
	if err != nil {
		utils.Log.Errorln(err)
		return err
	}
	rewardTotal = dec.Add(dec1).String()

	upsert_pairs := make([]interface{}, 4)
	upsert_pairs[0] = bson.M{"miner": miner}
	upsert_pairs[1] = bson.M{"block_reward_total": rewardTotal}
	upsert_pairs[2] = bson.M{"miner": miner}
	upsert_pairs[3] = bson.M{"$inc": bson.M{"win_count_total": wincount}}
	res, err := BulkUpdate(StatisticsPowerCollection, upsert_pairs)
	if err != nil {
		utils.Log.Errorln(err)
		return err
	}
	if res.Matched == 0 {
		return fmt.Errorf("matched==0 miner=%v", miner)
	}
	if res.Modified == 0 {
		return fmt.Errorf("modified==0 miner=%v", miner)
	}
	//	utils.Log.Tracef("%+v", res)
	return nil
}

func GetStatisticsByMiner(miner string, result *StatisticsPower) error {
	q := bson.M{"miner": miner}
	return FindOne(StatisticsPowerCollection, q, nil, result)
}

func OutputUpsert(param OutputStatistics) error {
	m := bson.M{"date": param.Data, "miner": param.Miner}
	tbyte, _ := json.Marshal(param)
	var p interface{}
	err := json.Unmarshal(tbyte, &p)
	if err != nil {
		return err
	}
	_, err = Upsert(OutputCollection, m, p)
	if err != nil {
		utils.Log.Errorln(err)
		return err
	}
	//utils.Log.Tracef("%+v", changeInfo)
	return nil
}

func MinerSetUpsert(param *MinerSet) error {
	//ok ==true 表示已经存在不需要入库
	// 给mongodb 减少io 压力
	if _, ok := utils.MinerSet.Load(param.Miner); ok {
		//utils.Log.Traceln(param.Miner, " 已经存在于数据裤中无需插入。。。")
		return nil
	}
	utils.MinerSet.Store(param.Miner, param.Created)
	m := bson.M{"miner": param.Miner}
	tbyte, _ := json.Marshal(param)
	var p interface{}
	err := json.Unmarshal(tbyte, &p)
	if err != nil {
		return err
	}
	_, err = Upsert(MinerAddressCollection, m, p)
	if err != nil {
		utils.Log.Errorln(err)
		return err
	}
	//	utils.Log.Tracef("miner=%v %+v", param.Miner, changeInfo)
	return nil
}
