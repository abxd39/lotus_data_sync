package syncer

import (


	"lotus_data_sync/module"
	"lotus_data_sync/utils"
	"fmt"
	"github.com/filecoin-project/go-address"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"math/big"
	"sort"
	"time"
	"errors"
)

// https://docs.mongodb.com/manual/reference/operator/aggregation/first/#grp._S_first
// https://stackoverflow.com/questions/6498506/mongodb-select-the-top-n-rows-from-each-group
// https://stackoverflow.com/questions/34375163/how-to-use-mongodb-aggregate-to-get-the-first-of-each-group-including-nulls
// https://stackoverflow.com/questions/34325714/how-to-get-lastest-n-records-of-each-group-in-mongodb
// https://stackoverflow.com/questions/16409719/can-i-get-first-document-not-field-in-mongodb-aggregate-query

type ModelsBlockReward struct {
	Height          uint64             `bson:"height"`
	ReleasedRewards *module.BsonBigint `bson:"reward"`
}

func (fs *Filscaner) doUpsertMiners() error {

	if fs.toUpdateMinerIndex <= 0 {
		return nil
	}

	var offset uint64

	if fs.toUpdateMinerIndex >= fs.toUpdateMinerSize {
		offset = fs.toUpdateMinerIndex * 2
	}

	if err := module.BulkUpsertMiners(fs.toUpsertMiners[0:offset]); err != nil {
		return err
	}
	fs.toUpdateMinerIndex = 0

	return nil
}

func (fs *Filscaner) modelsUpdateMiner(miner *module.MinerStateAtTipset) error {
	var err error
	//utils.Log.Tracef("trace debug ------001")
	if fs.toUpdateMinerIndex >= fs.toUpdateMinerSize {
		if err = fs.doUpsertMiners(); err != nil {
			return err
		}
	}
	//TODO
	//utils.Log.Tracef("trace debug ------002")
	miner.GmtCreate = time.Now().Unix()
	miner.GmtModified = miner.GmtCreate
	offset := fs.toUpdateMinerIndex * 2
	fs.toUpsertMiners[offset] = bson.M{"miner_addr": miner.MinerAddr, "tipset_height": miner.TipsetHeight}
	var minerInfo module.MinerInfo
	minerInfo.Power = miner.Power.String()
	minerInfo.Worker = miner.Worker
	minerInfo.Owner = miner.Owner
	minerInfo.PeerId = miner.PeerId
	minerInfo.PowerPercent = miner.PowerPercent

	minerInfo.TotalPower = miner.TotalPower.String()
	minerInfo.TipsetHeight = miner.TipsetHeight
	minerInfo.MineTime = miner.MineTime
	minerInfo.MinerAddr = miner.MinerAddr
	minerInfo.GmtCreate = miner.GmtCreate
	minerInfo.GmtModified = miner.GmtModified
	minerInfo.SectorCount = miner.SectorCount
	minerInfo.SectorSize = miner.SectorSize
	minerInfo.ProvingSectorSize = miner.ProvingSectorSize.String()
	minerInfo.LockedFunds = miner.LockedFunds
	minerInfo.AccountBalance = miner.AccountBalance
	minerInfo.AvailableBalance = miner.AvailableBalance
	minerInfo.InitialPledge = miner.InitialPledge
	minerInfo.WalletAddr = miner.WalletAddr
	minerInfo.BlockCountPercent = miner.BlockCountPercent
	minerInfo.BlockCount = miner.BlockCount
	//utils.Log.Tracef("更新的内容有%+v", minerInfo)
	fs.toUpsertMiners[offset+1] = minerInfo
	fs.toUpdateMinerIndex++

	return nil
}

func (fs *Filscaner) modelsGetMinerstateAtTipset(address address.Address,
	tipsetHeight uint64) (*module.MinerStateAtTipset, error) {
	return module.FindMinerStateAtTipset(address, tipsetHeight)
}

func (fs *Filscaner) modelsSearchMiner(searchTxt string) ([]string, error) {
	ms, c := module.Connect(module.MinerCollection)
	defer ms.Close()

	miners := struct {
		Count  uint64
		Miners []string
	}{}

	qFind := []bson.M{
		{"$match": bson.M{"$or": []bson.M{{"peer_id": searchTxt}, bson.M{"miner_addr": searchTxt}}}},
		{"$sort": bson.M{"miner_time": -1}},
		{"$group": bson.M{"_id": "$miner_addr", "record": bson.M{"$first": "$$ROOT"}}},
		{"$group": bson.M{"_id": nil, "miners": bson.M{"$push": "$record.miner_addr"}, "count": bson.M{"$sum": 1}}},
		{"$project": bson.M{"_id": 0}}}
	utils.Log.Traceln(qFind)
	if err := c.Pipe(qFind).One(&miners); err != nil {
		return nil, err
	}

	return miners.Miners, nil

}

