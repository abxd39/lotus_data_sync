package syncer

import (

	"lotus_data_sync/module"
	"lotus_data_sync/utils"
	"github.com/globalsign/mgo"

	"os"
	"runtime"
	"sync"
	"time"
)

const MaxMinerCacheCount = int64(6)

func (fs *Filscaner) initMinersCaches() error {
	beginTime := time.Now()
	defer func() {
		second := uint64(time.Since(beginTime).Seconds())
		utils.Log.Tracef("››››››››››››››››init_miner_states finished, time_used=%d(s)!!!››››››››››››››››\n", second)

	}()
	fs.minerStateChan = make(chan *module.MinerStateAtTipset, 256)
	session, db := module.Copy()
	defer session.Close()
	c := db.C(module.MinerCollection)
	var (
		timeNow      = time.Now().Unix()
		maxCacheSize = int64(24)
		timeDiff     = int64(3600)
		timeStart    = timeNow - (timeNow % timeDiff) - (timeDiff * (maxCacheSize - 2))
	)
	utils.Log.Tracef("››››››››››››››››init_miner_states finished %d", 1)
	miners, _, err := modelsMinerTopPower(c, timeNow, 0, MaxMinerCacheCount)
	if err != nil && err != mgo.ErrNotFound {
		utils.Log.Errorf("error, load_top_power_miner failed, message:%s\n", err.Error())
		return err
	}
	_=miners
	//fs.minerCache24h = (&fsMinerCache{}).init(timeDiff, timeStart, MaxMinerCacheCount, maxCacheSize)

	if false { // this is a testing for time_to_index
		timeNow = 23
		maxCacheSize = int64(3)
		timeDiff = int64(5)
		timeStart = timeNow - (timeNow % timeDiff) - (timeDiff * (maxCacheSize - 2))

		cache := (&fsMinerCache{}).init(timeDiff, timeStart, MaxMinerCacheCount, maxCacheSize)
		for i := int64(14); i < 27; i++ {
			index, ofst := cache.timeToIndex(i)
			var wantIndex = int64(0)
			var wantOffset = false
			if i <= 15 {
				wantIndex = 2
				wantOffset = false
			} else if i > 15 && i <= 20 {
				wantIndex = 1
				wantOffset = false
			} else if i > 20 && i <= 25 {
				wantIndex = 0
				wantOffset = false
			} else if i > 25 {
				wantIndex = 0
				wantOffset = true
			}

			result := "test ‹success›"
			if wantIndex != index || wantOffset != ofst {
				result = "test ‹failed›"
			}
			utils.Log.Infof("test result:%s, time_now=%d, index=%d,is_ofst=%t, want_index=%d, want_ofsetd=%t\n",
				result, i, index, ofst, wantIndex, wantOffset)
		}
	}
	utils.Log.Tracef("››››››››››››››››init_miner_states finished %d", 2)
	// if err := fs.minerCache24h.modelsSetIndexAndLoadHistroy(c, miners, 0, true); err != nil {
	// 	utils.Log.Errorf("error, miner_cache.load_history failed, message:%s\n", err.Error())
	// 	return err
	// }

	timeDiff = 86400
	maxCacheSize = 30
	timeStart = timeNow - (timeNow % timeDiff) - (timeDiff * (maxCacheSize - 1))
	utils.Log.Tracef("››››››››››››››››init_miner_states finished %d", 3)
	// fs.minerCache1day = (&fsMinerCache{}).init(timeDiff, timeStart, MaxMinerCacheCount, maxCacheSize)
	// if err := fs.minerCache1day.modelsSetIndexAndLoadHistroy(c, miners, 0, true); err != nil {
	// 	utils.Log.Errorf("error, miner_cache.load_history failed, message:%s\n", err.Error())
	// 	return err
	// }

	return nil
}

type fsMinerCache struct {
	miners            map[string][]*module.MinerStateAtTipset
	maxCachedSize     int64
	maxMinerCount     int64
	recentRefreshTime int64
	min               string
	max               string
	mutex             sync.Mutex
	timeDuration      int64
	startTime         int64
}

func (this *fsMinerCache) init(timeDuration, startTime, minerCount, cacheSize int64) *fsMinerCache {
	this.miners = make(map[string][]*module.MinerStateAtTipset)
	this.timeDuration = timeDuration
	this.startTime = startTime
	this.maxMinerCount = minerCount
	this.maxCachedSize = cacheSize

	if false {
		index, offset := this.timeToIndex(time.Now().Unix())
		utils.Log.Traceln(index, offset)
		index, offset = this.timeToIndex(this.startTime - 1)
		utils.Log.Traceln(index, offset)
	}
	return this
}

