package syncer

import (
	"container/list"
	"lotus_data_sync/module"
	"lotus_data_sync/utils"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/globalsign/mgo/bson"
	"sync"
)

const collectionMinSyncedHeight = "min_synced_height"
const collectionChianSyncedTipset = "synced_tipset"

type models_min_synced_height struct {
	MinHeight uint64 `bson:"min_height"`
}

// 同步状态信息
type fsSyncedTipset struct {
	Key       string `bson:"key"`
	ParentKey string `bson:"parent_key"`
	Height    uint64 `bson:"height"`
}

func (fs *fsSyncedTipset) updateWithTipset(tipset *types.TipSet) *fsSyncedTipset {
	if tipset == nil {
		return fs
	}
	fs.Key = tipset.Key().String()
	fs.ParentKey = tipset.Parents().String()
	fs.Height = uint64(tipset.Height())
	return fs
}

func (fs *fsSyncedTipset) updateWithParent(tipset *types.TipSet) bool {
	if fs.is_my_parent(tipset) {
		fs.updateWithTipset(tipset)
		return true
	}
	return false
}

func (fs *fsSyncedTipset) updateWithChild(tipset *types.TipSet) bool {
	if fs.is_my_child_(tipset) {
		fs.updateWithTipset(tipset)
		return true
	}
	return false
}

func (fs *fsSyncedTipset) is_my_child_(tipset *types.TipSet) bool {
	if fs == nil || tipset == nil {
		return false
	}

	utils.SugarLogger.Infof("info:is_my_child, fs.key:%s, tipset.parents.key:%s", fs.Key, tipset.Parents().String())

	return fs.Key == tipset.Parents().String()
}

func (fs *fsSyncedTipset) is_my_parent(tipset *types.TipSet) bool {
	if fs == nil || tipset == nil {
		return false
	}
	return fs.ParentKey == tipset.Key().String()
}

func (fs *fsSyncedTipset) is_my_child__(tipset *fsSyncedTipset) bool {
	if fs == nil || tipset == nil {
		return false
	}
	return fs.Key == tipset.ParentKey && fs.Height < tipset.Height
}

type fsSyncedTipsetPath struct {
	Head         *fsSyncedTipset `bson:"head"`
	Tail         *fsSyncedTipset `bson:"tail"`
	upsertSelect bson.M          `bson:"-"`
}

var fsNewStp = fsNewSyncedTipsetPath

func fsNewSyncedTipsetPath(head, tail *fsSyncedTipset) *fsSyncedTipsetPath {
	path := &fsSyncedTipsetPath{head, tail, nil}
	return path.refreshSelc()
}

func (fs *fsSyncedTipsetPath) tryMergeWith(fstp *fsSyncedTipsetPath) bool {
	var merged = false
	if fs.Head.Height > fstp.Head.Height {
		if fs.Tail.ParentKey == fstp.Head.Key || fs.Tail.Height <= fstp.Head.Height {
			fs.Tail = fstp.Tail
			merged = true
		}
	} else if fstp.Head.Height > fs.Head.Height {
		if fstp.Tail.ParentKey == fs.Head.Key || fstp.Tail.Height <= fs.Head.Height {
			fs.Head = fstp.Head
			merged = true
		}
	} else if fs.Head.Height == fstp.Head.Height {
		if fstp.Tail.Height < fs.Tail.Height {
			fs.Tail = fstp.Tail
		}
		merged = true
	}
	return merged
}

func (fs *fsSyncedTipsetPath) modelsDel() error {
	ms, c := module.Connect(collectionChianSyncedTipset)
	defer ms.Close()
	return c.Remove(fs.upsertSelect)
}

func (fs *fsSyncedTipsetPath) refreshSelc() *fsSyncedTipsetPath {
	fs.upsertSelect = bson.M{
		"head.height": fs.Head.Height,
		"tail.height": fs.Tail.Height}
	return fs
}

func (fs *fsSyncedTipsetPath) modelsUpsert() error {
	ms, c := module.Connect(collectionChianSyncedTipset)
	defer ms.Close()
	if _, err := c.Upsert(fs.upsertSelect, fs); err != nil {
		return err
	}
	fs.refreshSelc()
	return nil
}

type fsSyncedTipsetPathList struct {
	PathList *list.List
	mutex    sync.Mutex
}