func ModelsMinerStateInTime(c *mgo.Collection, miners []string, at, start uint64) ([]*module.MinerStateAtTipset, error) {
	minerSize := len(miners)
	if minerSize == 0 {
		return nil, nil
	}

	if c == nil {
		var session *mgo.Session
		session, c = module.Connect(module.MinerCollection)
		defer session.Close()
	}

	qRes := struct {
		Miners []*module.MinerStateAtTipset `bson:"miners"`
	}{}

	qPipe := []bson.M{
		{"$match": bson.M{"mine_time": bson.M{"$gt": start, "$lte": at}, "miner_addr": bson.M{"$in": miners}}},
		{"$sort": bson.M{"mine_time": -1}},
		{"$group": bson.M{"_id": bson.M{"miner_addr": "$miner_addr"}, "miner": bson.M{"$first": "$$ROOT"}}},
		{"$group": bson.M{"_id": nil, "miners": bson.M{"$push": "$miner"}}},
	}
	//utils.Log.Traceln(qPipe)
	collation := &mgo.Collation{Locale: "zh", NumericOrdering: true}
	if err := c.Pipe(qPipe).Collation(collation).AllowDiskUse().One(&qRes); err != nil {
		return nil, err
	}

	return qRes.Miners, nil
}

func modelsMinerStateExistNewer(miner string, time int64) bool {
	utils.Log.Traceln("modelsMinerStateExistNewer")
	c, err := module.FindCount(module.MinerCollection,
		bson.M{"miner_addr": miner, "mine_time": bson.M{"$gt": time}}, nil)
	if err != nil {
		return false
	}
	return c > 0
}

// todo: use a loop to search miner instead of use '$match in ...', which may improve proformance
func (fs *Filscaner) modelsMinerPowerIncreaseInTime(miners []string, start, end uint64) (map[string]*MinerIncreasedPowerRecord, error) {
	if start >= end {
		return nil, errors.New("invalid parameters")
	}

	ms, c := module.Connect(module.MinerCollection)
	defer ms.Close()

	var match = bson.M{}
	minerSize := len(miners)
	if minerSize != 0 {
		if true {
			ors := make([]bson.M, minerSize)
			for index, m := range miners {
				ors[index] = bson.M{"miner_addr": m}
			}
			match = bson.M{"$or": ors}
		} else {
			match = bson.M{"miner_addr": bson.M{"$in": miners}}
		}
	}

	var mineTimeMatch bson.M
	if start > 0 {
		mineTimeMatch = bson.M{"$gte": start}
	}
	if end > 0 {
		if mineTimeMatch != nil {
			mineTimeMatch["$lt"] = end
		} else {
			mineTimeMatch = bson.M{"$lt": end}
		}
	}

	if mineTimeMatch != nil {
		match["mine_time"] = mineTimeMatch
	}

	qPipe := []bson.M{
		{"$match": match},
		{"$sort": bson.M{"mine_time": -1}},
		{"$addFields": bson.M{"fpower": bson.M{"$toDouble": "$power"}}},
		{"$group": bson.M{"_id": bson.M{"miner_addr": "$miner_addr"}, "fmin_power": bson.M{"$last": "$fpower"}, "fmax_power": bson.M{"$first": "$fpower"}, "record": bson.M{"$first": "$$ROOT"}}},
		{"$project": bson.M{"increased_power": bson.M{"$subtract": []string{"$fmax_power", "$fmin_power"}}, "record": "$record"}}}

	var qRes []*MinerIncreasedPowerRecord
	utils.Log.Traceln(qPipe)
	if err := c.Pipe(qPipe).Collation(fs.collation).AllowDiskUse().All(&qRes); err != nil {
		return nil, err
	}
	res := make(map[string]*MinerIncreasedPowerRecord)
	for _, r := range qRes {
		res[r.Record.MinerAddr] = r
	}
	return res, nil
}

func (fs *Filscaner) getTotalPowerAtTime(timestop uint64) (*module.MinerStateAtTipset, error) {
	ms, c := module.Connect(module.MinerCollection)
	defer ms.Close()
	match := bson.M{}
	if timestop > 0 {
		match["mine_time"] = bson.M{"$lte": timestop}
	}

	ops := []bson.M{
		{"$match": match},
		{"$sort": bson.M{"tipset_height": -1, "power": -1}},
		{"$group": bson.M{
			"_id":        bson.M{"mine_addr": "$miner_addr"},
			"record":     bson.M{"$first": "$$ROOT"},
			"totalpower": bson.M{"$max": "$totalpower"}}},
	}

	var res []MinerStateRecordInterface
	if err := c.Pipe(ops).All(&res); err != nil {
		return nil, err
	}

	if len(res) == 0 {
		return nil, nil
	}
	utils.Log.Traceln(ops)
	var records []module.MinerStateRecord
	if err := utils.UnmarshalJSON(res, &records); err != nil {
		return nil, err
	}
	return records[0].Record, nil
}

