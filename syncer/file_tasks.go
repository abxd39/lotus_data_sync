package syncer

import (

	"lotus_data_sync/module"
	"lotus_data_sync/utils"
	"github.com/filecoin-project/lotus/chain/types"
	"time"
	"errors"
)

// idea of syncing:
// head -> parent -> parent...-> genesis
// tipset cache for un-confrimed tipseds, there is a comfirm number
// save synced state to database
func (fs *Filscaner) runSyncer() error {
reSync:
	notifs, err := fs.api.ChainNotify(fs.ctx)
	if err != nil {
		utils.Log.Traceln(err)
		time.Sleep(time.Second * 2)
		goto reSync
	}
	ping := time.NewTicker(time.Second * 7)
	for {
		select {
		case headers, ok := <-notifs:
			{
				if !ok {
					goto reSync
				}
				Tempheaders := headers
				go fs.handleNewHeaders(Tempheaders)
			}
		case <-ping.C:
			{
				// health check
				//utils.Log.Traceln("second*5")
				if _, err := fs.api.ID(fs.ctx); err != nil {
					utils.Log.Errorf("error, lotus api 'ping' failed, message:%s", err.Error())
					return err
				}
			}
		case <-fs.ctx.Done():
			{
				ping.Stop()
				//utils.Log.Errorln("run_syncer stopped by ctx.done()")
				return fs.ctx.Err()
			}
		}
	}
	 
}

func (fs *Filscaner) TaskStartSyncer() {
	fs.waitGroup.Add(1)
	go func() {
	reRunSyncer:
		if err := fs.runSyncer(); err != nil {
			if err == errors.New("notifier was closed") {
				utils.Log.Errorf("who closed fs.api.ChainNotify  reRunSyncer")
				goto reRunSyncer
			}
			utils.Log.Errorf("run_syncer error, message:%s", err.Error())
		}

		fs.waitGroup.Done()
	}()
}

func (fs *Filscaner) taskSyncToGenesis(tipset *types.TipSet) {
	fs.syncedTipsetPathList.pushNewPath(tipset)
	fs.waitGroup.Add(1)
	go func() {
		latest, err := fs.syncToGenesis(tipset)
		if latest != nil {
			utils.Log.Errorf("‹‹‹synced_to_tipst finished: (%d, %s)›››", latest.Height(), latest.Key().String())
		}
		if err != nil {
			utils.Log.Errorf("sync to genesis failed,message:%s,height:%v,cid:%v", err.Error(), tipset.Height(), tipset.Cids())
		}
		fs.waitGroup.Done()
	}()
}

func (fs *Filscaner) TaskStartHandleMinerState() {
	fs.waitGroup.Add(1)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				utils.Log.Errorf("%v", err)
			}
		}()
		fs.loopHandleRefreshMinerState()
		fs.waitGroup.Done()
	}()
}

func (fs *Filscaner) Task_StartHandleMessage() {
	utils.Log.Tracef("debug toUpdateMinerSize= %v toUpdateMinerIndex=%v", fs.toUpdateMinerSize, fs.toUpdateMinerIndex)
	fs.waitGroup.Add(1)
	go func() {
		fs.loopHandleMessages()
		fs.waitGroup.Done()
	}()
}

func (fs *Filscaner) refresh_height_state(header_height uint64) {
	fs.mutexForNumbers.Lock()
	defer fs.mutexForNumbers.Unlock()

	fs.headerHeight = header_height

	if fs.headerHeight > fs.tipsetCacheSize {
		fs.safeHeight = fs.headerHeight - fs.tipsetCacheSize
		fs.toSyncHeaderHeight = fs.safeHeight
	}
}

func (fs *Filscaner) defaultHandleMessage(hctype string, method *MethodCall) {}

func (fs *Filscaner) TaskInitBlockRewards() {
	fs.waitGroup.Add(1)
	go func() {
		fs.loopInitBlockRewards()
		fs.waitGroup.Done()
	}()
}

func (fs *Filscaner) TaskSyncTipsetRewardsDb() {
	fs.waitGroup.Add(1)
	go func() {
		module.LoopWalkthroughtipsetrewards(fs.ctx)
		fs.waitGroup.Done()
	}()
}
