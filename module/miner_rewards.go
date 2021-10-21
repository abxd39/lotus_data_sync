package module

import (
	"context"
	"lotus_data_sync/utils"
	"fmt"
	"github.com/filecoin-project/lotus/build"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"math/big"
	"strconv"
	"time"
)

const TipsetRewardsCollection = "tipset_rewards"

type minersBlocksRewards struct {
	Miner           string      `bson:"miner"`
	MinedBlockCount uint64      `bson:"mined_block_count"`
	Rewards         *BsonBigint `bson:"rewards"`
}

type TipsetBlockRewards struct {
	TipsetHeight         uint64                          `bson:"tipset_height"`
	TotalBlockCount      uint64                          `bson:"total_block_count"`
	TipsetBlockCount     uint64                          `bson:"tipset_block_count"`
	TimeStamp            uint64                          `bson:"time_stamp"`
	TipsetReward         *BsonBigint                     `bson:"current_tipset_rewards"`
	TotalReleasedRewards *BsonBigint                     `bson:"chain_released_rewards"`
	Miners               map[string]*minersBlocksRewards `bson:"miners"`
}

func CreateMinerIndex_old() {
	ms, c := Connect(MinerCollection)
	defer ms.Close()

	ms.SetMode(mgo.Monotonic, true)

	indexs := []mgo.Index{
		{Key: []string{"-mine_time"}, Unique: false, Background: true},
		{Key: []string{"miner_addr"}, Unique: false, Background: true},
		{Key: []string{"peer_id"}, Unique: false, Background: true},
	}
	for _, index := range indexs {
		if err := c.EnsureIndex(index); err != nil {
			panic(err)
		}
	}
}

func CreateTipsetRewardsIndex() {
	ms, c := Connect(TipsetRewardsCollection)
	defer ms.Close()
	ms.SetMode(mgo.Monotonic, true)

	indexs := []mgo.Index{
		{Key: []string{"time_stamp"}, Unique: true, Background: true},
		{Key: []string{"miners"}, Unique: false, Background: true},
	}
	for _, index := range indexs {
		if err := c.EnsureIndex(index); err != nil {
			panic(err)
		}
	}
}

func (tbr *TipsetBlockRewards) AddMinedBlock(reward *big.Int, miner_addr string) {
	tbr.TotalBlockCount++
	tbr.TipsetBlockCount++
	if reward == nil {
		reward = big.NewInt(int64(25))
	}
	tbr.TipsetReward.Add(tbr.TipsetReward.Int, reward)
	tbr.TotalReleasedRewards.Add(tbr.TotalReleasedRewards.Int, reward)

	if tbr.Miners == nil {
		tbr.Miners = make(map[string]*minersBlocksRewards)
	}
	miner, exist := tbr.Miners[miner_addr]
	if !exist || miner == nil {
		miner = &minersBlocksRewards{
			Rewards: &BsonBigint{Int: big.NewInt(0)},
			Miner:   miner_addr,
		}
		tbr.Miners[miner_addr] = miner
	}
	miner.Rewards.Add(miner.Rewards.Int, reward)
	miner.MinedBlockCount++
}

func lastBlockChainRewards() (*TipsetBlockRewards, error) {
	ms, c := Connect(TipsetRewardsCollection)
	defer ms.Close()

	last_tipset_rewards := &TipsetBlockRewards{}
	err := c.Find(nil).Sort("-tipset_height").Limit(1).One(last_tipset_rewards)
	return last_tipset_rewards, err
}

func blocksAtHeight(offset, count uint64) ([]FilscanBlockResult, error) {
	ms, c := connect(BlocksCollection)
	defer ms.Close()

	var res []FilscanBlockResult

	q_find := bson.M{"block_header.Height": bson.M{"$gte": offset, "$lt": offset + count}}
	q_sort := "block_header.Height"
	//	utils.Log.Traceln(q_find)
	err := c.Find(q_find).Sort(q_sort).All(&res)
	return res, err
}

