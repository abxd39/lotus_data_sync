package syncer

import (
	"lotus_data_sync/module"
	"lotus_data_sync/utils"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/globalsign/mgo"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"time"
)

type BlockMessage struct {
	Block   *types.BlockHeader
	BlkMsgs *api.BlockMessages
}

type TipsetBlockMessages struct {
	Tipset    *types.TipSet
	BlockRwds string
	Messages  []api.Message
	Receipts  []*types.MessageReceipt
	BlockMsgs []*BlockMessage

	isCached bool

	fsTipset *module.FilscanTipSet
	fsBlocks []*module.FilscanBlock
	fsMsgs   []*module.FilscanMsg
	fsMiners *TipsetMinerMessages

	fsBlockMessage []*module.BlockAndMsg
}

type TipsetBlockMessageList struct {
	TipsetBlockMessages []*TipsetBlockMessages
}

func (b *BlockMessage) fsBlock() *module.FilscanBlock {
	if b.Block == nil {
		return nil
	}
	now := time.Now().Unix()
	blockData, _ := b.Block.Serialize()
	fsBlock := &module.FilscanBlock{
		Cid:         b.Block.Cid().String(),
		BlockHeader: b.Block,
		MsgCids:     b.BlkMsgs.Cids,
		GmtCreate:   now,
		GmtModified: now,
		Size:        int64(len(blockData)),
	}
	return fsBlock
}

func (b *BlockMessage) fsMessages() []*module.FilscanMsg {
	if b.BlkMsgs == nil {
		return nil
	}
	var fsMsgList []*module.FilscanMsg
	//TODO 告警功能在次确认是何种消息类型
	now := time.Now().Unix()
	for _, v := range b.BlkMsgs.BlsMessages {
		data, _ := v.Serialize()
		fs_msg := &module.FilscanMsg{
			//解析参数
			Message:       *v,
			Cid:           v.Cid().String(),
			BlockCid:      b.Block.Cid().String(),
			RequiredFunds: v.RequiredFunds(),
			Size:          int64(len(data)),
			Height:        uint64(b.Block.Height),
			MsgCreate:     b.Block.Timestamp,
			GmtCreate:     now,
			GmtModified:   now}

		if v.Method == 0 {
			fs_msg.MethodName = "Transfer"
		} else {
			// if actor, method, err := ParseActorMessage(v); err == nil {
			// 	fs_msg.ActorName = actor.Name
			// 	fs_msg.MethodName = method.Name
			// }
		}

		//fs_msg.MessageParams = new(InterfaceParam).DecoderMessageParam(fs_msg.Cid, fs_msg.MethodName, fs_msg.Height, true)
		fsMsgList = append(fsMsgList, fs_msg)
	}
	for _, secp := range b.BlkMsgs.SecpkMessages {
		data, _ := secp.Message.Serialize()
		fs_msg := &module.FilscanMsg{
			Message:       secp.Message,
			Cid:           secp.Cid().String(),
			BlockCid:      b.Block.Cid().String(),
			RequiredFunds: secp.Message.RequiredFunds(),
			Size:          int64(len(data)),
			Height:        uint64(b.Block.Height),
			MsgCreate:     b.Block.Timestamp,
			GmtCreate:     now,
			GmtModified:   now,
			Signature:     secp.Signature}

		if secp.Message.Method == 0 {
			fs_msg.MethodName = "Transfer"
		} else {
			if actor, method, err := ParseActorMessage(&secp.Message); err == nil {
				fs_msg.ActorName = actor.Name
				fs_msg.MethodName = method.Name
			}
		}
		//fs_msg.MessageParams = new(InterfaceParam).DecoderMessageParam(fs_msg.Cid, fs_msg.MethodName, fs_msg.Height, false)
		fsMsgList = append(fsMsgList, fs_msg)
	}
	return fsMsgList
}

func (b *TipsetBlockMessages) equals(tps_blms *TipsetBlockMessages) bool {
	if b == tps_blms {
		return true
	}
	if b == nil || tps_blms == nil {
		return false
	}
	return b.Tipset.Equals(tps_blms.Tipset)
}

func (b *TipsetBlockMessages) receipts_ref() map[string]*types.MessageReceipt {
	receipt_ref := make(map[string]*types.MessageReceipt)
	for index, receipt := range b.Receipts {
		receipt_ref[b.Messages[index].Cid.String()] = receipt
	}
	return receipt_ref
}

