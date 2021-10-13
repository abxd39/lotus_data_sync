package syncer

import (
	"container/list"
	"lotus_data_sync/module"
	"lotus_data_sync/utils"
	"github.com/filecoin-project/lotus/chain/types"
	"sync"
)

// use *TipsetBlockMessages as list.element.Value
type FsTipsetCache struct {
	list    *list.List
	maxSize int
	mutex   sync.Mutex
}

func newFsCache(maxSize int) *FsTipsetCache {
	fsc := &FsTipsetCache{
		list:    list.New(),
		maxSize: maxSize}
	return fsc
}

func (fsc *FsTipsetCache) lock() {
	fsc.mutex.Lock()
}

func (fsc *FsTipsetCache) unlock() {
	fsc.mutex.Unlock()
}

func (fsc *FsTipsetCache) full() bool {
	fsc.lock()
	defer fsc.unlock()
	return fsc.list.Len() >= fsc.maxSize
}

func (fsc *FsTipsetCache) Size() int {
	fsc.lock()
	defer fsc.unlock()
	return fsc.list.Len()
}

func (fsc *FsTipsetCache) push_back(in *TipsetBlockMessages) {
	if in == nil {
		return
	}
	fsc.lock()
	defer fsc.unlock()

	if fsc.list.Len() > 0 {
		if in.equals(fsc.list.Back().Value.(*TipsetBlockMessages)) {
			return
		}
	}

	fsc.list.PushBack(in)
}

func (fsc *FsTipsetCache) pushFront(in *TipsetBlockMessages) (blockMessage *TipsetBlockMessages) {
	if in == nil || in.Tipset == nil {
		return
	}
	fsc.lock()
	defer fsc.unlock()

	if fsc.list.Len() > 0 {
		if fsc.list.Front().Value.(*TipsetBlockMessages).equals(in) {

			return
		}
	}

	in_tipset := in.Tipset
	if true {
		for front := fsc.list.Front(); front != nil; front = fsc.list.Front() {
			if f := front.Value.(*TipsetBlockMessages).Tipset; f != nil {
				if f.Key().String() == in_tipset.Parents().String() {
					break
				}
				utils.Log.Infof("this is a forked situation, income in:%d, removed in:%d",
					in_tipset.Height(), f.Height())
				fsc.list.Remove(front)
			}
		}
	} else {
		// check if in.Height() < fsc.list.front.height(),
		//   and sovle this chain 'forked' situation
		for fsc.list.Len() > 0 && in_tipset.Height() <= fsc.list.Front().Value.(*types.TipSet).Height() {
			fsc.list.Remove(fsc.list.Front())
		}
	}
	fsc.list.PushFront(in)
	
	if fsc.list.Len() > fsc.maxSize {
		blockMessage = fsc.list.Remove(fsc.list.Back()).(*TipsetBlockMessages)
		utils.Log.Traceln("fsc.list.Remove height=",blockMessage.fsTipset.Height)
		// child = fsc.list.Back().Value.(*Tipset_block_messages)
	}
	return
}

func (fsc *FsTipsetCache) Front() *types.TipSet {
	fsc.lock()
	defer fsc.unlock()
	if fsc.list.Len() > 0 {
		return fsc.list.Front().Value.(*TipsetBlockMessages).Tipset
	}
	return nil
}

type IsMatchFunc func(*TipsetBlockMessages) bool

func (fsc *FsTipsetCache) Loop(isMatch IsMatchFunc) interface{} {
	fsc.lock()
	defer fsc.unlock()

	for f := fsc.list.Front(); f != nil; f = f.Next() {
		if isMatch(f.Value.(*TipsetBlockMessages)) {
			// TODO:!!!!!
		}
	}

	return nil
}

func (fsc *FsTipsetCache) FindBlockOffsetCount(offset, count int) []*module.BlockAndMsg {
	fsc.lock()
	defer fsc.unlock()

	listSize := fsc.list.Len()
	if offset+count > listSize {
		count = listSize - offset
	}

	var blockmsgArr []*module.BlockAndMsg

	front := fsc.list.Front()
	blockIndex := 0
exit:
	for front != nil {
		tipsetBlockMessage := front.Value.(*TipsetBlockMessages)
		front = front.Next()

		for _, blockMsg := range tipsetBlockMessage.fsBlockMessage {
			if blockIndex < offset {
				blockIndex++
				continue
			}

			blockmsgArr = append(blockmsgArr, blockMsg)

			if blockIndex-offset >= count {
				break exit
			}
			blockIndex++
		}
	}
	return blockmsgArr
}

