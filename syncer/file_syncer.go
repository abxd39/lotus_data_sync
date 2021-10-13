package syncer

import (
	"context"
	"lotus_data_sync/module"

	"lotus_data_sync/utils"
	"fmt"
	"github.com/filecoin-project/go-state-types/abi"

	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"

	"time"
)

// TODO: use fs.api.StateChangedActors(),
//  to sync miner state change information
func (fs *Filscaner) syncToGenesis(from *types.TipSet) (*types.TipSet, error) {
	const maxSize = 25
	var child = from
	var parent *types.TipSet = nil
	var tbml = &TipsetBlockMessageList{}
	var tpstBlms *TipsetBlockMessages
	var err error
	var cidsSize = len(child.Parents().Cids())
	var beginTime = time.Now().Unix()
	var beginHeight = child.Height()
	utils.Log.Traceln("data manager for syncToGenesis")
	for cidsSize != 0 { // genesis case 'bafy2bzaceaxm23epjsmh75yvzcecsrbavlmkcxnva66bkdebdcnyw3bjrc74u'
	parent, err = fs.api.ChainGetTipSet(fs.ctx, child.Parents())
	if err != nil {
		utils.Log.Errorf("error:sync_to_genesis will exit, message:%s", err.Error())
		return nil, err
	}
	//utils.Log.Traceln("heitht =",parent.Height())
		if !fs.syncedTipsetPathList.insertTailParent(parent) {
			return parent, fmt.Errorf("error, sync to genesis, isn't a parents(%d), it's impossible", parent.Height())
		}
		tpstBlms, err = fs.buildPersistenceData(child, parent)
		if err != nil {
			return parent, err
		}
		tbml.TipsetBlockMessages = append(tbml.TipsetBlockMessages, tpstBlms)
		// todo: just merge but haven't write to store, but,
		//  the chain-notify may cause a writing..
		mergedPath := fs.syncedTipsetPathList.tryMergeFrontWithNext(true)
		cidsSize = len(parent.Parents().Cids())
		if len(tbml.TipsetBlockMessages) >= maxSize || mergedPath != nil || cidsSize == 0 {
			utils.Log.Infof("‹syncing›:models_upsert(cached_tipset_size=%d)", len(tbml.TipsetBlockMessages))
			if mergedPath != nil {
				utils.Log.Infof("‹syncing›:synced range is merged:(%d, %d)", mergedPath.Head.Height, mergedPath.Tail.Height)
			}
			modelsData := tbml.buildModelsData()
			fs.tipsetMinerMessagesNotifer <- modelsData.miners

			if err = modelsData.modelsUpsert(); err != nil {
				utils.Log.Errorf("error, tipset_block_message_list upsert failed, message:%s", err.Error())
				return parent, err
			}
			utils.Log.Infof("‹syncing›:sync tipset(%d) mongodb ok", parent.Height())
			if err := fs.syncedTipsetPathList.modelsUpsertFront(true); err != nil {
				utils.Log.Errorf("error, models_upsert_front failed, message:%s", err.Error())
				return parent, err
			}
			if mergedPath != nil {
				if err := mergedPath.modelsDel(); err != nil {
					utils.Log.Errorf("error, merged_path.models_del failed, message:%s", err.Error())
					return parent, err
				}
				if parent, err = fs.apiTipset(fs.syncedTipsetPathList.frontTail().Key); err != nil {
					utils.Log.Errorf("api_tipset failed, message:%s", err.Error())
					return nil, err
				} else {
					cidsSize = len(parent.Parents().Cids())
				}
			}
			tbml.TipsetBlockMessages = tbml.TipsetBlockMessages[:0]
			
		}
		utils.Log.Infof("‹syncing›:sync tipset(%d) finished", parent.Height())
		endHeight := parent.Height()
		if beginHeight-endHeight > 250 {
			endTime := time.Now().Unix()
			utils.Log.Infof("‹syncing›:sync from:%d to:[%d], count=%d, used time=%dm:%ds",beginHeight, endHeight, beginHeight-endHeight,(endTime-beginTime)/60, (endTime-beginTime)%60)
			beginTime = endTime
			beginHeight = endHeight
		}
		child = parent
	}
	fs.tipsetMinerMessagesNotifer <- nil
	return child, err
}