// 返回值:bool, 是否所有数据为nil, 都为伪造出来的
// func (this *fsMinerCache) index(index int) ([]*fspt.MinerState, bool) {
// 	if index < 0 || int64(index) >= this.maxCachedSize {
// 		return nil, false
// 	}
// 	this.lock()
// 	defer this.unlock()
// 	size := len(this.miners)
// 	stats := make([]*fspt.MinerState, size+1)
// 	var (
// 		maxTotal = big.NewInt(0)
// 		other    = big.NewInt(0)
// 	)
// 	i := 0
// 	var minerState *fspt.MinerState
// 	var allIsNil = true
// 	for _, v := range this.miners {
// 		state := v[index]
// 		if state == nil {
// 			minerState = &fspt.MinerState{
// 				Address:      v[0].MinerAddr,
// 				Power:        "0",
// 				PowerPercent: "0.00%"}
// 		} else {
// 			allIsNil = false
// 			minerState = state.State()
// 			other.Add(other, state.Power.Int)
// 			if maxTotal.Cmp(state.TotalPower.Int) < 0 {
// 				maxTotal.Set(state.TotalPower.Int)
// 			}
// 		}
// 		stats[i] = minerState
// 		i++
// 	}
// 	other.Sub(maxTotal, other)
// 	powerStr := "0"
// 	powerPercentStr := "0.00%"
// 	if other.Cmp(big.NewInt(0)) > 0 {
// 		powerStr = utils.XSizeString(other)
// 		powerPercentStr = utils.BigToPercent(other, maxTotal)
// 	}
// 	stats[i] = &fspt.MinerState{
// 		Address:      "other",
// 		Power:        powerStr,
// 		PowerPercent: powerPercentStr,
// 	}
// 	return stats, allIsNil
// }

func (this *fsMinerCache) display(address string) {
	if arr, exist := this.miners[address]; exist {
		for index, miner := range arr {
			if miner == nil {
				continue
			}
			utils.Log.Infof("index:%d, power:%s\n", index, utils.ToXSize(miner.Power.Int, utils.TB))
		}
	}
}

func (this *fsMinerCache) modelsSetIndexAndLoadHistroy(c *mgo.Collection, miners []*module.MinerStateAtTipset, index int64, lock bool) error {
	startTimeModelsSetIndexLoadHistory := time.Now()
	defer func() {
		utils.Log.Infof(" models_set_index_and_load_histroy, used_time = %d(s)",
			int(time.Since(startTimeModelsSetIndexLoadHistory).Seconds()))
	}()

	if len(miners) == 0 {
		return nil
	}

	if lock {
		this.lock()
		defer this.unlock()
	}
	this.setMinersAtIndex(miners, index)
	slcMiners := utils.SlcObjToSlc(miners, "MinerAddr").([]string)
	timeAt := this.startTime
	var start uint64
	if c == nil {
		session, db := module.Copy()
		c = db.C(module.MinerCollection)
		defer session.Close()
	}
	for idex := this.maxCachedSize - 1; idex > index; idex-- {
		if idex == this.maxCachedSize-1 {
			start = 0
		} else {
			start = uint64(timeAt - this.timeDuration)
		}
		minerAtTipsets, err := ModelsMinerStateInTime(c, slcMiners, uint64(timeAt), start)
		if err != nil {
			if err == mgo.ErrNotFound {
				timeAt += this.timeDuration
				continue
			}
			return err
		}
		this.setMinersAtIndex(minerAtTipsets, idex)
		timeAt += this.timeDuration
	}
	return nil
}

func (this *fsMinerCache) lock() {
	this.mutex.Lock()
}

func (this *fsMinerCache) unlock() {
	this.mutex.Unlock()
}

func (this *fsMinerCache) nextRefreshTime() int64 {
	return this.startTime + this.maxCachedSize*this.timeDuration
}

func (this *fsMinerCache) timeToIndex(time int64) (int64, bool) {
	offset := false
	if time <= this.startTime {
		return this.maxCachedSize - 1, false
	} else if time > (this.startTime + ((this.maxCachedSize - 1) * this.timeDuration)) {
		return 0, true
	}
	diff := time - this.startTime
	diff += this.timeDuration - 1
	index := diff / this.timeDuration
	index = this.maxCachedSize - 1 - index

	return index, offset
}

