package syncer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/lotus/api"
	po "github.com/filecoin-project/lotus/chain/actors/builtin/power"
	"github.com/filecoin-project/lotus/chain/types"
	"io"
	"io/ioutil"
	"lotus_data_sync/module"
	"lotus_data_sync/utils"
	"net/http"
	"time"
)

func (fs *Filscaner) apiMinerStateAtTipset(minerAddr address.Address, tipset *types.TipSet) (*module.MinerStateAtTipset, error) {
	var (
		power                 *api.MinerPower
		provingSectorSize     = module.NewBigintFromInt64(0)
		PeerId, worker, owner string
		SectorSize            uint64
		WalletAddr            string
		PowerPercent          float64
		err                   error
	)

	fmt.Println(minerAddr.String())
	// TODO:把minerPeerId和MinerSectorSize缓存起来,可以减少 lotus rpc访问量     =>  mongodb

	if Info, err := module.MinerByAddress(minerAddr.String()); err == nil {
		if Info != nil {
			PeerId = Info.PeerId
			SectorSize = Info.SectorSize
			WalletAddr = Info.WalletAddr
		} else {
			if minerInfo, err := fs.api.StateMinerInfo(fs.ctx, minerAddr, tipset.Key()); err == nil {
				if minerInfo.PeerId == nil {
					return nil, err
				}
				if minerInfo.PeerId != nil {
					PeerId = minerInfo.PeerId.String()
					SectorSize = uint64(minerInfo.SectorSize)
					WalletAddr = minerInfo.Owner.String()
				}
			}
		}
	}

	// Sector size
	mi, err := fs.api.StateMinerInfo(fs.ctx, minerAddr, types.EmptyTSK)
	if err != nil {
		return nil, err
	}
	worker = mi.Worker.String()
	owner = mi.Owner.String()
	mact, err := fs.api.StateGetActor(fs.ctx, minerAddr, types.EmptyTSK)

	//tbs := blockstore.NewTieredBstore(blockstore.NewAPIBlockstore(fs.api), blockstore.NewMemory())
	//mas, err := miner.Load(adt.WrapStore(fs.ctx, cbor.NewCborStore(tbs)), mact)
	secCounts, err := fs.api.StateMinerSectorCount(fs.ctx, minerAddr, types.EmptyTSK)
	if err != nil {
		return nil, err
	}

	proving := secCounts.Active + secCounts.Faulty
	nfaults := secCounts.Faulty
	fmt.Printf("\tCommitted: %s\n", types.SizeStr(types.BigMul(types.NewInt(secCounts.Live), types.NewInt(uint64(mi.SectorSize)))))
	if nfaults == 0 {
		fmt.Printf("\tProving: %s\n", types.SizeStr(types.BigMul(types.NewInt(proving), types.NewInt(uint64(mi.SectorSize)))))
	} else {
		var faultyPercentage float64
		if secCounts.Live != 0 {
			faultyPercentage = float64(10000*nfaults/secCounts.Live) / 100.
		}
		fmt.Printf("\tProving: %s (%s Faulty, %.2f%%)\n",
			types.SizeStr(types.BigMul(types.NewInt(proving), types.NewInt(uint64(mi.SectorSize)))),
			types.SizeStr(types.BigMul(types.NewInt(nfaults), types.NewInt(uint64(mi.SectorSize)))),
			faultyPercentage)
	}

	fmt.Printf("Miner Balance:    %s\n", color.YellowString("%s", types.FIL(mact.Balance).Short()))
	if power, err = fs.api.StateMinerPower(fs.ctx, minerAddr, tipset.Key()); err != nil {
		errMessage := err.Error()
		if errMessage == "failed to get miner power from chain (exit code 1)" {
			utils.SugarLogger.Errorf("get miner(%s) power failed, message:%s\n", minerAddr.String(), errMessage)
			if power, err = fs.api.StateMinerPower(fs.ctx, address.Undef, tipset.Key()); err == nil {
				power.MinerPower = po.Claim{RawBytePower: abi.NewStoragePower(0), QualityAdjPower: abi.NewStoragePower(0)}
			}
		}

		if err != nil {
			utils.SugarLogger.Errorf("get miner(%s) power failed, message:%s\n", minerAddr.String(), err.Error())
			return nil, err
		}

	}
	//sectors
	// 这里应该是把错误的数据使用最近的数据来代替19807040628566131532430835712
	if len(power.TotalPower.RawBytePower.String()) >= 29 {
		if fs.latestTotalPower != nil {
			power.TotalPower.RawBytePower.Set(fs.latestTotalPower)
		} else {
			power.TotalPower.RawBytePower.SetUint64(0)
		}
	} else {
		if fs.latestTotalPower == nil {
			fs.latestTotalPower = big.NewInt(0).Int
		}
		fs.latestTotalPower.Set(power.TotalPower.RawBytePower.Int)
	}
	if power.TotalPower.RawBytePower.Uint64() == 0 {
		PowerPercent = 0.0
	} else {
		rpercI := types.BigDiv(types.BigMul(power.MinerPower.RawBytePower, types.NewInt(1000000)), power.TotalPower.RawBytePower)
		PowerPercent = float64(rpercI.Int64()) / 10000
	}
	miner := &module.MinerStateAtTipset{
		Worker:            worker,
		Owner:             owner,
		PeerId:            PeerId,
		MinerAddr:         minerAddr.String(),
		Power:             module.NewBigInt(power.MinerPower.RawBytePower.Int), //types.SizeStr(pow.MinerPower.RawBytePower)
		TotalPower:        module.NewBigInt(power.TotalPower.RawBytePower.Int), //TotalPower.RawBytePow
		SectorSize:        SectorSize,
		WalletAddr:        WalletAddr,
		SectorCount:       uint64(0),
		TipsetHeight:      uint64(tipset.Height()),
		ProvingSectorSize: provingSectorSize,
		PowerPercent:      PowerPercent,
		MineTime:          tipset.MinTimestamp(),
	}
	//	utils.Log.Tracef("debug 新怎加的字段没有加进去 %+v", miner)
	if err = module.UpsertMinerStateInTipset(miner); err != nil {
		utils.Log.Errorln(err)
	}

	//upsert_pairs := make([]interface{}, 2)
	//upsert_pairs[0] = bson.M{"peer_id": miner.PeerId}
	//upsert_pairs[1] = bson.M{"$set": bson.M{"worker": miner.Worker, "owner": miner.Owner}}
	//
	//result, err := module.UpdateAll(module.MinerCollection, upsert_pairs)
	//if err != nil {
	//	utils.Log.Errorln(err)
	//}
	//utils.Log.Tracef("peer_id=%v 更新结果%+v", miner.PeerId, result)

	return miner, nil
}