func LoopWalkthroughtipsetrewards(ctx context.Context) error {
	var blocks []FilscanBlockResult
	var err error
	var lastTipsetRewards *TipsetBlockRewards

	lastTipsetRewards, err = lastBlockChainRewards()
	if err != nil {
		if err == mgo.ErrNotFound {
			lastTipsetRewards = &TipsetBlockRewards{
				Miners:               map[string]*minersBlocksRewards{},
				TipsetHeight:         0,
				TotalBlockCount:      0,
				TipsetReward:         &BsonBigint{Int: big.NewInt(0)},
				TotalReleasedRewards: &BsonBigint{Int: big.NewInt(0)}}
		} else {
			return err
		}
	}

	remainingFilcoin := types.FromFil(build.FilBase)
	remainingFilcoin.Sub(remainingFilcoin.Int, lastTipsetRewards.TotalReleasedRewards.Int)

	for {
		blocks, err = blocksAtHeight(lastTipsetRewards.TipsetHeight, 201)
		if err != nil {
			utils.Log.Errorln(err)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Second * time.Duration(60)):
			if err != nil || len(blocks) == 0 {
				continue
			}
		}

		var blockReward types.BigInt
		tipsetRewardsMap := make(map[uint64]*TipsetBlockRewards)

		// todo : 没有multi这里需要检查一下..
		for _, block := range blocks {
			height := block.BlockHeader.Height
			minerAddr := block.BlockHeader.Miner
			res, _ := strconv.ParseInt(block.BlockReword, 10, 64)
			blockReward = types.BigInt{big.NewInt(int64(res))}

			tipsetRewards, exist := tipsetRewardsMap[height]

			// 由于blocks 是按正序排序的
			// 所以, 只要!exist, 则表明, 块高发生了变化, 需要重新计算爆块奖励
			if !exist || tipsetRewards == nil {
				previouseTipsetRewards, exist := tipsetRewardsMap[height-1]
				if !exist {
					previouseTipsetRewards = lastTipsetRewards
				}

				tipsetRewards = &TipsetBlockRewards{
					TipsetHeight:         height,
					TotalBlockCount:      previouseTipsetRewards.TotalBlockCount,
					TipsetBlockCount:     0,
					Miners:               make(map[string]*minersBlocksRewards),
					TimeStamp:            block.BlockHeader.Timestamp,
					TipsetReward:         &BsonBigint{Int: big.NewInt(0)},
					TotalReleasedRewards: &BsonBigint{Int: big.NewInt(0).Set(previouseTipsetRewards.TotalReleasedRewards.Int)},
				}

				tipsetRewardsMap[height] = tipsetRewards
				lastTipsetRewards = tipsetRewards
			}

			// 取最大block.timestamp作为tipset的timestamp
			if tipsetRewards.TimeStamp < block.BlockHeader.Timestamp {
				tipsetRewards.TimeStamp = block.BlockHeader.Timestamp
			}

			tipsetRewards.AddMinedBlock(blockReward.Int, minerAddr)
			remainingFilcoin.Sub(remainingFilcoin.Int, blockReward.Int)
		}

		bulkUpsertTipsetRewards(tipsetRewardsMap)
	}
}

func bulkUpsertTipsetRewards(tipsetRewards map[uint64]*TipsetBlockRewards) error {
	size := len(tipsetRewards)
	bulkElements := make([]interface{}, size*2)
	index := 0

	for k, v := range tipsetRewards {
		bulkElements[index*2] = bson.M{"tipset_height": k}
		bulkElements[index*2+1] = v
		index++
	}
	//utils.Log.Traceln(bulkElements)
	_, err := BulkUpsert(nil, TipsetRewardsCollection, bulkElements)
	return err
}

type blockAndRewards struct {
	Height uint64
	Reward *big.Int
}

func (br *blockAndRewards) RewardFil() float64 {
	return utils.ToFil(br.Reward)
}

type MinerBlockRewards struct {
	Miner           string
	TotalReward     *big.Int
	MinedBlockCount uint64
	BlockRewards    []*blockAndRewards
}

func (mmbr *MinerBlockRewards) AddOneBlockReward(height uint64, reward *big.Int) {
	mmbr.MinedBlockCount++
	mmbr.TotalReward.Add(mmbr.TotalReward, reward)
	mmbr.BlockRewards = append(mmbr.BlockRewards, &blockAndRewards{
		Height: height,
		Reward: big.NewInt(0).Set(reward)})
}

func MinerRewardInTimeRange(start, diff uint64, miners []string, is_height bool) (map[string]*MinerBlockRewards, error) {
	ms, c := connect(TipsetRewardsCollection)
	defer ms.Close()

	var trs []*TipsetBlockRewards

	var fieldName string = "time_stamp"
	if is_height {
		fieldName = "tipset_height"
	}
	qMatch := bson.M{fieldName: bson.M{"$gte": start, "$lt": start + diff}}
	qFind := []bson.M{{"$match": qMatch}}

	minerSize := len(miners)
	if minerSize > 0 {
		qMatchOr := make([]bson.M, minerSize)
		for index, miner := range miners {
			qMatchOr[index] = bson.M{fmt.Sprintf(`bson"miners".%s`, miner): bson.M{"$exists": true}}
		}
		qMatch["$or"] = qMatchOr
	}

	// mgo.SetDebug(true)
	// fmt.Printf("%v\n", q_find)
	utils.Log.Traceln(qFind)
	err := c.Pipe(qFind).AllowDiskUse().All(&trs)
	// mgo.SetDebug(false)

	mbrm := make(map[string]*MinerBlockRewards)
	mm := map[string]struct{}{}

	for _, m := range miners {
		mm[m] = struct{}{}
	}

	for _, trw := range trs { // 遍历tipset
		for miner_addr, mb := range trw.Miners { // 遍历tipset中的矿工奖励
			// 过滤不需要的miner地址.
			if _, exist := mm[miner_addr]; !exist {
				continue
			}
			miner, exist := mbrm[miner_addr]
			if !exist {
				miner = &MinerBlockRewards{
					Miner:           miner_addr,
					MinedBlockCount: 0,
					TotalReward:     big.NewInt(0),
					BlockRewards:    []*blockAndRewards{},
				}
				mbrm[miner_addr] = miner
			}
			miner.AddOneBlockReward(trw.TipsetHeight, mb.Rewards.Int)
		}
	}
	return mbrm, err
}

func GetLatestReward() (string, error) {
	ms, c := connect(TipsetRewardsCollection)
	defer ms.Close()
	var trs []*TipsetBlockRewards
	err := c.Find(nil).Sort("-tipset_height").Limit(1).All(&trs)
	if err != nil {
		return "", err
	}
	if len(trs) == 0 {
		return "", nil
	}
	tipsetReward, err := types.BigFromString(trs[0].TipsetReward.String())
	count := types.NewInt(trs[0].TipsetBlockCount)
	reward := types.FIL(types.BigDiv(tipsetReward, count)).String()
	return reward, nil
}