func modelsMinerTopPower(c *mgo.Collection, timeAt, offSet, limit int64) ([]*module.MinerStateAtTipset, uint64, error) {
	// if c == nil {
	// 	var session *mgo.Session
	// 	session, c = module.Connect(module.MinerCollection)
	// 	defer session.Close()
	// }

	qMatch := bson.M{}

	if timeAt != 0 {
		qMatch = bson.M{"mine_time": bson.M{"$lt": timeAt}}
	}

	qCount := []bson.M{
		{"$match": qMatch},
		{"$group": bson.M{"_id": bson.M{"miner_addr": "$miner_addr"}}},
		{"$group": bson.M{"_id": "", "count": bson.M{"$sum": 1}}},
	}
	//	utils.Log.Traceln(qCount)
	collation := &mgo.Collation{Locale: "zh", NumericOrdering: true}
	qCountRes := struct{ Count uint64 }{}
	if err := c.Pipe(qCount).Collation(collation).AllowDiskUse().One(&qCountRes); err != nil {
		return nil, 0, err
	}

	qResult := []bson.M{
		{"$match": qMatch},
		{"$sort": bson.M{"tipset_height": -1, "power": -1}},
		{"$group": bson.M{
			"_id":    bson.M{"miner_addr": "$miner_addr"},
			"record": bson.M{"$first": "$$ROOT"},
		}},
		{"$sort": bson.M{"record.power": -1}},
		{"$skip": offSet},
		{"$limit": limit},
	}

	var res []struct {
		Record *module.MinerStateAtTipset `json:"record" bson:"record"`
	}
	//	utils.Log.Traceln(qResult)
	// if err := c.Pipe(qResult).Collation(collation).AllowDiskUse().All(&res); err != nil {
	// 	return nil, 0, err
	// }
	_=qResult

	miners := make([]*module.MinerStateAtTipset, len(res))

	for index, miner := range res {
		miners[index] = miner.Record
	}

	return miners, qCountRes.Count, nil
}

func (fs *Filscaner) deleteMinerStateAt(tipsetHeight uint64) error {
	return module.Remove(
		module.MinerCollection,
		bson.M{"tipset_height": tipsetHeight})
}

func (fs *Filscaner) getMinerStateLte2(address address.Address, smollerthan uint64) ([]*module.MinerStateAtTipset, error) {
	var miner []*module.MinerStateAtTipset
	utils.Log.Traceln("getMinerStateLte2")
	err := module.FindSortLimit(module.MinerCollection, "-tipset_height",
		bson.M{
			"tipset_height": bson.M{"$lte": smollerthan},
			"miner_addr":    address.String(),
		},
		nil, &miner, 0, 2)
	return miner, err
}

// func (fs *Filscaner) toRespSlice(in []*module.MinerStateAtTipset) []*MinerState {
// 	var minerStates = make([]*MinerState, len(in))
// 	for index, miner := range in {
// 		minerStates[index] = miner.State()
// 	}
// 	return minerStates
// }

// func (fs *Filscaner) toRespMap(in map[string]*module.MinerStateAtTipset) map[string]*MinerState {
// 	var minerStates = make(map[string]*MinerState)

// 	for k, miner := range in {
// 		minerStates[k] = &MinerState{
// 			Address:      miner.MinerAddr,
// 			Power:        utils.XSizeString(miner.Power.Int),
// 			PowerPercent: utils.BigToPercent(miner.Power.Int, miner.TotalPower.Int),
// 			PeerId:       miner.PeerId,
// 		}
// 	}
// 	return minerStates
// }

func (fs *Filscaner) getMinerStateActivateAtTime(atTime uint64) ([]*module.MinerStateAtTipset, error) {
	ms, c := module.Connect(module.MinerCollection)
	defer ms.Close()

	beginTime := atTime - (60 * 60 * 24)
	time.Unix(int64(atTime), 0)

	ops := []bson.M{
		{"$match": bson.M{"mine_time": bson.M{"$gte": beginTime, "$lt": atTime}}},
		{"$sort": bson.M{"tipset_height": -1, "power": -1}},
		{"$group": bson.M{"_id": bson.M{"mine_addr": "$miner_addr"},
			"record": bson.M{"$first": "$$ROOT"},
		}},
	}
	utils.Log.Traceln(ops)
	var res []MinerStateRecordInterface
	if err := c.Pipe(ops).All(&res); err != nil {
		return nil, err
	}

	var records []module.MinerStateRecord
	if err := utils.UnmarshalJSON(res, &records); err != nil {
		return nil, err
	}

	var minerStates = make([]*module.MinerStateAtTipset, len(records))
	for index, record := range records {
		minerStates[index] = record.Record
	}

	return minerStates, nil
}