func (fsc *FsTipsetCache) FindMessageOffsetCount(offset, count int) []*module.FilscanMsg {
	fsc.lock()
	defer fsc.unlock()

	listSize := fsc.list.Len()
	if offset+count > listSize {
		count = listSize - offset
	}

	var msgArr = []*module.FilscanMsg{}

	front := fsc.list.Front()
	blockIndex := 0
exit:
	for front != nil {
		tipsetBlockMessage := front.Value.(*TipsetBlockMessages)
		front = front.Next()
		for _, msg := range tipsetBlockMessage.fsMsgs {
			if blockIndex < offset {
				blockIndex++
				continue
			}
			msgArr = append(msgArr, msg)

			if blockIndex-offset >= count {
				break exit
			}
			blockIndex++
		}
	}
	return msgArr
}

func (fsc *FsTipsetCache) FindtipsetHeight(height uint64) *module.Element {
	fsc.lock()
	defer fsc.unlock()

	for back := fsc.list.Back(); back != nil; back = back.Prev() {
		blockMessage := back.Value.(*TipsetBlockMessages)
		if uint64(blockMessage.Tipset.Height()) == height {
			return &module.Element{Tipset: blockMessage.Tipset, Blocks: blockMessage.fsBlockMessage}
		}
	}
	return nil
}

func (fsc *FsTipsetCache) FindtipsetInHeight(start, end uint64) []*module.Element {
	fsc.lock()
	defer fsc.unlock()
	var eleArr []*module.Element

	for back := fsc.list.Back(); back != nil; back = back.Prev() {
		blockMessage := back.Value.(*TipsetBlockMessages)

		if uint64(blockMessage.Tipset.Height()) > end {
			break
		}

		if uint64(blockMessage.Tipset.Height()) < start {
			continue
		}

		eleArr = append(eleArr, &module.Element{Tipset: blockMessage.Tipset, Blocks: blockMessage.fsBlockMessage})
	}

	return eleArr
}

func (fsc *FsTipsetCache) FindMesage_block(block_id string) []*module.FilscanMsg {
	fsc.lock()
	defer fsc.unlock()

	for front := fsc.list.Front(); front != nil; front = front.Next() {
		tipset_blockmessage := front.Value.(*TipsetBlockMessages)
		for _, blockmessage := range tipset_blockmessage.fsBlockMessage {
			if blockmessage.Block.Cid == block_id {
				return blockmessage.Msg
			}
		}
	}

	return []*module.FilscanMsg{}
}

func (fsc *FsTipsetCache) FindMesage_method(method string) []*module.FilscanMsg {
	var fsMsgArr []*module.FilscanMsg
	if method == "" {
		return fsMsgArr
	}

	fsc.lock()
	defer fsc.unlock()

	for front := fsc.list.Front(); front != nil; front = front.Next() {
		tipset_blockmessage := front.Value.(*TipsetBlockMessages)

		for _, msg := range tipset_blockmessage.fsMsgs {
			if msg.MethodName == method {
				fsMsgArr = append(fsMsgArr, msg)
			}
		}
	}

	return fsMsgArr
}

func (fsc *FsTipsetCache) FindMesage_id(cid string) *module.FilscanMsg {
	fsc.lock()
	defer fsc.unlock()

	for front := fsc.list.Front(); front != nil; front = front.Next() {
		tipset_blockmessage := front.Value.(*TipsetBlockMessages)

		for _, msg := range tipset_blockmessage.fsMsgs {
			if msg.Cid == cid {
				return msg
			}
		}
	}
	return nil
}

func (fsc *FsTipsetCache) FindMesage_blocks(blocks []string) []*module.FilscanMsg {
	mblocks := make(map[string]struct{})
	for _, id := range blocks {
		mblocks[id] = struct{}{}
	}

	fsc.lock()
	defer fsc.unlock()

	var msgArr []*module.FilscanMsg
	for front := fsc.list.Front(); front != nil; front = front.Next() {
		tipset_blockmessage := front.Value.(*TipsetBlockMessages)
		for _, block := range tipset_blockmessage.fsBlockMessage {
			if _, exist := mblocks[block.Block.Cid]; exist {
				msgArr = append(msgArr, block.Msg[:]...)
			}
		}
	}
	return msgArr
}