func (fs *Filscaner) syncTipsetCacheFallThrough(child, parent *types.TipSet) (*types.TipSet, error) {
	var err error
	var blockMessage *TipsetBlockMessages
	for !fs.tipsetsCache.full() {
		if blockMessage, err = fs.buildPersistenceData(child, parent); err != nil {
			return nil, err
		}
		blockMessage.buildModelsData() //消息处理

		fs.tipsetsCache.push_back(blockMessage)

		child = parent
		if parent, err = fs.api.ChainGetTipSet(fs.ctx, child.Parents()); err != nil {
			return nil, err
		}
	}
	return parent, nil
}

//upload apiBlockRewards
func (fs *Filscaner) apiBlockRewards(tipset types.TipSetKey) string {
	// rewardActor, err := fs.api.StateGetActor(fs.ctx, reward.Address, tipset)
	// if err != nil {
	// 	return "0.0"
	// }
	// rewardActorState, err := reward.Load(cwutil.NewAPIIpldStore(fs.ctx, fs.api), rewardActor)
	// if err != nil {
	// 	return "0.0"
	// }
	// // fmt.Println(rewardActorState)

	// ThisEpochReward, err := rewardActorState.ThisEpochReward()
	// if err != nil {
	// 	return "0.0"
	// }

	// Reward, _ := strconv.ParseFloat(ThisEpochReward.String(), 64)
	// P := 5.0 * 10000000000
	// f := Reward / P / 100000000
	return fmt.Sprintf("%.16f", 1.000)

}
func (fs *Filscaner) syncTipsetWithRange(last_ *types.TipSet, head_height, foot_height uint64) (*types.TipSet, error) {
	utils.Log.Infof("do_sync_lotus, from:%d, to:%d", head_height, foot_height)

	headTipset, err := fs.api.ChainGetTipSetByHeight(fs.ctx, abi.ChainEpoch(head_height), last_.Key())
	footTipset, err := fs.api.ChainGetTipSetByHeight(fs.ctx, abi.ChainEpoch(foot_height), last_.Key())
	tipsetList, err := fs.api.ChainGetPath(fs.ctx, footTipset.Key(), headTipset.Key())
	if err != nil {
		utils.Log.Errorf("error:api.chain_get_path failed, message:%s", err.Error())
		return nil, err
	}

	if last_ == nil {
		if last_, err = fs.apiChildTipset(tipsetList[0].Val); err != nil {
			utils.Log.Errorf("error:api_child_tipset failed, message:%s", err.Error())
			return nil, err
		}
	}

	tbml := &TipsetBlockMessageList{}
	for _, t := range tipsetList {
		tipset := t.Val
		tbm, err := fs.buildPersistenceData(last_, tipset)
		if err != nil {
			return nil, err
		}

		tbml.TipsetBlockMessages = append(tbml.TipsetBlockMessages, tbm)

		last_ = tipset
	}

	return last_, tbml.buildModelsData().modelsUpsert()
}

