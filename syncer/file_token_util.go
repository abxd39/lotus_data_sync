package syncer

import (

	// "github.com/filecoin-project/lotus/chain/types"
	// "github.com/filecoin-project/specs-actors/actors/builtin"

)

// // const PrecisionDefault = 8 // float64(0.00001)

// // WEN
// var blocksPerEpoch = big.NewInt(int64(5))

// // 返回每个周期中的奖励filcoin数量和释放的奖励数量
// func (fs *Filscaner) futureBlockRewards(timediff, repeate uint64) ([]*big.Int, *big.Int, error) {
// 	coffer, err := fs.api.WalletBalance(fs.ctx, builtin.RewardActorAddr)
// 	if err != nil {
// 		return nil, nil, err
// 	}
// 	utils.SugarLogger.Infof("\n!!!!!!!net work balance=%.3f", utils.ToFil(coffer.Int))

// 	released := big.NewInt(0).Set(coffer.Int)
// 	blockDaliy := big.NewInt(2 * 60 * 24) // 每日预计出块数量
// 	rewardDaliy := big.NewInt(0)
// 	blockDiff := timediff / 30

// 	sums := make([]*big.Int, repeate)
// 	sum := new(big.Int)

// 	for i := uint64(0); i < repeate; i++ {
// 		sums[i] = big.NewInt(0)

// 		for c := uint64(0); c < blockDiff; c += blockDaliy.Uint64() {
// 			// TODO WEN
// 			// a 应该是当前单个区块奖励
// 			//a := vm.MiningReward(coffer)
// 			a := big.NewInt(0)
// 			a.Mul(a, blocksPerEpoch)
// 			rewardDaliy.Mul(a, blockDaliy)
// 			sum.Add(sum, rewardDaliy)
// 			sums[i].Add(sums[i], rewardDaliy)
// 			coffer.Sub(coffer.Int, rewardDaliy)
// 		}
// 	}

// 	released.Add(released, sum)
// 	return sums, released, nil
// }

// func SelfTipsetRewards(remainingReward *big.Int) *big.Int {
// 	remaining := types.NewInt(0)
// 	remaining.Set(remainingReward)

// 	//TODO WEN
// 	//rewards := vm.MiningReward(remaining)
// 	//return rewards.Mul(rewards.Int, blocksPerEpoch)
// 	return big.NewInt(remaining.Int64())
// }

// func (fs *Filscaner) releasedRewardAtHeight(height uint64) *big.Int {
// 	releaseRewards, err := modelsBlockReleasedRewardsAtHeight(height)
// 	if err != nil {
// 		releaseRewards = &ModelsBlockReward{
// 			Height:          0,
// 			ReleasedRewards: &module.BsonBigint{Int: big.NewInt(0)},
// 		}
// 	}

// 	remainRewards := big.NewInt(0).Sub(TotalRewards, releaseRewards.ReleasedRewards.Int)
// 	skipped := height - releaseRewards.Height

// 	rewards := SelfTipsetRewards(remainRewards)
// 	rewards.Mul(rewards, big.NewInt(int64(skipped)))

// 	return rewards.Add(rewards, releaseRewards.ReleasedRewards.Int)
// }

func (fs *Filscaner) listGenesisMiners() (*TipsetMinerMessages, error) {
	tipset, err := fs.api.ChainGetGenesis(fs.ctx)
	if err != nil {
		return nil, err
	}
	miners, err := fs.api.StateListMiners(fs.ctx, tipset.Key())
	if err != nil {
		return nil, err
	}
	tipestMinerMessages := &TipsetMinerMessages{
		miners: make(map[string]struct{}),
		tipset: tipset}

	for _, v := range miners {
		tipestMinerMessages.miners[v.String()] = struct{}{}
	}

	return tipestMinerMessages, nil
}
