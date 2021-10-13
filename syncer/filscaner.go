package syncer

import (
	"context"
	"errors"
	
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/v1api"
	"github.com/filecoin-project/lotus/chain/types"
	//"github.com/filecoin-project/node/config"
	"github.com/globalsign/mgo"
	"lotus_data_sync/module"
	"lotus_data_sync/utils"
	"lotus_data_sync/force/factors"
	"math/big"
	"sync"
)

type TipsetMinerMessages struct {
	miners map[string]struct{}
	tipset *types.TipSet
}

type MethodCall struct {
	actorName string
	*types.Message
	*factors.MethodInfo
}
type Filscaner struct {
	api    v1api.FullNode
	ctx    context.Context
	cancel context.CancelFunc
	//conf                       config.Configer
	headNotifier               chan *api.HeadChange
	tipsetMinerMessagesNotifer chan []*TipsetMinerMessages

	// 已经同步到的tipset高度,当程序重启时,
	// 需要从此高度同步到first_notifiedTipsetHeight
	tipsetCacheSize        uint64
	toSyncHeaderHeight     uint64
	safeHeight             uint64
	headerHeight           uint64
	mutexForNumbers        sync.Mutex
	chainGenesisTime       uint64
	waitGroup              sync.WaitGroup
	collation              *mgo.Collation
	toUpsertMiners         []interface{}
	toUpdateMinerSize      uint64
	toUpdateMinerIndex     uint64
	latestTotalPower       *big.Int
	displayTracks          bool
	syncedTipsetPathList   *fsSyncedTipsetPathList // tipset synced status loaded from database
	tipsetsCache           *FsTipsetCache          // un-confrimed tipses in front of chain head
	safeTipsetChannel      chan *types.TipSet      //
	lastSafeTipset         *types.TipSet
	lastApplyTippet        *types.TipSet
	isSyncToGenesisRunning bool
	handleApplyTippet      func(child, parent *types.TipSet)
	handleSafeTipset       func(blockMessage *TipsetBlockMessages)
	//minerCache24h          *fsMinerCache
	//minerCache1day         *fsMinerCache
	minerStateChan chan *module.MinerStateAtTipset
}

//var Inst = &Filscaner{ }
var Inst = &Filscaner{}

func NewInstance(ctx context.Context, config_path string, lotusApi v1api.FullNode) (*Filscaner, error) {
	filscaner := &Filscaner{}
	if err := filscaner.Init(ctx, config_path, lotusApi); err != nil {
		return nil, err
	}
	return filscaner, nil
}

func (fs *Filscaner) initConfiguration() error {
	var err error
	var cacheSize int64
	// if fs.conf, err = config.NewConfig("ini", filepath); err != nil {
	// 	return err
	// }
	if cacheSize, err = utils.Initconf.Int64("tipset_cache_size"); err != nil || cacheSize < 0 {
		return err
	}
	fs.tipsetCacheSize = uint64(cacheSize)
	return nil
}

func (fs *Filscaner) List() *FsTipsetCache {
	return fs.tipsetsCache
}

func (fs *Filscaner) initLotusClient(lotusApi v1api.FullNode) error {
	if lotusApi == nil {
		return errors.New("invalid parameters")
	}
	fs.api = lotusApi
	if err := fs.iniChainGenesisTime(); err != nil {
		return err
	}

	tipset, err := fs.api.ChainHead(context.TODO())
	if err != nil {
		return err
	}
	fs.refresh_height_state(uint64(tipset.Height()))
	return nil
}

func (fs *Filscaner) Init(ctx context.Context, config_path string, lotusApi v1api.FullNode) error {
	fs.ctx, fs.cancel = context.WithCancel(ctx)

	var err error
	if err = fs.initConfiguration(); err != nil {
		utils.Log.Errorln(err)
		return err
	}

	if err = fs.initLotusClient(lotusApi); err != nil {
		utils.Log.Errorln(err)
		return err
	}

	fs.headNotifier = make(chan *api.HeadChange)
	fs.tipsetMinerMessagesNotifer = make(chan []*TipsetMinerMessages, 1000)

	fs.toUpdateMinerSize = 512
	fs.toUpdateMinerIndex = 0
	fs.toUpsertMiners = make([]interface{}, fs.toUpdateMinerSize*2)

	fs.displayTracks = true

	fs.collation = &mgo.Collation{Locale: "zh", NumericOrdering: true}

	// if fs.syncedTipsetPathList, err = modelsNewSyncedTipsetList(); err != nil {
	// 	utils.Log.Errorln(err)
	// 	return err
	// }

	// fs.tipsetsCache = newFsCache(int(fs.tipsetCacheSize))
	// fs.safeTipsetChannel = make(chan *types.TipSet, 100)

	// fs.handleSafeTipset = fs.handleFirstSafeTipset
	// fs.handleApplyTippet = fs.handleFirstApplyTippet
	// utils.Log.Traceln("init begin ")
	// if err := fs.initMinersCaches(); err != nil {
	// 	utils.Log.Errorln(err)
	// 	return err
	// }

	return nil
}

func (fs *Filscaner) Run() {
	fs.TaskStartHandleMinerState()
	fs.Task_StartHandleMessage()
	fs.TaskStartSyncer()
	fs.TaskSyncTipsetRewardsDb()
	fs.TaskInitBlockRewards()
}

func (fs *Filscaner) iniChainGenesisTime() error {
	if fs.chainGenesisTime != 0 {
		return nil
	}

	genesis, err := fs.api.ChainGetGenesis(fs.ctx)
	if err != nil {
		return err
	}

	fs.chainGenesisTime = genesis.MinTimestamp()
	return nil
}
