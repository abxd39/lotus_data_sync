package syncer

// import (
// 	"context"
// 	"lotus_data_sync/module"
// 	"lotus_data_sync/utils"
// 	"fmt"
// 	"github.com/filecoin-project/go-address"
// 	"github.com/filecoin-project/go-state-types/abi"
// 	big2 "github.com/filecoin-project/go-state-types/big"
// 	"github.com/filecoin-project/lotus/build"
// 	"github.com/filecoin-project/lotus/chain/types"
// 	"github.com/ipfs-force-community/common"
// 	"math/big"
// 	"time"
// )

// var respSuccess = &common.Result{Code: 3, Msg: "success"}
// var respSearchError = &common.Result{Code: 5, Msg: "search failed"}
// var respInvalidParama = &common.Result{Code: 5, Msg: "invalid param"}
// var respLotusApiError = &common.Result{Code: 5, Msg: "lotus api failed"}

// func (fs *Filscaner) errorResp(err error) *common.Result {
// 	return common.NewResult(3, err.Error())
// }

// func (fs *Filscaner) FilNetworkBlockReward(ctx context.Context, req *FutureBlockRewardReq) (*FutureBlockRewardResp, error) {
// 	resp := &FutureBlockRewardResp{}

// 	timediff := req.TimeDiff
// 	repeate := req.Repeate
// 	time_now := uint64(time.Now().Unix())
// 	rewards, _, err := fs.futureBlockRewards(timediff, repeate)
// 	if err != nil {
// 		resp.Res = respSearchError
// 		return nil, err
// 	}

// 	resp.Data = make([]*FutureBlockRewardResp_Data, req.Repeate)

// 	for index, v := range rewards {
// 		resp.Data[index] = &FutureBlockRewardResp_Data{
// 			Time:         time_now,
// 			BlockRewards: utils.ToFilStr(v)}
// 		time_now += timediff
// 	}

// 	resp.Res = respSuccess
// 	return resp, nil
// }

// var TotalRewards = types.FromFil(build.FilAllocStorageMining).Int

// func (fs *Filscaner) FilOutStanding(ctx context.Context, req *FilOutstandReq) (*FiloutstandResp, error) {
// 	start := req.TimeAt
// 	diff := req.TimeDiff
// 	repeat := req.Repeate

// 	timeNow := uint64(time.Now().Unix())
// 	if start == 0 {
// 		start = timeNow
// 	}

// 	start = start - (diff * repeat)

// 	resp := &FiloutstandResp{}

// 	var data []*FiloutstandResp_Data

// 	setWithLastData := func(data []*FiloutstandResp_Data, iii *FiloutstandResp_Data) []*FiloutstandResp_Data {
// 		length := len(data)
// 		if length == 0 {
// 			zero_fil := utils.ToFilStr(big.NewInt(0))
// 			iii.Floating = zero_fil
// 			iii.PlegeCollateral = zero_fil
// 			iii.PlegeCollateral = zero_fil
// 		} else {
// 			iii = data[length-1]
// 		}
// 		return append(data, iii)
// 	}

// 	for i := uint64(0); i < repeat; i++ {
// 		if start < fs.chainGenesisTime {
// 			continue
// 		}
// 		if start > timeNow {
// 			break
// 		}

// 		filoutrespData := &FiloutstandResp_Data{TimeStart: start, TimeEnd: start + diff}
// 		_, maxHeight, _, err := fs.modelsBlockcountTimeRange(start, start+diff)
// 		start += diff

// 		if err != nil {
// 			utils.Log.Errorln(err)
// 			continue
// 		}

// 		maxReleasedReward := fs.releasedRewardAtHeight(maxHeight)

// 		filoutrespData.Floating = utils.ToFilStr(maxReleasedReward)

// 		tipset, err := fs.api.ChainGetTipSetByHeight(fs.ctx, abi.ChainEpoch(maxHeight), types.EmptyTSK)
// 		if err != nil {
// 			utils.Log.Errorf("chain_get_tipset_by_height(%d) failed,message;%s", tipset.Height(), err.Error())
// 			continue
// 		}
// 		// WEN
// 		// pleged, err := fs.api.StatePledgeCollateral(ctx, tipset.Key())  V0.4.2 return big.Zero()
// 		tipset = tipset
// 		pleged := big2.Zero()
// 		if err != nil {
// 			setWithLastData(data, filoutrespData)
// 			utils.Log.Errorf("StatePledgeCollateral failed,message;%s", err.Error())
// 			return nil, err
// 		}