func modelsNewSyncedTipsetList() (*fsSyncedTipsetPathList, error) {
	synced := &fsSyncedTipsetPathList{}
	err := synced.modelsLoad()
	return synced, err
}

func (fs *fsSyncedTipsetPathList) lock() {
	fs.mutex.Lock()
}

func (fs *fsSyncedTipsetPathList) unlock() {
	fs.mutex.Unlock()
}

func (fs *fsSyncedTipsetPathList) pushNewPath(tipset *types.TipSet) error {
	fs.lock()
	defer fs.unlock()

	stp := fsNewStp((&fsSyncedTipset{}).updateWithTipset(tipset),
		(&fsSyncedTipset{}).updateWithTipset(tipset))

	fs.PathList.PushFront(stp)

	return stp.modelsUpsert()
}

func (fs *fsSyncedTipsetPathList) tryMergeFrontWithNext(lock bool) *fsSyncedTipsetPath {
	if lock {
		fs.lock()
		defer fs.unlock()
	}

	if eFront := fs.PathList.Front(); eFront != nil {
		if eNext := eFront.Next(); eNext != nil {
			front := eFront.Value.(*fsSyncedTipsetPath)
			next := eNext.Value.(*fsSyncedTipsetPath)
			if front.tryMergeWith(next) {
				fs.PathList.Remove(eNext)
				return next
			}

		}
	}
	return nil
}

func (fs *fsSyncedTipsetPathList) modelsUpsertFront(lock bool) error {
	if lock {
		fs.lock()
		defer fs.unlock()
	}
	return fs.PathList.Front().Value.(*fsSyncedTipsetPath).modelsUpsert()
}

func (fs *fsSyncedTipsetPathList) frontSyncedPath() bson.M {
	fs.lock()
	defer fs.unlock()
	if fs.PathList.Len() != 0 {
		return fs.PathList.Front().Value.(*fsSyncedTipsetPath).upsertSelect
	}
	return nil
}

func (fs *fsSyncedTipsetPathList) frontTail() *fsSyncedTipset {
	fs.lock()
	defer fs.unlock()
	if fs.PathList.Len() > 0 {
		return fs.PathList.Front().Value.(*fsSyncedTipsetPath).Tail
	}
	return nil
}

func (fs *fsSyncedTipsetPathList) modelsLoad() error {
	fs.lock()
	defer fs.unlock()

	ms, c := module.Connect(collectionChianSyncedTipset)
	defer ms.Close()

	var pathArr []*fsSyncedTipsetPath
	utils.Log.Traceln("modelsLoad")
	if err := c.Find(nil).Sort("-head.tipset_height").Limit(20).All(&pathArr); err != nil {
		return err
	}

	fs.PathList = list.New()

	for _, path := range pathArr {
		// 加载的时候, 刷新一下更新要使用的selc
		path.refreshSelc()
		fs.PathList.PushFront(path)
	}

	changed := false
	for merged := fs.tryMergeFrontWithNext(false); merged != nil; {
		changed = true
		if err := merged.modelsDel(); err != nil {
			return err
		}
	}
	if changed {
		return fs.modelsUpsertFront(false)
	}
	return nil
}

func (fs *fsSyncedTipsetPathList) insertTailParent(tipset *types.TipSet) bool {
	fs.lock()
	defer fs.unlock()

	eF := fs.PathList.Front()

	if eF == nil {
		eF = fs.PathList.PushFront(fsNewSyncedTipsetPath(
			(&fsSyncedTipset{}).updateWithTipset(tipset),
			(&fsSyncedTipset{}).updateWithTipset(tipset)))
		return true
	}
	return eF.Value.(*fsSyncedTipsetPath).Tail.updateWithParent(tipset)
}

func (fs *fsSyncedTipsetPathList) insertHeadChild(tipset *types.TipSet) bool {
	fs.lock()
	defer fs.unlock()
	eF := fs.PathList.Front()

	if eF == nil {
		eF = fs.PathList.PushFront(fsNewSyncedTipsetPath(
			(&fsSyncedTipset{}).updateWithTipset(tipset),
			(&fsSyncedTipset{}).updateWithTipset(tipset)))
		return true
	}
	return eF.Value.(*fsSyncedTipsetPath).Head.updateWithChild(tipset)
}