func (b *TipsetBlockMessages) modelsUpsert() error {
	return (&TipsetBlockMessageList{[]*TipsetBlockMessages{b}}).buildModelsData().modelsUpsert()
}

func (b *TipsetBlockMessages) buildModelsData() (*module.FilscanTipSet, []*module.FilscanBlock, []*module.FilscanMsg, *TipsetMinerMessages) {
	//	utils.Log.Traceln("消息----------------整理")
	if b.isCached {
		return b.fsTipset, b.fsBlocks, b.fsMsgs, b.fsMiners
	}

	var modelTipset *module.FilscanTipSet
	var modelBlocks []*module.FilscanBlock
	var modelMsgs []*module.FilscanMsg

	var modelBlockMsgs []*module.BlockAndMsg

	var receiptRef = b.receipts_ref()
	var minerMessages = &TipsetMinerMessages{
		tipset: b.Tipset,
		miners: make(map[string]struct{})}

	modelTipset = toFsTipset(b.Tipset)
	for _, blMsg := range b.BlockMsgs {
		fsBlock := blMsg.fsBlock()
		fsMessages := blMsg.fsMessages()

		fsBlockMessage := &module.BlockAndMsg{
			Block: fsBlock}

		fsBlock.BlockReward = b.BlockRwds
		for _, fsmsg := range fsMessages {
			// 设置message相关的receipt
			if receipt, exist := receiptRef[fsmsg.Cid]; exist && receipt != nil {
				fsmsg.ExitCode = strconv.Itoa(int(receipt.ExitCode))
				fsmsg.GasUsed = strconv.FormatInt(receipt.GasUsed, 10)
				fsmsg.Return = string(receipt.Return)
			}
			modelMsgs = append(modelMsgs, fsmsg)

			fsBlockMessage.Msg = append(fsBlockMessage.Msg, fsmsg)
			if fsmsg.ActorName == "" {
				fsmsg.ActorName = "fil/1/account"
			}

			if fsmsg.ActorName == "fil/1/storageminer" {
				minerMessages.miners[fsmsg.Message.To.String()] = struct{}{}
			}
		}

		modelBlocks = append(modelBlocks, fsBlock)
		modelBlockMsgs = append(modelBlockMsgs, fsBlockMessage)
	}

	b.fsTipset = modelTipset
	b.fsBlocks = modelBlocks
	b.fsMsgs = modelMsgs
	b.fsMiners = minerMessages
	b.fsBlockMessage = modelBlockMsgs

	b.isCached = true

	return modelTipset, modelBlocks, modelMsgs, minerMessages
}

type fsModelsData struct {
	tipsets  []*module.FilscanTipSet
	blocks   []*module.FilscanBlock
	messages []*module.FilscanMsg
	miners   []*TipsetMinerMessages
}

func (b *fsModelsData) modelsUpsert() error {
	return modelsBulkUpsertBlockMessageTipset(b.messages, b.blocks, b.tipsets)
}

func (b *TipsetBlockMessageList) buildModelsData() *fsModelsData {
	modelsData := &fsModelsData{
		miners: make([]*TipsetMinerMessages, len(b.TipsetBlockMessages)),
	}

	for index, tpst_blms := range b.TipsetBlockMessages {

		tmpFsTipset, tmpFsBlocks, tmpFsMsgs, tmpMinerMsgs := tpst_blms.buildModelsData()

		modelsData.tipsets = append(modelsData.tipsets, tmpFsTipset)
		modelsData.blocks = append(modelsData.blocks, tmpFsBlocks[:]...)
		modelsData.messages = append(modelsData.messages, tmpFsMsgs[:]...)
		modelsData.miners[index] = tmpMinerMsgs
	}

	return modelsData
}
func modelsBulkUpsertTispet(col *mgo.Collection, tipsetList []*module.FilscanTipSet) error {
	size := len(tipsetList)
	if size == 0 {
		return nil
	}

	bulk_items := make([]interface{}, size*2)

	for index, tipset := range tipsetList {
		i := index * 2
		bulk_items[i] = bson.M{"height": tipset.Height}
		bulk_items[i+1] = utils.ToInterface(tipset)
	}

	_, err := module.BulkUpsert(col, "tipset", bulk_items)
	return err
}