func (fs *Filscaner) buildPersistenceData(child, parent *types.TipSet) (*TipsetBlockMessages, error) {
	if child.Parents().String() != parent.Key().String() {
		return nil, fmt.Errorf("child(%d, %s).parentkey(%s)!=tipset(%d).key(%s)",
			child.Height(), child.Key().String(), child.Parents().String(),
			parent.Height(), parent.Key().String())
	}

	childKeys := child.Key().Cids()

	if len(childKeys) == 0 {
		return nil, fmt.Errorf("tipset(%d, %s) have no blocks",
			child.Height(), child.Key().String())
	}
	return fs.apiTipsetBlockMessagesAndReceipts(parent, childKeys[0])
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

//消息参数解析
func (fs *Filscaner) ParamTemp(parent *types.TipSet, string2 string) {
	var blocks = parent.Blocks()
	//var cids = parent.Cids()
	//clen := len(cids)
	//blen := len(blocks)
	//utils.Log.Tracef("blocks 长度为%d cids 的长度为=%d height=%d", blen, clen, parent.Height())
	for _, block := range blocks {
		blockCid := block.Cid().String()
		//utils.Log.Tracef("block_cid=%s", blockCid)
		if m, err := fs.api.ChainGetBlockMessages(fs.ctx, block.Cid()); err == nil {
			for _, v := range m.BlsMessages {
				fs.Temp(v, uint64(block.Height), blockCid, string2)
			}
			for _, v := range m.SecpkMessages {
				fs.Temp(&v.Message, uint64(block.Height), blockCid, string2)
				//if v.Message.Method == 0 {
				//	continue
				//}
				//MethodName := ""
				//if _, method, err := ParseActorMessage(&v.Message); err == nil {
				//
				//	MethodName = method.Name
				//}
				//utils.Log.Tracef("cid=%s MethodName=%s block.Height=%d",v.Cid().String(),MethodName,uint64(block.Height))
				//new(InterfaceParam).DecoderMessageParam(v.Cid().String(), MethodName, uint64(block.Height), true)
			}

		} else {
			utils.Log.Errorf("ChainGetBlockMessages err cid=%s", block.Cid().String())
			continue
		}
	}
}

func (fs *Filscaner) Temp(msg *types.Message, height uint64, blackcid, tp string) {
	// var p = &InterfaceParam{}
	// //if utils.TipMark == false {
	// //	utils.Log.Errorln("lotus 同步异常！！")
	// //	return
	// //}
	// //utils.Log.Tracef("%+v", msg)
	// addr := msg.To.String()
	// if !p.IsOk(addr) {
	// 	return//过滤掉不需要监控的节点
	// }
	// cid := msg.Cid().String()
	// // if msg.Method == 25 || msg.Method == 26 {
	// // 	utils.Log.Tracef("\n⬇---------------------------------\n%s\n%s\ntype=%s\nmsg.To=%s  \n\n\n", blackcid, cid, tp, msg.To)
	// // }
	// if msg.Method == 0 {
	// 	return
	// }
	// MethodName := ""
	// if _, method, err := ParseActorMessage(msg); err == nil {
	// 	MethodName = method.Name
	// }
	// if MethodName == "" {
	// 	if msg.Method == 25 {
	// 		MethodName = "PreCommitSectorBatch"
	// 	}
	// 	if msg.Method == 26 {
	// 		MethodName = "ProveCommitAggregate"
	// 	}
	// }

	// ctx := context.TODO()
	// result, err := fs.api.StateDecodeParams(ctx, msg.To, msg.Method, msg.Params, types.EmptyTSK)
	// if err != nil {
	// 	utils.Log.Errorln(err)
	// 	return
	// }
	// mjson, _ := json.Marshal(result)
	// if msg.Method == 25 || msg.Method == 26 {
	// 	//utils.Log.Traceln(string(mjson))
	// 	//utils.Log.Tracef("cid=%s MethodName=%s block.Height=%d type=%s", cid, MethodName, height, tp)
	// }

	// if !p.MessageType(int(msg.Method)) { //过滤不要的消息
	// 	return
	// }
	// hei, err := utils.Initconf.Int64("BlockHeight")
	// if err != nil {
	// 	utils.Log.Errorln(err)
	// 	return
	// }

	// if hei > int64(height) { //历史高度不用处理
	// 	return
	// }

	// //TODO 在此过滤 用户
	// //addr := msg.To.String()
	// //utils.Log.Tracef("cid=%s MethodName=%s block.Height=%d type=%s", cid, MethodName, height, tp)
	// //if p.IsOk(addr) {
	// //redis 去重
	// key := fmt.Sprintf("messag:cid:%s", cid)
	// if redisdb.SetNXDB1(key, "message去重", 48*3600) { //上层重复调用 过滤重复消息
	// 	//utils.Log.Tracef("cid=%s addr=%s msg.MethodName =%s method=%d block.Height=%d type=%s", cid, addr, MethodName, msg.Method, height, tp)
	// 	go p.NotifyQyChat(msg.To.String(), cid, MethodName, mjson) //快速响应
	// }
	// //}
	return

}

func (fs *Filscaner) handleFirstSafeTipset(blockmessage *TipsetBlockMessages) {
	utils.Log.Traceln("handleFirstSafeTipset")
	//TODO message block  tipsets
	utils.Log.Tracef("height=%d",blockmessage.Tipset.Height())
	if err := blockmessage.modelsUpsert(); err != nil {
		utils.Log.Errorf("error, Tipset_block_messages.models_upsert failed, message:%s",
			err.Error())
		return
	}
	fs.taskSyncToGenesis(blockmessage.Tipset)
	fs.handleSafeTipset = fs.handleSecodSafeTipset
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
		utils.Log.Traceln("最新区块入库 height=",in.Tipset.Height())
		fs.tipsetMinerMessagesNotifer <- modelsData.miners

		if err = modelsData.modelsUpsert(); err != nil {
			utils.Log.Errorf("error, Tipset_block_message_list.upsert failed, message:%s", err.Error())
			return
		}
			utils.Log.Traceln("最新区块入库 mongodb ok height=",in.Tipset.Height())
	}
}

