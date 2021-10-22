package syncer

import (
	"context"
	"encoding/json"
	"lotus_data_sync/module"
	"strconv"
	"strings"
	"sync"

	"fmt"
	"lotus_data_sync/utils"

	"time"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/lotus/blockstore"
	"github.com/filecoin-project/lotus/chain/actors/adt"
	"github.com/filecoin-project/lotus/chain/actors/builtin/reward"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"
	cbor "github.com/ipfs/go-ipld-cbor"
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
			//fs.tipsetMinerMessagesNotifer <- modelsData.miners

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
			utils.Log.Infof("‹syncing›:sync from:%d to:[%d], count=%d, used time=%dm:%ds", beginHeight, endHeight, beginHeight-endHeight, (endTime-beginTime)/60, (endTime-beginTime)%60)
			beginTime = endTime
			beginHeight = endHeight
		}
		child = parent
	}
	//fs.tipsetMinerMessagesNotifer <- nil
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
func (fs *Filscaner) apiBlockRewards(ipset types.TipSetKey) string {
	rewardActor, err := fs.api.StateGetActor(fs.ctx, reward.Address, ipset)
	if err != nil {
		utils.Log.Errorln(err)
		return "0.0"
	}
	tbs := blockstore.NewTieredBstore(blockstore.NewAPIBlockstore(fs.api), blockstore.NewMemory())
	rewardActorState, err := reward.Load(adt.WrapStore(fs.ctx, cbor.NewCborStore(tbs)), rewardActor)
	if err != nil {
		utils.Log.Errorln(err)
		return "0.0"
	}
	// fmt.Println(rewardActorState)

	ThisEpochReward, err := rewardActorState.ThisEpochReward()
	if err != nil {
		utils.Log.Errorln(err)
		return "0.0"
	}

	//Reward, _ := strconv.ParseFloat(ThisEpochReward.String(), 64)
	// P := 5.0 * 10000000000
	// f := Reward / P / 100000000
	fil := big.Div(ThisEpochReward, big.NewInt(5))

	utils.Log.Tracef("爆块奖励为%v", types.FIL(fil).Short())
	str := types.FIL(fil).Short()
	index := strings.Index(str, "FIL")

	return types.FIL(fil).Short()[:index-1]

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
	fs.apiTipsetBlockMessagesAndReceiptsNew(child)                    //入库
	return fs.apiTipsetBlockMessagesAndReceipts(parent, childKeys[0]) //计算Gas
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
		tpstBlms.BlockRwds = fs.apiBlockRewards(tipset.Key())
	}
	tpstBlms.Tipset = tipset
	//BlockRewards

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

//var MessagMap map[int]map[string]*module.MessageInfo