// 		filoutrespData.PlegeCollateral = utils.ToFilStr(pleged.Int)
// 		filoutrespData.Outstanding = fmt.Sprintf("%.4f", utils.ToFil(maxReleasedReward)+utils.ToFil(pleged.Int))
// 		data = append(data, filoutrespData)
// 	}

// 	resp.Data = data
// 	resp.Res = respSuccess
// 	return resp, nil
// }

// // 计算历史时间周期内的区块奖励
// func (fs *Filscaner) CumulativeBlockRewardsOverTime(ctx context.Context, req *CBROReq) (*CBROResp, error) {
// 	start := req.TimeStart
// 	diff := req.TimeDiff
// 	repeate := req.Repeate

// 	if start < fs.chainGenesisTime {
// 		start = fs.chainGenesisTime
// 	}
// 	resp := &CBROResp{}
// 	// 这个数据也是大致计算的,并不完全准确, 完全准确的数据应该是:
// 	// current_reward_remain - vm.blockreward(rewards_remain * blocks_count_in_tipset)
// 	// vm.MiningReward()
// 	time_now := uint64(time.Now().Unix())
// 	// TODO:需要检查时间合法性!!!
// 	// rewards := make([]*big.Int, repeate)
// 	var data []*CBROResp_Data
// 	var maxReleased *big.Int
// 	offset := 0
// 	for i := uint64(0); i < repeate; i++ {
// 		s := start
// 		e := start + diff
// 		start += diff
// 		if start > time_now {
// 			break
// 		}
// 		cbrrespData := &CBROResp_Data{
// 			TimeStart: start,
// 			TimeEnd:   start + diff}
// 		// 从数据库读取时间周期内的块高变化
// 		_, maxHeight, minerCount, err := fs.modelsBlockcountTimeRange(s, e)
// 		if err != nil || maxHeight == 0 {
// 			if offset > 0 {
// 				cbrrespData.BlocksReward = data[offset-1].BlocksReward
// 			} else {
// 				continue
// 			}
// 		} else {
// 			maxReleased = fs.releasedRewardAtHeight(maxHeight)
// 			cbrrespData.BlocksReward = utils.ToFilStr(maxReleased)
// 			cbrrespData.MinerCount = minerCount
// 		}

// 		data = append(data, cbrrespData)
// 		offset++
// 	}
// 	resp.Data = data
// 	resp.Res = respSuccess
// 	return resp, nil
// }
// func (fs *Filscaner) MinerRewards(ctx context.Context, req *MinerRewardsReq) (*MinerRewardsResp, error) {
// 	resp := &MinerRewardsResp{}

// 	var start, count uint64
// 	var isHeight bool
// 	if req.HeightCount != 0 {
// 		isHeight = true
// 		start = req.HeightStart
// 		count = req.HeightCount
// 	} else {
// 		isHeight = false
// 		start = req.TimeStart
// 		count = req.TimeDiff
// 	}
// 	if count == 0 {
// 		resp.Res = respInvalidParama
// 		return resp, nil
// 	}
// 	// convert t3 address to t0 address
// 	var miners = req.Miners
// 	var workerMap map[string]string // t0 -> t3
// 	if len(miners) == 0 && len(req.Workers) != 0 {
// 		var err error
// 		if workerMap, err = module.GetMinersByT3(req.Workers); err != nil {
// 			resp.Res = respSearchError
// 			return resp, nil
// 		} else {
// 			miners = make([]string, len(workerMap))
// 			index := 0
// 			for t0, _ := range workerMap {
// 				miners[index] = t0
// 			}
// 		}
// 	}