func (fs *Filscaner) apiTipset(tpstk string) (*types.TipSet, error) {
	tipsetk := utils.Tipsetkey_from_string(tpstk)
	if tipsetk == nil {
		return nil, fmt.Errorf("convert string(%s) to tipsetkey failed", tpstk)
	}

	return fs.api.ChainGetTipSet(fs.ctx, *tipsetk)
}

func (fs *Filscaner) apiChildTipset(tipset *types.TipSet) (*types.TipSet, error) {
	if tipset == nil {
		return nil, nil
	}

	fs.mutexForNumbers.Lock()
	var header_height = fs.headerHeight
	fs.mutexForNumbers.Unlock()

	for i := uint64(tipset.Height()) + 1; i < header_height; i++ {
		if child, err := fs.api.ChainGetTipSetByHeight(fs.ctx, abi.ChainEpoch(i), types.EmptyTSK); err != nil {
			return nil, err
		} else {
			if child.Parents().String() == tipset.Key().String() {
				return child, nil
			} else {
				return nil, fmt.Errorf("child(%d)'s parent key(%s) != key(%d, %s)\n",
					child.Height(), child.Parents().String(),
					tipset.Height(), tipset.Key().String())
			}

		}
	}
	return nil, errors.New("not found")
}

func (fs *Filscaner) apiMinerDeadLines(minerAddr address.Address, tipset *types.TipSet) {

}