func (fs *Filscaner) apiTipsetBlockMessagesAndReceiptsNew(tipset *types.TipSet) error {
	var err error
	var msg module.BlockMsg
	var blocks = tipset.Blocks()
	//uniqueMap := make(map[string]*module.MessageInfo, 0)
	blockMap := make(map[string]*module.FilscanBlock)
	height := int(tipset.Height())
	b, err := utils.Rdb16.SetNX(tipset.Key().String(), "repeat block", 3600*time.Second).Result()
	if err != nil {
		utils.Log.Errorln(err)
		return err
	}
	if !b {
		utils.Log.Tracef("tipset repeat key=%s", tipset.Key().String())
		return nil
	}
	for _, block := range blocks {
		param := &module.MinerCache{}
		if _, ok := blockMap[block.Cid().String()]; !ok {
			//在此整理block 信息
			now := time.Now().Unix()
			blockData, _ := block.Serialize()
			fsBlock := &module.FilscanBlock{
				Cid:         block.Cid().String(),
				BlockHeader: block,
				//MsgCids:     b.BlkMsgs.Cids,
				GmtCreate:   int64(block.Timestamp),
				GmtModified: now,
				Size:        int64(len(blockData)),
				BlockReward: fs.apiBlockRewards(tipset.Key()),
			}
			blockMap[block.Cid().String()] = fsBlock
			fsBlock.InsertMany(fsBlock)
			param.Addr = block.Miner.String()
			param.Height = int64(height)
			param.Timestamp = int64(block.Timestamp)
			param.TotalBlock = 1
			param.TotalRewards, _ = strconv.ParseFloat(fsBlock.BlockReward, 64)
			Info, err := fs.api.StateMinerInfo(fs.ctx, block.Miner, tipset.Key())
			if err != nil {
				utils.Log.Errorln(err)
			} else {
				utils.Log.Tracef("%+v", Info)
				param.SectorSize = Info.SectorSize.ShortString()
				if !Info.Worker.Empty() {
					param.Worker = Info.Worker.String()
				} else {
					utils.Log.Errorf("%s 没有查到 Worker id", param.Addr)
				}
				if !Info.Owner.Empty() {
					param.Owner = Info.Owner.String()

				} else {
					utils.Log.Errorf("%s 没有查到Owner id", param.Addr)
				}
				if Info.PeerId != nil {
					if Info.PeerId.Validate() == nil {
						param.PeerId = Info.PeerId.String()
					} else {
						utils.Log.Errorf("%s 没有查到peer id", param.Addr)
					}
				}
			}

		} else {
			utils.Log.Traceln("重复的block")
			continue
		}

		//如何去重
		msg.Msg = make([]*module.MessageInfo, 0)

		if m, err := fs.api.ChainGetBlockMessages(fs.ctx, block.Cid()); err == nil {
			msg.BlockCid = block.Cid().String()
			msg.Crated = time.Now().Unix()
			msg.Height = int64(block.Height)
			for _, v := range m.BlsMessages {
				minfo := module.MessageInfo{}
				minfo.Cid = v.Cid().String()
				minfo.From = v.From.String()
				minfo.To = v.To.String()
				minfo.Version = v.Version
				minfo.GasFeeCap = v.GasFeeCap.Int64()
				minfo.GasLimit = v.GasLimit
				minfo.Method = int(v.Method)
				minfo.Nonce = v.Nonce
				minfo.Params = v.Params
				minfo.GasPremium = v.GasPremium.Int64()
				minfo.Value = v.Value.Int64()
				minfo.Timestamp = int64(block.Timestamp)
				msg.Msg = append(msg.Msg, &minfo)

			}
			for _, v := range m.SecpkMessages {
				minfo := module.MessageInfo{}
				minfo.Cid = v.Cid().String()
				minfo.From = v.Message.From.String()
				minfo.To = v.Message.To.String()
				minfo.Version = v.Message.Version
				minfo.GasFeeCap = v.Message.GasFeeCap.Int64()
				minfo.GasLimit = v.Message.GasLimit
				minfo.Method = int(v.Message.Method)
				minfo.Nonce = v.Message.Nonce
				minfo.Params = v.Message.Params
				minfo.GasPremium = v.Message.GasPremium.Int64()
				minfo.Value = v.Message.Value.Int64()
				minfo.Timestamp = int64(block.Timestamp)
				msg.Msg = append(msg.Msg, &minfo)

			}
		}
		//入库
		var sy module.SyncInfo
		sy.BlockCid = block.Cid().String()
		sy.Height = int64(height)
		sy.Created = time.Now().Unix()
		utils.Log.Tracef(" block_cid=%s height=%d", block.Cid(), height)
		if err = new(module.SyncInfo).InsertOne(sy); err != nil {
			//utils.Log.Errorln(err)
			continue
		}
		utils.Log.Tracef(" block_cid=%s height=%d", block.Cid(), height)

		if err = new(module.BlockMsg).InsertMany(msg); err != nil {
			//utils.Log.Errorln(err)
			continue
		}
		go fs.RedisCache(*param)
		//new(module.Miner).Upsert(param)
	}
	//if _, ok := mm[heightPre-1]; ok {
	//delete(MessagMap, height-1)
	//delete(BlockMap, height-1)

	//}
	return err
}

