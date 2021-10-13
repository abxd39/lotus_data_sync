package syncer

import (
	"lotus_data_sync/module"
	"lotus_data_sync/utils"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/chain/types"
)

func (fs *Filscaner) handleStorageMinerMessage(tipset *types.TipSet, miner string) {
	address, err := address.NewFromString(miner)
	if err != nil {
		utils.Log.Errorf("handle miner(%s) message failed, message:%s", miner, err)
	}
	minerState, err := fs.apiMinerStateAtTipset(address, tipset)
	if err != nil {
		utils.Log.Errorf("api_get_miner_state(%s) at tipset(%d) message failed, message:%s",
			miner, tipset.Height(), err.Error())
		return
	}
	if minerState == nil {
		return
	}
	fs.modelsUpdateMiner(minerState)
	fs.minerStateChan <- minerState
}

func (fs *Filscaner) handleMinerState(miner *module.MinerStateAtTipset) {
	// fs.minerCache24h.update(miner)
	// fs.minerCache1day.update(miner)
}