func (fsc *FsTipsetCache) MessageAll() []*module.FilscanMsg {
	fsc.lock()
	defer fsc.unlock()

	var fsMsgArr []*module.FilscanMsg
	for front := fsc.list.Front(); front != nil; front = front.Next() {
		tipset_blockmessage := front.Value.(*TipsetBlockMessages)
		fsMsgArr = append(fsMsgArr, tipset_blockmessage.fsMsgs[:]...)
	}
	return fsMsgArr
}

func (fsc *FsTipsetCache) FindMesage_block_method(block_cid, method string) []*module.FilscanMsg {
	if block_cid != "" {
		blockmsg := fsc.FindBlock_id(block_cid)
		if blockmsg == nil {
			return nil
		}
		if method == "" {
			return blockmsg.Msg
		} else {
			var flmsgs []*module.FilscanMsg
			for _, msg := range blockmsg.Msg {
				if msg.MethodName == method {
					flmsgs = append(flmsgs, msg)
				}
			}
			return flmsgs
		}
	} else if method != "" {
		return fsc.FindMesage_method(method)
	} else {
		return fsc.MessageAll()
	}
}

func (fsc *FsTipsetCache) FindBlock_id(id string) *module.BlockAndMsg {
	fsc.lock()
	defer fsc.unlock()

	for front := fsc.list.Front(); front != nil; front = front.Next() {
		block_msg_arr := front.Value.(*TipsetBlockMessages).fsBlockMessage
		for _, block_msg := range block_msg_arr {
			if block_msg.Block.Cid == id {
				return block_msg
			}
		}
	}
	return nil
}

func (fsc *FsTipsetCache) FindBlock_miners(miners []string) []*module.FilscanBlock {
	fsc.lock()
	defer fsc.unlock()

	var blocks []*module.FilscanBlock

	mminers := utils.SlcToMap(miners, "", false).(map[string]struct{})

	for front := fsc.list.Front(); front != nil; front = front.Next() {
		blockMsgArr := front.Value.(*TipsetBlockMessages).fsBlockMessage
		for _, blockMsg := range blockMsgArr {
			if _, exist := mminers[blockMsg.Block.BlockHeader.Miner.String()]; exist {
				blocks = append(blocks, blockMsg.Block)
			}
		}
	}

	return blocks
}

func (fsc *FsTipsetCache) FindMesage_address(address, fromto, method string, beginTime, endTime int64) []*module.FilscanMsg {
	fsc.lock()
	defer fsc.unlock()

	var msgArr []*module.FilscanMsg
	for front := fsc.list.Front(); front != nil; front = front.Next() {
		tipsetBlockMessage := front.Value.(*TipsetBlockMessages)
		for _, msg := range tipsetBlockMessage.fsMsgs {
			if fromto == "from" && msg.Message.From.String() != address {
				break
			} else if fromto == "to" && msg.Message.To.String() != address {
				break
			} else if msg.Message.From.String() != address && msg.Message.To.String() != address {
				break
			}
			if method != "" && msg.MethodName != method {
				break
			}
			if beginTime != 0 && msg.GmtCreate < beginTime {
				break
			}
			if endTime != 0 && msg.GmtCreate > endTime {
				break
			}
			msg.Message.Params = make([]byte, 0) //参数不反回
			msgArr = append(msgArr, msg)
		}
	}
	return msgArr
}

func (fsc *FsTipsetCache) Blocks() []*module.BlockAndMsg {
	fsc.lock()
	defer fsc.unlock()

	var block_messages []*module.BlockAndMsg
	for front := fsc.list.Front(); front != nil; front = front.Next() {
		block_messages = append(block_messages, front.Value.(*TipsetBlockMessages).fsBlockMessage[:]...)
	}
	return block_messages
}

func (fsc *FsTipsetCache) TipsetCountInTime(start, end int64) int64 {
	fsc.lock()
	defer fsc.unlock()

	count := int64(0)
	for front := fsc.list.Front(); front != nil; front = front.Next() {
		at := int64(front.Value.(*TipsetBlockMessages).Tipset.MinTimestamp())
		if at >= start && at < end {
			count++
		}
	}
	return count
}

func (fsc *FsTipsetCache) LatestBlockRewards() string {
	fsc.lock()
	defer fsc.unlock()

	if front := fsc.list.Front(); front != nil {
		return front.Value.(*TipsetBlockMessages).BlockRwds
	}
	return ""
}