// 	minerRewardsMap, err := module.MinerRewardInTimeRange(start, count, miners, isHeight)
// 	if err != nil {
// 		resp.Res = respSearchError
// 		return resp, nil
// 	}
// 	// resp.Data = &MinerRewardsResp_Data { }
// 	minersRewards := make(map[string]*MinerRewards)
// 	for addr, re := range minerRewardsMap {
// 		mrds, exist := minersRewards[addr]
// 		if mrds == nil || !exist {
// 			mrds = &MinerRewards{
// 				Miner: addr, TotalRewards: 0}
// 			minersRewards[addr] = mrds
// 		}
// 		if workerMap != nil {
// 			if worker, exist := workerMap[addr]; worker != "" && exist {
// 				mrds.Woker = worker
// 			}
// 		}
// 		for _, xxx := range re.BlockRewards {
// 			rewardFil := float32(xxx.RewardFil())
// 			mrds.Items = append(mrds.Items, &MinerRewards_Item{
// 				Rewards: rewardFil,
// 				Height:  xxx.Height})
// 			mrds.TotalRewards = float32(utils.TruncateNative(float64(mrds.TotalRewards+rewardFil), utils.PrecisionDefault))
// 		}
// 	}
// 	resp.Res = respSuccess
// 	if len(minersRewards) != 0 {
// 		resp.Data = &MinerRewardsResp_Data{
// 			MinerRewards: minersRewards,
// 		}
// 	}
// 	return resp, nil
// }
// func (fs *Filscaner) BalanceIncreased(ctx context.Context, req *BalanceIncreaseReq) (*BalanceIncreaseResp, error) {
// 	resp := &BalanceIncreaseResp{}
// 	timeStart := req.TimeStart
// 	timeEnd := req.TimeEnd
// 	miner, err := address.NewFromString(req.Address)
// 	if err != nil {
// 		resp.Res = &common.Result{Code: 3, Msg: "invalid address"}
// 		return resp, nil
// 	}
// 	module.GetTipsetByTime(int64(timeStart))
// 	heightStart, err := fs.modelsGetTipsetAtTime(timeStart, false)
// 	if heightStart == 0 {
// 		heightStart = 1
// 	}
// 	if err != nil {
// 		resp.Res = respSearchError
// 		utils.Log.Errorf("get_first_tipset_after_time faild, message:%s", err.Error())
// 		return resp, nil
// 	}
// 	heightEnd, err := fs.modelsGetTipsetAtTime(timeEnd, true)
// 	if err != nil {
// 		resp.Res = respSearchError
// 		utils.Log.Errorf("get_first_tipset_after_time faild, message:%s\n", err.Error())
// 		return resp, nil
// 	}

// 	if heightStart >= heightEnd {
// 		resp.Res = common.NewResult(3, "invalid tipset_height")
// 		return resp, nil
// 	}

// 	tipset_start, err := fs.api.ChainGetTipSetByHeight(fs.ctx, abi.ChainEpoch(heightStart), types.EmptyTSK)
// 	if err != nil {
// 		utils.Log.Errorf("chain_get_tipset_by_height failed, message:%s\n", err.Error())
// 		resp.Res = respLotusApiError
// 		return resp, nil
// 	}

// 	tipset_end, err := fs.api.ChainGetTipSetByHeight(fs.ctx, abi.ChainEpoch(heightEnd), types.EmptyTSK)
// 	if err != nil {
// 		utils.Log.Errorf("chain_get_tipset_by_height failed, message:%s\n", err.Error())
// 		resp.Res = respLotusApiError
// 		return resp, nil
// 	}

// 	balance_start, err := fs.api.StateGetActor(fs.ctx, miner, tipset_start.Key())
// 	if err != nil {
// 		utils.Log.Errorf("state_get_actor failed, message:%s\n", err.Error())
// 		resp.Res = respLotusApiError
// 		return resp, nil
// 	}
// 	balance_end, err := fs.api.StateGetActor(fs.ctx, miner, tipset_end.Key())
// 	if err != nil {
// 		utils.Log.Errorf("state_get_actor failed, message:%s\n", err.Error())
// 		resp.Res = respLotusApiError
// 		return resp, nil
// 	}

// 	balanceIncreased := balance_end.Balance.Sub(balance_end.Balance.Int, balance_start.Balance.Int)

// 	resp.Res = respSuccess
// 	resp.Data = &BalanceIncreaseResp_Data{
// 		Address:           req.Address,
// 		TimeStart:         req.TimeStart,
// 		TimeEnd:           req.TimeEnd,
// 		TipsetHeightStart: heightStart,
// 		TipsetHeigthEnd:   heightEnd,
// 		BalanceIncreased:  utils.ToFilStr(balanceIncreased)}

// 	return resp, nil
// }