func modelsBulkUpsertMessage(col *mgo.Collection, msg_list []*module.FilscanMsg) error {
	size := len(msg_list)
	if size == 0 {
		return nil
	}

	const maxSize = 256
	bulkItems := make([]interface{}, maxSize*2)

	for size > 0 {
		realSize := size
		if size > maxSize {
			realSize = maxSize
		}

		var index = 0
		for ; index < realSize; index++ {
			i := index * 2
			bulkItems[i] = bson.M{"cid": msg_list[index].Cid}
			bulkItems[i+1] = utils.ToInterface(msg_list[index])
		}

		if _, err := module.BulkUpsert(col, module.MsgCollection, bulkItems[:index*2]); err != nil {
			return err
		}

		msg_list = msg_list[index:]
		size = len(msg_list)
	}

	return nil
}

func modelsBulkUpsertBlock(col *mgo.Collection, blockList []*module.FilscanBlock) error {
	size := len(blockList)
	if size == 0 {
		return nil
	}

	bulkItems := make([]interface{}, size*2)
	for index, block := range blockList {
		i := index * 2
		bulkItems[i] = bson.M{"cid": block.Cid}
		bulkItems[i+1] = utils.ToInterface(block)
	}
	_, err := module.BulkUpsert(col, module.BlocksCollection, bulkItems)
	return err
}

func modelsBulkUpsertBlockMessageTipset(fs_messages []*module.FilscanMsg, fs_blocks []*module.FilscanBlock, fs_tipsets []*module.FilscanTipSet) error {
	ms, db := module.Copy()
	defer ms.Close()

	if err := modelsBulkUpsertMessage(db.C(module.MsgCollection), fs_messages); err != nil {
		return err
	}
	if err := modelsBulkUpsertBlock(db.C(module.BlocksCollection), fs_blocks); err != nil {
		return err
	}
	if err := modelsBulkUpsertTispet(db.C("tipset"), fs_tipsets); err != nil {
		return err
	}

	//TODO 写入统计表
	//统计历史
	// var bh module.BlockHistory
	// blockminers = make([]string, 0)
	// for _, v := range fs_blocks {
	// 	if v.BlockHeader.Miner.String() != "" && v.Cid != "" {
	// 		bh.Miner = v.BlockHeader.Miner.String()
	// 		bh.BlockCid = v.Cid
	// 		bh.BlockReword = v.BlockReward
	// 		bh.Created = time.Now().Format(utils.TimeString)
	// 		bh.Date = time.Now().Format(utils.TimeDate)
	// 		blockStatistics(&bh)
	// 	}
	// }
	//	utils.Log.Traceln("not found", blockminers)
	return nil
	//TODO 写入统计表
	//统计历史
}

var blockminers []string

func blockStatistics(param *module.BlockHistory) {
	//创建事务 看了下代码这个包好像不支持事务
	if param.Miner == "" {
		return
	}
	param.Date = time.Now().Format(utils.TimeDate)
	param.Created = time.Now().Format(utils.TimeString)
	if err := new(module.BlockHistory).InsertOne(param); err != nil {
		//	utils.Log.Errorln(err)
		return
	}
	if err := module.UpdateStatistics(param.Miner, param.BlockReword, 1); err != nil {
		//utils.Log.Errorln(err)
		blockminers = append(blockminers, param.Miner)
		//删除数据
		new(module.BlockHistory).Remove(param.BlockCid)
	}
}

func toFsTipset(tipset *types.TipSet) *module.FilscanTipSet {
	now := time.Now().Unix()
	fs_tipset := &module.FilscanTipSet{
		Key:          tipset.Key().String(),
		ParentKey:    tipset.Parents().String(),
		Cids:         tipset.Cids(),
		Height:       uint64(tipset.Height()),
		Mintime:      tipset.MinTimestamp(),
		Parents:      tipset.Parents().Cids(),
		GmtCreate:    now,
		GmtModified:  now,
		MinTicketCId: tipset.MinTicketBlock().Cid(),
	}
	return fs_tipset
}

//定时任务查库统计数据 把区块奖励和出块数量累计入统计表
func DbDatestatistics() {

}