//消息解析
func (fs *Filscaner) apiTipsetBlockMessagesAndReceipts(tipset *types.TipSet, childCid cid.Cid) (*TipsetBlockMessages, error) {
	var tpstBlms = &TipsetBlockMessages{}
	var err error
	var blocks = tipset.Blocks()
	for _, block := range blocks {
		if message, err := fs.api.ChainGetBlockMessages(fs.ctx, block.Cid()); err == nil {
			blmsg := &BlockMessage{block, message}
			tpstBlms.BlockMsgs = append(tpstBlms.BlockMsgs, blmsg)

		} else {
			return nil, err
		}
	}
	tpstBlms.Tipset = tipset
	//BlockRewards
	tpstBlms.BlockRwds = fs.apiBlockRewards(tipset.Key())
	tpstBlms.Messages, err = fs.api.ChainGetParentMessages(fs.ctx, childCid)
	if err != nil {
		utils.Log.Errorf("err ChainGetParentMessages:%v", err)
		return nil, err
	}
	tpstBlms.Receipts, err = fs.api.ChainGetParentReceipts(fs.ctx, childCid)
	if err != nil {
		tpstBlms.Receipts = nil
		err = nil
		utils.Log.Errorf("ChainGetParentReceipts:%v", err)
	}
	tpstBlms.buildModelsData()
	return tpstBlms, err

}

//查询block表 所有的区块奖励 统计到统计表
func (fs *Filscaner) SelectBlockStatistics() {
	//查询表写入数据
	//根据连上所有矿工的地址来入裤
	utils.Log.Traceln("--------------查询数据库统计到统计表----> 开始-----------", time.Now().Format(utils.TimeString))
	ctx := context.TODO()
	//acts, err := fs.api.StateListActors(ctx, types.EmptyTSK)
	//if err != nil {
	//	utils.Log.Errorln(err)
	//	return
	//}
	tp, err := fs.api.ChainGetGenesis(ctx)
	if err != nil {
		utils.Log.Errorln(err)
		return
	}
	address, err := fs.api.StateListMiners(ctx, tp.Key())
	if err != nil {
		utils.Log.Errorln(err)
		return
	}
	index := 0
	for _, addr := range address {
		result, total, err := module.GetBlockListByMiner([]string{addr.String()}, index, 1000)
		if err != nil {
			utils.Log.Errorln(err)
			break
		}
		var statis module.BlockHistory
		for _, b := range result {
			if b.BlockHeader.Miner == "" {
				continue
			}
			statis.BlockReword = b.BlockReword
			statis.BlockCid = b.Cid
			statis.Miner = b.BlockHeader.Miner
			blockStatistics(&statis)
		}
		index += 1000
		if index > total {
			break
		}
	}
	utils.Log.Traceln("--------------查询数据库统计到统计表----> 结束-----------", time.Now().Format(utils.TimeString))
}
