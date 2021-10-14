package syncer

import (
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/store"
	"github.com/filecoin-project/lotus/chain/types"
	"lotus_data_sync/utils"
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
		utils.Log.Tracef("---------------------------------type=%v height=%d ---------------------------------", header.Type, header.Val.Height())
		// if header.Type == store.HCCurrent {
		// 	//
		// 	//utils.Log.Tracef("---------------------------------这特么的是什么情况呢？tore.HCCurrent 当前高度为 %d---------------------------------", header.Val.Height())

		// }
		if header.Type == store.HCApply {
			//utils.Log.Traceln("这特么的是什么情况呢？ store.HCApply")

			// fs.handleApplyTippet(header.Val, nil)
			// fs.lastApplyTippet = header.Val
			fs.HandleLotusData(header.Val,nil)
		}
	}
}
func (fs *Filscaner) HandleLotusData(child, parent *types.TipSet) {
	if child == nil {
		return
	}
	var err error
	if parent == nil || child.Parents().String() != parent.Key().String() {
		if parent, err = fs.api.ChainGetTipSet(fs.ctx, child.Parents()); err != nil {
			utils.Log.Errorf("error, get tipset(%d,%s) failed, message:%s",
				parent.Height()-1, parent.Parents().String(), err.Error())
			return
		}
	}

	blockMessage, err := fs.buildPersistenceData(child, parent)
	if err != nil {
		utils.Log.Errorf(" build_persistence_data(child:%d, parent:%d) failed, message:%s",child.Height(), parent.Height(), err.Error())
		return
	}
	utils.Log.Tracef(" build_persistence_data(child:%d, parent:%d) ", child.Height(), parent.Height())
	if err := blockMessage.modelsUpsert(); err != nil {
		utils.Log.Errorf("error, Tipset_block_messages.models_upsert failed, message:%s",err.Error())
		return
	}

}

func (fs *Filscaner) handleFirstApplyTippet(child, parent *types.TipSet) {
	utils.Log.Traceln("handleFirstApplyTippet")
	if child == nil {
		return
	}
	var err error
	if parent == nil || child.Parents().String() != parent.Key().String() {
		if parent, err = fs.api.ChainGetTipSet(fs.ctx, child.Parents()); err != nil {
			utils.Log.Errorf("error, get tipset(%d,%s) failed, message:%s",
				parent.Height()-1, parent.Parents().String(), err.Error())
			return
		}
	}
	fs.syncTipsetCacheFallThrough(child, parent)
	fs.handleApplyTippet = fs.handleSecondApplyTippet
}

func (fs *Filscaner) handleSecondApplyTippet(child, this_is_nil_value_do_not_use *types.TipSet) {
	utils.Log.Traceln("handleSecondApplyTippet")
	if fs.lastApplyTippet.Height() == child.Height() {
		// p1, _ := fs.api.ChainGetTipSet(fs.ctx, child.Parents())
		// p2, _ := fs.api.ChainGetTipSet(fs.ctx, fs.last_appl_tipset.Parents())
		// fs.Printf("p1 equals p2 = %v\n", p1.Equals(p2))
		return
	}

	parent, err := fs.api.ChainGetTipSet(fs.ctx, child.Parents())
	if err != nil {
		utils.Log.Errorf("error, get child(%d,%s) failed, message:%s",
			child.Height()-1, child.Parents().String(), err.Error())
		return
	}
	//go fs.ParamTemp(child)

	blockMessage, err := fs.buildPersistenceData(child, parent)
	if err != nil {
		utils.Log.Errorf("error, build_persistence_data(child:%d, parent:%d) failed, message:%s",
			child.Height(), parent.Height(), err.Error())
		return
	}

	if ftsp := fs.tipsetsCache.Front(); ftsp != nil && child.Height() <= ftsp.Height() {
		utils.Log.Errorf("‹forked›‹›‹›at child:‹%d›, current head is:‹%d›",
			child.Height(), ftsp.Height())
	}

	if blockMessage = fs.tipsetsCache.pushFront(blockMessage); blockMessage != nil {
		fs.handleSafeTipset(blockMessage)
	}
}

func (fs *Filscaner) handleFirstSafeTipset(blockmessage *TipsetBlockMessages) {
	utils.Log.Traceln("handleFirstSafeTipset")
	//TODO message block  tipsets
	utils.Log.Tracef("height=%d", blockmessage.Tipset.Height())
	if err := blockmessage.modelsUpsert(); err != nil {
		utils.Log.Errorf("error, Tipset_block_messages.models_upsert failed, message:%s",
			err.Error())
		return
	}

	// fs.taskSyncToGenesis(blockmessage.Tipset)
	// fs.handleSafeTipset = fs.handleSecodSafeTipset

	//TODO 处理新增的 block
	//tmpFsTipset, tmpFsBlocks, tmpFsMsgs, tmpMinerMsgs := blockmessage.buildModelsData()
	//
	//tempf := func(param []*module.FilscanBlock) {
	//	for _, v := range param {
	//		var gas module.Gas
	//		gas.v.BlockReward
	//
	//	}
	//modelsData.tipsets = append(modelsData.tipsets, tmpFsTipset)
	//modelsData.blocks = append(modelsData.blocks, tmpFsBlocks[:]...)
	//modelsData.messages = append(modelsData.messages, tmpFsMsgs[:]...)
	//modelsData.miners[index] = tmpMinerMsgs
	//}
	//go tempf(tmpFsBlocks)
}

func (fs *Filscaner) handleSecodSafeTipset(in *TipsetBlockMessages) {
	utils.Log.Traceln("handleSecodSafeTipset")
	if in == nil {
		utils.Log.Errorln("debug 出去了")
		return
	}
	var tbml = &TipsetBlockMessageList{
		TipsetBlockMessages: []*TipsetBlockMessages{in}}

	var err error

	if fs.syncedTipsetPathList.insertHeadChild(in.Tipset) {
		if err = fs.syncedTipsetPathList.modelsUpsertFront(true); err != nil {
			utils.Log.Errorf("error, models_upsert_front failed, message:%s", err.Error())
			return
		}

		modelsData := tbml.buildModelsData()
		utils.Log.Traceln("最新区块入库 height=", in.Tipset.Height())
		fs.tipsetMinerMessagesNotifer <- modelsData.miners

		if err = modelsData.modelsUpsert(); err != nil {
			utils.Log.Errorf("error, Tipset_block_message_list.upsert failed, message:%s", err.Error())
			return
		}
		utils.Log.Traceln("最新区块入库 mongodb ok height=", in.Tipset.Height())
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
