package syncer

import (
	
	"lotus_data_sync/utils"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/store"
	
	"time"
)

func (fs *Filscaner) displayNotifi(header *api.HeadChange) {
	parent, _ := fs.api.ChainGetTipSet(fs.ctx, header.Val.Parents())
	head, _ := fs.api.ChainHead(fs.ctx)

	utils.Log.Infof("new notify:>>>>%s(%d)<<<< : %s >> parent(%d,%s),chain head(%d, %s)",
		header.Type, header.Val.Height(), header.Val.Key().String(),
		parent.Height(), parent.Key().String(),
		head.Height(), head.Key())
}

func (fs *Filscaner) handleNewHeaders(headers []*api.HeadChange) {
	for _, header := range headers {
		if header == nil {
			continue
		}
		//fs.displayNotifi(header)
		utils.Log.Tracef("---------------------------------%v---------------------------------", time.Now().Format(utils.TimeString))
		// if header.Type == store.HCCurrent {
		// 	//
		// 	//utils.Log.Tracef("---------------------------------这特么的是什么情况呢？tore.HCCurrent 当前高度为 %d---------------------------------", header.Val.Height())
		
		// }
		if header.Type == store.HCApply {
			//utils.Log.Traceln("这特么的是什么情况呢？ store.HCApply")
	
			fs.handleApplyTippet(header.Val, nil)
			fs.lastApplyTippet = header.Val
		}
	}
}

func (fs *Filscaner) loopHandleMessages() {
	ticker := time.NewTicker(time.Second * 90)
	utils.Log.Tracef("debug toUpdateMinerSize= %v toUpdateMinerIndex=%v", fs.toUpdateMinerSize, fs.toUpdateMinerIndex)
	for {
		select {
		case minerMessagesList, ok := <-fs.tipsetMinerMessagesNotifer:
			if !ok {
				utils.Log.Errorf("messge notifier is closed stop handle message")
				return
			}

			isNil := false
			if minerMessagesList == nil { // if syncor reached genesis, it send a 'nil' message
				isNil = true
				if tipsetMinerMessages, err := fs.listGenesisMiners(); err != nil {
					utils.Log.Errorf("list_genesis_miners failed, message:%s\n", err.Error())
				} else {
					minerMessagesList = []*TipsetMinerMessages{tipsetMinerMessages}
				}
			}
			for _, minerMessages := range minerMessagesList {
				utils.Log.Infof("handle storage_miner_actor messages at tipset:%d, miner count:%d",
					minerMessages.tipset.Height(), len(minerMessages.miners))
				for address, _ := range minerMessages.miners {
					fs.handleStorageMinerMessage(minerMessages.tipset, address)
				}
			}
			if isNil {
				fs.doUpsertMiners()
				if syncedPath := fs.syncedTipsetPathList.frontSyncedPath(); syncedPath != nil {
					utils.Log.Infof(`successed handled genesis tipset messages, current synced state : head.height:%d, tail.height:%d * successed handled genesis tipset messages `,
						syncedPath["head.height"], syncedPath["tail.height"])
				}
			}
		case <-ticker.C:
			fs.doUpsertMiners()
		case <-fs.ctx.Done():
			utils.Log.Traceln("ctx.done, exit loop_handle_messages")
			return
		}
	}
}

//refresh miner state  for minerStateChan
func (fs *Filscaner) loopHandleRefreshMinerState() {
	// 5 分钟触发一次刷新最新状态..
	//ticker := time.NewTicker(time.Second * 300)
	//level2Cache := make(map[string]*module.MinerStateAtTipset)
	// for {
	// 	select {
	// 	case minerState, ok := <-fs.minerStateChan:
	// 		if !ok {
	// 			utils.Log.Traceln("message notifier is closed stop handle message")
	// 			return
	// 		}
	// 		if true {
	// 			fs.waitGroup.Add(1)
	// 			go func() {
	// 				fs.handleMinerState(minerState)
	// 				fs.waitGroup.Done()
	// 			}()
	// 		} else {
	// 			if miner, exist := level2Cache[minerState.MinerAddr]; !exist {
	// 				level2Cache[minerState.MinerAddr] = miner
	// 			}
	// 			if len(level2Cache) > 20 {
	// 				// todo: 批量更新
	// 				fs.handleMinerState(minerState)
	// 				for k, _ := range level2Cache {
	// 					delete(level2Cache, k)
	// 				}
	// 			}
	// 		}
	// 	case <-fs.ctx.Done():
	// 		utils.Log.Infof("ctx.done, exit loop_handle_messages")
	// 		return
	// 	}
	// }
}

func (fs *Filscaner) loopInitBlockRewards() {
	// tipset, err := fs.api.ChainHead(fs.ctx)
	// if err != nil {
	// 	return
	// }

	// headBlockRewards, err := modelsBlockRewardHead()
	// if err != nil {
	// 	return
	// }

	// headHeight := tipset.Height()
	// totalRewards := big.NewInt(0).Sub(TotalRewards, headBlockRewards.ReleasedRewards.Int)

	// bulkSize := 20
	// upsertRewards := make([]*ModelsBlockReward, bulkSize)
	// offset := 0

	// releasedRewards := headBlockRewards.ReleasedRewards.Int

	// // 每20 * 25个height间隔保存一次
	// for i := headBlockRewards.Height; i < uint64(headHeight); i++ {
	// 	rewards := SelfTipsetRewards(totalRewards)
	// 	totalRewards.Sub(totalRewards, rewards)
	// 	releasedRewards.Add(releasedRewards, rewards)
	// 	if i%25 == 0 { //每25个height间隔一个数据库记录
	// 		upsertRewards[offset] = &ModelsBlockReward{
	// 			Height:          i,
	// 			ReleasedRewards: &module.BsonBigint{Int: releasedRewards}}
	// 		offset++
	// 		if offset == bulkSize {
	// 			err := modelsBulkUpsertBlockReward(upsertRewards, offset-1)
	// 			if err != nil {
	// 				// TODO: handle error
	// 				utils.Log.Errorf("modelsBulkUpsertBlockReward is failed,err:%v", err)
	// 			}
	// 			offset = 0
	// 			time.Sleep(time.Millisecond * 500)
	// 		}
	// 	}
	// }
	// if offset != 0 {
	// 	err := modelsBulkUpsertBlockReward(upsertRewards, offset-1)
	// 	if err != nil {
	// 	}
	// }
}