func colorTokenAmount(format string, amount abi.TokenAmount) {
	if amount.GreaterThan(big.Zero()) {
		color.Green(format, types.FIL(amount).Short())
	} else if amount.Equals(big.Zero()) {
		color.Yellow(format, types.FIL(amount).Short())
	} else {
		color.Red(format, types.FIL(amount).Short())
	}
}

//发送GET请求
//url:请求地址
//response:请求返回的内容
func Get(url string) (response string) {
	client := http.Client{Timeout: 5 * time.Second}
	resp, error := client.Get(url)
	defer resp.Body.Close()
	if error != nil {
		panic(error)
	}
	var buffer [512]byte
	result := bytes.NewBuffer(nil)
	for {
		n, err := resp.Body.Read(buffer[0:])
		result.Write(buffer[0:n])
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
	}
	response = result.String()
	return response
}

//发送POST请求
//url:请求地址		data:POST请求提交的数据		contentType:请求体格式，如：application/json
//content:请求返回的内容
func Post(url string, data interface{}, contentType string) (content string) {
	jsonStr, _ := json.Marshal(data)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Add("content-type", contentType)
	if err != nil {
		panic(err)
	}
	defer req.Body.Close()
	client := &http.Client{Timeout: 5 * time.Second}
	resp, error := client.Do(req)
	if error != nil {
		panic(error)
	}
	defer resp.Body.Close()
	result, _ := ioutil.ReadAll(resp.Body)
	content = string(result)
	return
}

/*
func MiningReward(remainingReward types.BigInt) types.BigInt {
	ci := big.NewInt(0).Set(remainingReward.Int)
	res := ci.Mul(ci, build.InitialReward)
	res = res.Div(res, miningRewardTotal.Int)
	res = res.Div(res, blocksPerEpoch.Int)
	return types.BigInt{res}
}
*/

// func (fs *Filscaner) QyChat(ctx context.Context, req *filscanproto.ChatReq) (*filscanproto.ChatResp, error) {
// 	res := &filscanproto.ChatResp{
// 		Code: 10001,
// 		Msg:  "success",
// 	}
// 	//op := req.GetOp()
// 	//param := bytes.NewBufferString("大家好！接口测试看到忽略。")
// 	//notify.SendQyMessage(int(op), param.Bytes())

// 	utils.Log.Traceln("line=", req.Line, " addr=", req.Addr, "cid=", req.Cid)
// 	//t := time.Now().Unix() + TimeAfter(int64(req.Line), req.Addr)
// 	cidd, err := cid2.Decode(req.Cid)
// 	if err != nil {
// 		utils.Log.Errorln(err)
// 		return nil, err
// 	}
// 	ms, err := fs.api.ChainGetMessage(ctx, cidd)
// 	if err != nil {
// 		utils.Log.Errorln(err)
// 		return nil, err
// 	}
// 	fs.Temp(ms, 208280, "", "")
// 	b, err := ms.MarshalJSON()
// 	if err != nil {
// 		utils.Log.Errorln(err)
// 		return nil, err
// 	}
// 	res.Msg = string(b)
// 	res.CurrTime = res.Msg
// 	fs.Ali(req.Addr)
// 	return res, nil
// }

// func (fs *Filscaner) Ali(addr string) {
// 	err := AliyunVoiceSignPhon("15920038315")
// 	//err := AliyunVoice(addr)
// 	if err != nil {
// 		utils.Log.Errorln(err)
// 	}

// }