func (fs *Filscaner) modelActiveMinerCountAtTime(atTime, timeDiff uint64) (uint64, error) {
	ms, c := module.Connect(module.MinerCollection)
	defer ms.Close()

	var beginTime uint64
	if atTime > timeDiff {
		beginTime = atTime - timeDiff
	} else {
		beginTime = 0
	}

	ops := []bson.M{
		{"$match": bson.M{"mine_time": bson.M{"$gte": beginTime, "$lt": atTime}}},
		{"$sort": bson.M{"tipset_height": -1, "power": -1}},
		{"$group": bson.M{"_id": bson.M{"mine_addr": "$miner_addr"}}},
		{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}}}

	res := &struct {
		Count uint64
	}{}
	//	utils.Log.Traceln(ops)
	if err := c.Pipe(ops).Collation(fs.collation).AllowDiskUse().One(res); err != nil {
		return 0, err
	}

	return res.Count, nil
}

func (fs *Filscaner) modelsMinerPowerIncreaseTopN(start, end, offset, limit uint64) ([]*MinerIncreasedPowerRecord, uint64, error) {
	if start >= end {
		return nil, 0, errors.New("invalid parameters")
	}

	ms, c := module.Connect(module.MinerCollection)
	defer ms.Close()

	qCount := []bson.M{
		{"$match": bson.M{"mine_time": bson.M{"$gte": start, "$lt": end}}},
		{"$group": bson.M{"_id": bson.M{"miner_addr": "$miner_addr"}}},
		{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
	}

	qCountRes := struct {
		Count uint64
	}{}
	if err := c.Pipe(qCount).One(&qCountRes); err != nil {
		panic(err)
		return nil, 0, nil
	}

	qPipe := []bson.M{
		{"$match": bson.M{"mine_time": bson.M{"$gte": start, "$lt": end}}},
		{"$sort": bson.M{"mine_time": -1}},
		{"$addFields": bson.M{"fpower": bson.M{"$toDouble": "$power"}}},
		{"$group": bson.M{"_id": bson.M{"miner_addr": "$miner_addr"}, "fmin_power": bson.M{"$last": "$fpower"}, "fmax_power": bson.M{"$first": "$fpower"}, "record": bson.M{"$first": "$$ROOT"}, "old_power": bson.M{"$last": "$power"}}},
		{"$project": bson.M{"increased_power": bson.M{"$subtract": []string{"$fmax_power", "$fmin_power"}}, "record": "$record"}},
		{"$sort": bson.M{"increased_power": -1}},
		{"$skip": offset}, {"$limit": limit}}

	var qRes []*MinerIncreasedPowerRecord
	utils.Log.Traceln(qPipe)
	if err := c.Pipe(qPipe).Collation(fs.collation).AllowDiskUse().All(&qRes); err != nil {
		return nil, 0, err
	}

	return qRes, qCountRes.Count, nil
}

func (fs *Filscaner) modelsMinerBlockTopN(start, end, offset, limit uint64) ([]*MinedBlock, uint64, error) {
	if start >= end {
		return nil, 0, errors.New("invalid parameters")
	}

	ms, c := module.Connect(module.BlocksCollection)
	defer ms.Close()
	qMinerCount := []bson.M{
		{"$match": bson.M{"block_header.Timestamp": bson.M{"$gte": start, "$lt": end}}},
		{"$group": bson.M{"_id": bson.M{"miner": "$block_header.Miner"}}},
		{"$group": bson.M{"_id": nil, "miner_count": bson.M{"$sum": 1}}},
	}
	qCountRes := struct {
		MinerCount uint64 `bson:"miner_count" json:"miner_count"`
	}{}
	if err := c.Pipe(qMinerCount).One(&qCountRes); err != nil {
		panic(err)
		return nil, 0, nil
	}

	// TODO:这里可能需要把满足条件的区块数量计算出来再返回..
	qMiners := []bson.M{
		{"$match": bson.M{"block_header.Timestamp": bson.M{"$gte": start, "$lt": end}}},
		{"$group": bson.M{
			"_id":               bson.M{"miner": "$block_header.Miner"},
			"mined_block_count": bson.M{"$sum": 1},
			"miner":             bson.M{"$first": "$block_header.Miner"}}},
		{"$sort": bson.M{"mined_block_count": -1}},
		{"$skip": offset},
		{"$limit": limit}}

	var qMinerRes []*MinedBlock
	utils.Log.Traceln(qMiners)
	if err := c.Pipe(qMiners).All(&qMinerRes); err != nil {
		return nil, 0, nil
	}

	return qMinerRes, qCountRes.MinerCount, nil
}

func (fs *Filscaner) modelsGetTipsetAtTime(timeAt uint64, before bool) (uint64, error) {
	var res []*struct {
		Height uint64 `json:"height" bson:"height"`
	}

	cond := "$gte"
	s := "mine_time"

	if before {
		cond = "$lt"
		s = "-mine_time"
	}
	utils.Log.Traceln("tipset")
	err := module.FindSortLimit("tipset", s, bson.M{"mine_time": bson.M{cond: timeAt}},
		bson.M{"height": 1}, &res, 0, 1)
	if err != nil {
		return 0, err
	}

	if len(res) > 0 {
		return res[0].Height, nil
	}
	return 0, nil
}

func (fs *Filscaner) modelsBlockcountTimeRange(start, end uint64) (uint64, uint64, uint64, error) {
	ms, c := module.Connect(module.BlocksCollection)
	defer ms.Close()
	qPipe := []bson.M{
		{"$match": bson.M{"block_header.Timestamp": bson.M{"$gte": start, "$lt": end}}},
		{"$sort": bson.M{"block_header.Height": -1}},
		{"$group": bson.M{"_id": bson.M{"miner": "$block_header.Miner"},
			"mx_height": bson.M{"$first": "$block_header.Height"},
			"mi_height": bson.M{"$last": "$block_header.Height"}}},
		{"$group": bson.M{"_id": nil, "miner_count": bson.M{"$sum": 1},
			"min_height": bson.M{"$min": "$mi_height"},
			"max_height": bson.M{"$max": "$mx_height"}}}}

	res := &struct {
		MaxHeight  uint64 `json:"max_height" bson:"max_height"`
		MinHeight  uint64 `json:"min_height" bson:"min_height"`
		MinerCount uint64 `json:"miner_count" bson:"miner_count"`
	}{}
	//	utils.Log.Traceln(qPipe)
	err := c.Pipe(qPipe).One(res)
	if err != nil {
		return 0, 0, 0, err
	}

	return res.MinHeight, res.MaxHeight, res.MinerCount, nil
}

func (fs *Filscaner) modelsTotalBlockCount() (uint64, error) {
	utils.Log.Traceln("modelsTotalBlockCount")
	ms, c := module.Connect(module.BlocksCollection)
	defer ms.Close()

	total, err := c.Find(bson.M{}).Count()
	return uint64(total), err
}

func (fs *Filscaner) modelsBlockCountByMiner(miner string) (uint64, error) {
	utils.Log.Traceln("modelsBlockCountByMiner")
	ms, c := module.Connect(module.BlocksCollection)
	defer ms.Close()

	blockCount, err := c.Find(bson.M{"block_header.Miner": miner}).Count()
	if err != nil {
		return 0, err
	}
	return uint64(blockCount), nil
}

// 查找某段时间内的爆块总数, 和指定miner的爆块数量
func (fs *Filscaner) modelsBlockCountTimeRangeWithMiners(miners []string, start, end uint64) (map[string]uint64, uint64, error) {
	ms, c := module.Connect(module.BlocksCollection)
	defer ms.Close()

	totalCount, err := c.Find(bson.M{"block_header.Timestamp": bson.M{"$gt": start, "$lte": end}}).Count()
	if err != nil {
		return nil, 0, err
	}

	var q_pipe = []bson.M{
		{"$match": bson.M{"block_header.Miner": bson.M{"$in": miners}, "block_header.Timestamp": bson.M{"$gt": start, "$lte": end}}},
		{"$group": bson.M{"_id": bson.M{"miner": "$block_header.Miner"}, "block_count": bson.M{"$sum": 1}}},
		{"$project": bson.M{"_id": 0, "miner": "$_id.miner", "block_count": 1}}}
	var qRes []struct {
		Miner      string `bson:"miner"`
		BlockCount uint64 `bson:"block_count"`
	}
	utils.Log.Traceln(q_pipe)
	err = c.Pipe(q_pipe).All(&qRes)
	if err != nil {
		return nil, 0, err
	}

	res := make(map[string]uint64)
	for _, m := range qRes {
		res[m.Miner] = m.BlockCount
	}

	return res, uint64(totalCount), nil
}

type ModelsMinerList struct {
	TotalIncreasedPower  float64                  `bson:"total_increased_power"`
	TotalMinedBlockCount uint64                   `bson:"total_mined_block_count"`
	MinerCount           uint64                   `bson:"miner_count"`
	Miners               []*ModulesMinerListMiner `bson:"miners"`
	less                 func(i, j int) bool
}

type ModelsMinerReword struct {
	Miner      string  `bson:"miner" `
	Reword     float64 `bson:"reword"`
	BlockCount uint64  `bson:"block_count"`
}

func (ml *ModelsMinerList) GetMiners() []string {
	length := len(ml.Miners)
	if length == 0 {
		return nil
	}

	miners := make([]string, length)
	for index, m := range ml.Miners {
		miners[index] = m.MinerAddress
	}
	return miners
}

func (ml *ModelsMinerList) GetMinersMap() map[string]*ModulesMinerListMiner {
	if len(ml.Miners) == 0 {
		return nil
	}

	minerMap := make(map[string]*ModulesMinerListMiner)
	for _, m := range ml.Miners {
		minerMap[m.MinerAddress] = m
	}
	return minerMap
}

// func (ml *ModelsMinerList) APIRespData() *MinerListResp_Data {
// 	data := &MinerListResp_Data{}

// 	data.TotalIncreasedPower = strconv.FormatFloat(ml.TotalIncreasedPower, 'f', -1, 64)
// 	data.TotalIncreasedBlock = ml.TotalMinedBlockCount
// 	data.MienrCount = ml.MinerCount

// 	data.Miners = make([]*MinerInfo, len(ml.Miners))

// 	for index, m := range ml.Miners {
// 		var tag string
// 		res, err := module.GetTagByAddress(m.MinerAddress)
// 		if err != nil {

// 			//fmt.Println("get address name is err:", err)
// 			tag = "- -"
// 		} else {
// 			tag = res.Name
// 		}

// 		info := &MinerInfo{
// 			IncreasedPower:   strconv.FormatFloat(m.IncreasedPower, 'f', -1, 64),
// 			IncreasedBlock:   m.MinedBlockCount,
// 			Miner:            m.MinerAddress,
// 			PeerId:           m.PeerId,
// 			PowerPercent:     utils.FloatToPercent(m.IncreasedPower, ml.TotalIncreasedPower),
// 			BlockPercent:     utils.IntToPercent(m.MinedBlockCount, ml.TotalMinedBlockCount),
// 			MiningEfficiency: m.MiningEfficiency,
// 			StorageRate:      m.PowerRate,
// 			Tag:              tag,
// 		}
// 		// StorageRate:      m.PowerRate}
// 		data.Miners[index] = info
// 	}
// 	return data
// }

// returns no-setted miner addresses
func (ml *ModelsMinerList) SetPowerValues(ml_src *ModelsMinerList) {
	miner_map := ml_src.GetMinersMap()
	if miner_map == nil {
		return
	}

	ml.TotalIncreasedPower = ml_src.TotalIncreasedPower

	for _, miner := range ml.Miners {
		src_miner, isok := miner_map[miner.MinerAddress]
		if !isok || src_miner == nil {
			miner.PowerRate = "0.00"
			continue
		}
		miner.IncreasedPower = src_miner.IncreasedPower
		miner.PeerId = src_miner.PeerId
		miner.WalletAddress = src_miner.WalletAddress
		miner.PowerRate = src_miner.PowerRate
		// miner.MiningEfficiency = src_miner.MiningEfficiency
	}
}

func (ml *ModelsMinerList) SetBlockValues(ml_src *ModelsMinerList) {
	miner_map := ml_src.GetMinersMap()
	if miner_map == nil {
		return
	}
	ml.TotalMinedBlockCount = ml_src.TotalMinedBlockCount
	for _, miner := range ml.Miners {
		src_miner, isok := miner_map[miner.MinerAddress]
		if !isok || src_miner == nil {
			continue
		}
		miner.MinedBlockCount = src_miner.MinedBlockCount
	}
}

func (ml *ModelsMinerList) SortBYMiningEfficency(sort_type int) {
	ml.less = ml.less_mining_efficency
	if sort_type < 0 {
		sort.Sort(sort.Reverse(ml))
	} else {
		sort.Sort(ml)
	}
}

func (ml *ModelsMinerList) Len() int {
	return len(ml.Miners)
}

func (ml *ModelsMinerList) less_mining_efficency(i, j int) bool {
	if ml.Miners[i].MiningEfficiency == "+Inf" {
		return true
	}
	if ml.Miners[j].MiningEfficiency == "+Inf" {
		return false
	}

	fi := utils.StringToFloat(ml.Miners[i].MiningEfficiency)
	fj := utils.StringToFloat(ml.Miners[j].MiningEfficiency)
	return fi < fj
}

func (ml *ModelsMinerList) Less(i, j int) bool {
	return ml.less(i, j)
}

func (ml *ModelsMinerList) Swap(i, j int) {
	ml.Miners[i], ml.Miners[j] = ml.Miners[j], ml.Miners[i]
}

func (ml *ModelsMinerList) SetBlockEfficiency(ml_src *ModelsMinerList) {
	miner_map := ml_src.GetMinersMap()
	unit_gb := float64(1 << 30)
	if miner_map == nil {
		return
	}

	for _, miner := range ml.Miners {
		src_miner, isok := miner_map[miner.MinerAddress]
		if !isok || src_miner == nil {
			continue
		}
		miner.PeerId = src_miner.PeerId
		miner.MiningEfficiency = fmt.Sprintf("%.4f", float64(miner.MinedBlockCount)*unit_gb/src_miner.IncreasedPower)
	}
}

type ModulesMinerListMiner struct {
	IncreasedPower   float64 `bson:"increased_power" json:"increased_power"`
	MinedBlockCount  uint64  `bson:"mined_block_count" json:"mined_block_count"`
	MinerAddress     string  `bson:"miner_addr" json:"miner_addr"`
	WalletAddress    string  `bson:"wallet_addr" json:"wallet_addr"`
	PowerRate        string  `bson:"power_rate"`
	MiningEfficiency string  `bson:"mining_efficiency"`
	PeerId           string  `bson:"peer_id" json:"peer_id"`
}

func (fs *Filscaner) modelsMinerListSortPower(miners []string, start, end, offset, limit uint64, sortField string, sort int) (*ModelsMinerList, error) {
	ms, c := module.Connect(module.MinerCollection)
	defer ms.Close()

	var powerSortFieldMap = map[string]string{
		"power":      "miner.increased_power",
		"power_rate": "miner.power_rate"}

	sortField, useSort := powerSortFieldMap[sortField]
	qPipe := []bson.M{
		{"$match": bson.M{"mine_time": bson.M{"$gte": start, "$lt": end}}},
		{"$sort": bson.M{"mine_time": -1}},
		{"$addFields": bson.M{"fpower": bson.M{"$toDouble": "$power"}}},
		{"$group": bson.M{"_id": bson.M{"miner": "$miner_addr"}, "fmin_power": bson.M{"$last": "$fpower"}, "fmax_power": bson.M{"$first": "$fpower"}, "miner": bson.M{"$first": "$$ROOT"}}},
		{"$set": bson.M{"miner.increased_power": bson.M{"$subtract": []string{"$fmax_power", "$fmin_power"}}}},
		{"$set": bson.M{"miner.power_rate": bson.M{"$divide": []interface{}{"$miner.increased_power", (float64(end-start) / 3600 * 1024 * 1024 * 1024)}}}}}

	if useSort && sortField != "" {
		qPipe = append(qPipe, bson.M{"$sort": bson.M{sortField: sort}})
	}

	qPipe = append(qPipe, bson.M{"$set": bson.M{"miner.power_rate": bson.M{"$toString": "$miner.power_rate"}}})
	qPipe = append(qPipe, bson.M{"$group": bson.M{"_id": nil, "miners": bson.M{"$push": "$miner"}, "total_increased_power": bson.M{"$sum": "$miner.increased_power"}, "miner_count": bson.M{"$sum": 1}}})

	if len(miners) != 0 {
		qPipe = append(qPipe, bson.M{"$set": bson.M{"miner_filters": miners}})
		qPipe = append(qPipe, bson.M{"$project": bson.M{"miner_count": 1, "total_increased_power": 1,
			"miners": bson.M{"$filter": bson.M{"input": "$miners", "as": "miners", "cond": bson.M{"$in": []string{"$$miners.miner_addr", "$miner_filters"}}}}}})
	}

	qPipe = append(qPipe, bson.M{"$project": bson.M{"miner_count": 1, "total_increased_power": 1, "miners": bson.M{"$slice": []interface{}{"$miners", offset, limit}}}})
	qRes := &ModelsMinerList{}
	//utils.Log.Traceln(qPipe)
	if err := c.Pipe(qPipe).Collation(fs.collation).AllowDiskUse().One(&qRes); err != nil {
		return nil, err
	}

	return qRes, nil
}

func (fs *Filscaner) modelsMinerListSortBlock(miners []string, start, end, offset, limit uint64, sortField string, sort int) (*ModelsMinerList, error) {
	ms, c := module.Connect(module.BlocksCollection)
	defer ms.Close()

	var blockSortFieldMap = map[string]string{
		"block":             "miner.mined_block_count",
		"mining_efficiency": "miner.mining_efficiency"}

	sortField, useSort := blockSortFieldMap[sortField]
	qPipe := []bson.M{
		{"$match": bson.M{"block_header.Timestamp": bson.M{"$gte": start, "$lt": end}}},
		{"$group": bson.M{"_id": bson.M{"miner": "$block_header.Miner"}, "block_count": bson.M{"$sum": 1}, "miner": bson.M{"$first": "$block_header.Miner"}}},
		{"$set": bson.M{"miner.mined_block_count": "$block_count", "miner.miner_addr": "$miner"}},
	}
	if useSort {
		qPipe = append(qPipe, bson.M{"$sort": bson.M{sortField: sort}})
	}

	qPipe = append(qPipe, bson.M{"$set": bson.M{"miner.mining_efficiency": bson.M{"$toString": "$miner.mining_efficiency"}}})
	qPipe = append(qPipe, bson.M{"$group": bson.M{"_id": nil, "miner_count": bson.M{"$sum": 1}, "total_mined_block_count": bson.M{"$sum": "$miner.mined_block_count"}, "miners": bson.M{"$push": "$miner"}}})

	if len(miners) != 0 {
		qPipe = append(qPipe, bson.M{"$set": bson.M{"miner_filters": miners}})
		qPipe = append(qPipe,
			bson.M{"$project": bson.M{"_id": 0, "miner_count": 1, "total_mined_block_count": 1,
				"miners": bson.M{"$filter": bson.M{"input": "$miners", "as": "ms", "cond": bson.M{"$in": []string{"$$ms.miner_addr", "$miner_filters"}}}}}})
	}
	qPipe = append(qPipe, bson.M{"$project": bson.M{"miner_count": 1, "total_mined_block_count": 1, "miners": bson.M{"$slice": []interface{}{"$miners", offset, limit}}}})

	qRes := &ModelsMinerList{}
	//utils.Log.Traceln(qPipe)
	if err := c.Pipe(qPipe).Collation(fs.collation).AllowDiskUse().One(&qRes); err != nil {
		return nil, err
	}
	return qRes, nil
}

func modelsBlockReleasedRewardsAtHeight(height uint64) (*ModelsBlockReward, error) {
	ms, c := module.Connect(module.BlockRewardCollection)
	defer ms.Close()

	blockReward := &ModelsBlockReward{}

	err := c.Find(bson.M{"height": bson.M{"$lte": height}}).Sort("-height").Limit(1).One(blockReward)
	if err != nil {
		return nil, err
	}
	return blockReward, nil
}

func modelsBulkUpsertBlockReward(brs []*ModelsBlockReward, size int) error {
	upsertPairs := make([]interface{}, size*2)

	for i := 0; i < size; i++ {
		br := brs[i]
		upsertPairs[i*2] = bson.M{"height": br.Height}
		upsertPairs[i*2+1] = br
	}
	//	utils.Log.Traceln(upsertPairs)
	_, err := module.BulkUpsert(nil, module.BlockRewardCollection, upsertPairs)
	return err
}

func modelsBlockRewardHead() (*ModelsBlockReward, error) {
	utils.Log.Traceln("modelsBlockRewardHead")
	ms, c := module.Connect(module.BlockRewardCollection)
	defer ms.Close()

	blockReward := &ModelsBlockReward{}
	err := c.Find(nil).Sort("-height").One(&blockReward)
	if err != nil {
		if err == mgo.ErrNotFound {
			blockReward.ReleasedRewards = &module.BsonBigint{Int: big.NewInt(0)}
			return blockReward, nil
		}
		return nil, err
	}
	return blockReward, nil
}

func (fs *Filscaner) minerBlockReward(addr string) (string, error) {
	ms, c := module.Connect(module.BlocksCollection)
	defer ms.Close()
	qPipe := []bson.M{
		{"$match": bson.M{"block_header.Miner": addr}},
		{"$addFields": bson.M{"doubleReword": bson.M{"$toDouble": "$block_reword"}}},
		{"$addFields": bson.M{"WinCount": "$block_header.ElectionProof.WinCount"}},
		{"$group": bson.M{"_id": bson.M{"miner": "$block_header.Miner"}, "block_count": bson.M{"$sum": 1}, "miner": bson.M{"$first": "$block_header.Miner"}}},
		//{"$set": bson.M{"reword": bson.M{"$multiply": []string{"$WinCount", "$doubleReword"}}}},
		{"$set": bson.M{"block_count": "$block_count", "miner": "$miner"}},
		{"$project": bson.M{"reword": 1, "block_count": 1, "miner": 1}},
	}
	qRes := &ModelsMinerReword{}
	utils.Log.Traceln(qPipe)
	if err := c.Pipe(qPipe).Collation(fs.collation).AllowDiskUse().One(&qRes); err != nil {
		return "", err
	}
	return "", nil
}