func (this *fsMinerCache) doOffset() {
	for _, minerStates := range this.miners {
		for index := this.maxCachedSize - 1; index > 0; index-- {
			minerStates[index] = minerStates[index-1]
		}
	}
	this.startTime += this.timeDuration
}

func (this *fsMinerCache) update(in *module.MinerStateAtTipset) error {
	defer func() {
		if err := recover(); err != nil {
			utils.Log.Errorf("%s\n", err)
			buf := make([]byte, 1<<16)
			runtime.Stack(buf, true)
			utils.Log.Errorln("buf", string(buf))
			os.Exit(1)
		}

	}()
	this.lock()
	defer this.unlock()
	// todo : checkout why in day duration, first update, ofseted is 'true'
	index, ofsted := this.timeToIndex(int64(in.MineTime))
	if ofsted && index == 0 {
		this.doOffset()
	}
	if miners, exist := this.miners[in.MinerAddr]; exist { // 如果已经存在
		if true {
			this.setMinersAtIndex([]*module.MinerStateAtTipset{in}, index)
		} else {
			if miners[index] == nil || miners[index].MineTime < in.MineTime {
				miners[index] = in
				for i := index - 1; i > 0; i-- {
					if miners[i] == nil {
						miners[i] = miners[i+1]
					}
				}
			}
		}
	} else {
		if int64(len(this.miners)) < this.maxMinerCount {
			if err := this.modelsSetIndexAndLoadHistroy(nil,
				[]*module.MinerStateAtTipset{in}, index, false); err != nil {
				return err
			}
		} else { // 检查最低算力是否小于in_miner的算力
			minPowerMiner := this.miners[this.min][0]
			if minPowerMiner.Power.Cmp(in.Power.Int) < 0 {
				// check to_insert_miner, if exist a newer state in database, do nothing
				if index != 0 && modelsMinerStateExistNewer(in.MinerAddr, int64(in.MineTime)) {
					return nil
				}
				if err := this.modelsSetIndexAndLoadHistroy(nil,
					[]*module.MinerStateAtTipset{in}, index, false); err != nil {
					return err
				}
			} // else // nothing is needed to do
		}
	}
	return nil
}

func (this *fsMinerCache) refreshMinMax(lock bool) {
	if lock {
		this.lock()
		defer this.unlock()
	}

	var min, max *module.MinerStateAtTipset
	for _, v := range this.miners {
		if min == nil || v[0].Power.Cmp(min.Power.Int) < 0 {
			min = v[0]
		}
		if max == nil || v[0].Power.Cmp(max.Power.Int) > 0 {
			max = v[0]
		}
	}
	if min != nil {
		this.min = min.MinerAddr
	}
	if max != nil {
		this.max = max.MinerAddr
	}
}

func (this *fsMinerCache) setMinersAtIndex(miners []*module.MinerStateAtTipset, index int64) {
	inMiners := utils.SlcToMap(miners, "MinerAddr", true).(map[string]*module.MinerStateAtTipset)
	var arr []*module.MinerStateAtTipset
	for ink, inv := range inMiners {
		var exist = false
		if arr, exist = this.miners[ink]; exist {
			if arr[index] == nil || arr[index].MineTime < inv.MineTime {
				arr[index] = inv
			}
			for i := index - 1; i >= 0; i-- {
				if arr[i] != nil && arr[i].MineTime >= inv.MineTime {
					break
				}
				arr[i] = inv
			}
		} else { // not exist
			if len(this.miners) < int(this.maxMinerCount) {
				arr = make([]*module.MinerStateAtTipset, this.maxCachedSize)
				arr[index] = inv
				this.miners[ink] = arr
				for i := index - 1; i >= 0; i-- {
					arr[i] = arr[i+1]
				}
			} else {
				arr = this.miners[this.min]
				// check and replace min power miner state
				if arr[0].Power.Cmp(inv.Power.Int) < 0 {
					arr[index] = inv
					for i := index - 1; i >= 0; i-- {
						arr[i] = arr[i+1]
					}
					for i := index + 1; i < this.maxCachedSize; i++ {
						arr[i] = nil
					}
					this.miners[inv.MinerAddr] = arr
					delete(this.miners, this.min)
				}
			}
			this.refreshMinMax(false)
		}
	}

	if false {
		if index == this.maxCachedSize-1 {
			return
		}
		for k, v := range this.miners {
			if _, exist := inMiners[k]; exist {
				continue
			}
			if v[index+1] != nil && (v[index] == nil || v[index].MineTime < v[index].MineTime) {
				v[index] = v[index+1]
			}
		}
	}
}