func (fs *Filscaner) apiTipsetBlockMessagesAndReceiptsGasUsage(tipset *types.TipSet, childCid cid.Cid) error {
	var err error
	var msg module.BlockMsg
	var blocks = tipset.Blocks()
	blockMapGas :=make(map[string]*module.FilscanBlock)
	//height := int(tipset.Height())
	
	key := fmt.Sprintf("gas_%s", tipset.Key().String())
	b, err := utils.Rdb16.SetNX(key, "repeat block", 3600*time.Second).Result()
	if err != nil {
		utils.Log.Errorln(err)
		return err
	}
	if !b {
		utils.Log.Tracef("tipset repeat key=%s", key)
		return nil
	}

	for _, block := range blocks {
		param := module.MinerCache{}
		if _, ok := blockMapGas[block.Cid().String()]; !ok {
			//在此整理block 信息
			blockMapGas[block.Cid().String()] = &module.FilscanBlock{}
			param.Addr = block.Miner.String()

		} else {
			continue
		}

		Messages, err := fs.api.ChainGetParentMessages(fs.ctx, childCid)
		if err != nil {
			utils.Log.Errorf("err ChainGetParentMessages:%v", err)
			continue
		}
		//utils.Log.Tracef("block_cid=%s 消息个数为%d", childCid, len(Messages))
		Receipts, err := fs.api.ChainGetParentReceipts(fs.ctx, childCid)
		if err != nil {
			utils.Log.Errorf("ChainGetParentReceipts:%v", err)
			continue
		}
		//utils.Log.Tracef("block_cid=%s Receipts%d", childCid, len(Receipts))
		receipt_ref := make(map[string]*types.MessageReceipt)
		for index, receipt := range Receipts {
			receipt_ref[Messages[index].Cid.String()] = receipt
		}
		//utils.Log.Traceln(receipt_ref)
		//如何去重
		//msg.Msg = make([]*module.MessageInfo, 0)

		if m, err := fs.api.ChainGetBlockMessages(fs.ctx, block.Cid()); err == nil {
			msg.BlockCid = block.Cid().String()
			msg.Crated = time.Now().Unix()
			msg.Height = int64(block.Height)
			//msg.Msg = append(msg.Msg, m.BlsMessages...)
			for _, v := range m.BlsMessages {
				minfo := module.MessageInfo{}
				minfo.Cid = v.Cid().String()

				if _, ok := receipt_ref[minfo.Cid]; !ok {
					//utils.Log.Errorf("block_cid=%s msg_cid=%s GasUsage=0", block.Cid().String(), minfo.Cid)
				} else {
					//minfo.GasUsage = receipt_ref[minfo.Cid].GasUsed
					param.TotalGas += receipt_ref[minfo.Cid].GasUsed
				}
				// if _, ok := MessagMap[height][minfo.Cid]; !ok {
				// 	MessagMap[height][minfo.Cid] = &minfo
				//msg.Msg = append(msg.Msg, &minfo)
				// }
			}
			for _, v := range m.SecpkMessages {
				minfo := module.MessageInfo{}
				minfo.Cid = v.Cid().String()
				if _, ok := receipt_ref[minfo.Cid]; !ok {
					//utils.Log.Errorf("block_cid=%s msg_cid=%s GasUsage=0", block.Cid().String(), minfo.Cid)
				} else {
					//minfo.GasUsage = receipt_ref[minfo.Cid].GasUsed
					param.TotalGas += receipt_ref[minfo.Cid].GasUsed
				}

				// if _, ok := MessagMap[height][minfo.Cid]; !ok {
				// 	MessagMap[height][minfo.Cid] = &minfo
				//msg.Msg = append(msg.Msg, &minfo)
				// }

			}
		}
		//入库
		go fs.RedisCache(param)
		//new(module.Miner).Upsert(param)
	}
	//if _, ok := mm[heightPre-1]; ok {
	//delete(MessagMap, height-1)
	//delete(BlockMapGas, height-1)

	//}
	return err
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

//一天写一次mongodb,有新的miner 就写一个空的结构到mongo
const hsetNxKey = "hashminers"

func (fs *Filscaner) RedisCache(param module.MinerCache) {
	body, _ := json.Marshal(param)
	if b, err := utils.Rdb16.SetNX(param.Addr, string(body), 0).Result(); err != nil {
		utils.Log.Errorln(err)
	} else if b {
		//新的miner
		utils.Rdb16.HSetNX(hsetNxKey, param.Addr, param.Tag) //保存所有的miner
	} else {

		//存在更新字段
		body, err := utils.Rdb16.Get(param.Addr).Bytes()
		if err != nil {
			utils.Log.Errorln(err)
			return
		}
		result := &module.MinerCache{}
		if err = json.Unmarshal(body, result); err != nil {
			utils.Log.Errorln(err)
			return
		}
		if param.Height != 0 {

			result.Height = param.Height
		}
		if param.PeerId != "" {

			result.PeerId = param.PeerId
		}
		if param.Tag != "" {

			result.Tag = param.Tag
		}
		if param.Timestamp != 0 {

			result.Timestamp = param.Timestamp
		}
		result.TotalBlock += param.TotalBlock
		result.TotalGas += param.TotalGas
		result.TotalRewards += param.TotalRewards
		resultBytes, _ := json.Marshal(result)
		utils.Rdb16.Set(param.Addr, string(resultBytes), 0)
	}
}
