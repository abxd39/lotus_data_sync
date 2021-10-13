package syncer

import (
	"context"
	"encoding/json"

	"lotus_data_sync/module"

	"fmt"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/build"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/actors/builtin"
	"io/ioutil"
	"log"
	"lotus_data_sync/utils"
	"net/http"

	"errors"
	
	"github.com/hako/durafmt"
	cid2 "github.com/ipfs/go-cid"
	
	"lotus_data_sync/force/factors"
	"time"
)

type MinerStateRecord struct {
	Id     string                     `bson:"_id" json:"id"`
	Record *module.MinerStateAtTipset `bson:"record" json:"record"`
}

type MinerStateRecordInterface struct {
	Id     string      `bson:"_id" json:"id"`
	Record interface{} `bson:"record" json:"record"`
}

type MinerIncreasedPowerRecord struct {
	IncreasedPower uint64                     `bson:"increased_power" json:"increased_power"`
	Record         *module.MinerStateAtTipset `bson:"record" json:"record"`
}

type MinerBlockRecord struct {
	Blockcount uint64                     `bson:"block_count" json:"block_count"`
	Record     *module.MinerStateAtTipset `bson:"record" json:"record"`
}

type MinedBlock struct {
	Miner      string `bson:"miner" json:"miner"`
	BlockCount uint64 `bson:"mined_block_count" json:"mined_block_count"`
}

func ParseActorMessage(message *types.Message) (*factors.ActorInfo, *factors.MethodInfo, error) {
	ErrActorNotFound := errors.New("can't found actor")
	ErrMethodNotFound := errors.New("can't found method in actor")
	if message.Method == 0 {
		return nil, nil, ErrActorNotFound
	}
	actor, exist := factors.LookupByAddress(message.To)
	if !exist {
		if actor, exist = factors.Lookup(builtin.StorageMinerActorCodeID); !exist {
			return nil, nil, ErrActorNotFound
		}
	}

	method, exist := actor.LookupMethod(uint64(message.Method))
	if !exist {
		return nil, nil, ErrMethodNotFound
	}

	return &actor, &method, nil
}

// type InterfaceParam struct {
// 	IsSend                 bool
// 	NotifyList             string
// 	m                      *types.Message
// 	winPost                *miner.SubmitWindowedPoStParams
// 	Term                   *miner.TerminateSectorsParams
// 	Recovered              *miner.DeclareFaultsRecoveredParams
// 	Faults                 *miner.DeclareFaultsParams
// 	ProvCommitC2           *miner.ProveCommitSectorParams
// 	PreCommitP2            *miner.PreCommitSectorParams
// 	PreCommitP2Batch       *miner.PreCommitSectorBatchParams
// 	ProveCommitC2Aggregate *miner.ProveCommitAggregateParams
// }

// //TODO 告警监控入口
// func (p *InterfaceParam) DecoderMessageParam(cids, methodName string, blockHeight uint64, b bool) string {
// 	ctx := context.TODO()
// 	cid, err := cid2.Decode(cids)
// 	if err != nil {
// 		utils.Log.Errorln(err)
// 	}
// 	msg, err := utils.LotusApi.ChainGetMessage(ctx, cid)
// 	if err != nil {
// 		utils.Log.Errorln(err)
// 		return ""
// 	}

// 	result, err := utils.LotusApi.StateDecodeParams(ctx, msg.To, msg.Method, msg.Params, types.EmptyTSK)
// 	mjson, _ := json.Marshal(result)
// 	mString := string(mjson)

// 	return mString
// }

// func (p *InterfaceParam) IsOk(add string) bool {
// 	for k, _ := range utils.MinerAddr {
// 		if k == add {
// 			return true
// 		}
// 	}
// 	return false

// }
// func (p *InterfaceParam) MessageType(Method int) bool {
// 	switch Method {
// 	case 5, 6, 7, 11, 25, 26:
// 		return true
// 	default:
// 		return false
// 	}
// }

// //消息推送
// //上层过滤需要发送的miner

// func (p *InterfaceParam) NotifyQyChat(addr, cid, tp string, param []byte) {
// 	switch tp {
// 	case "DeclareFaultsRecovered": //扇区恢复中
// 		p.DeclareFaultsRecoveredUnmarshalParam(addr, cid, param)
// 	case "TerminateSectors": //手动移除扇区
// 		p.TerminateSectorsUnmarshalParam(addr, cid, param)
// 	case "SubmitWindowedPoSt": //时空抽查 需要解码
// 		p.SubmitWindowedPoStUnmarshalParam(addr, cid, param)
// 	case "DeclareFaults": //版本问题。 该消息类型只存在于旧版本
// 		p.DeclareFaultsUnmarshalParam(addr, cid, param)
// 	case "PreCommitSector": //p2 单个扇区 Method=6
// 		// p.BaseUpP2(addr)
// 		// p.SignEfficient(addr, cid, 6, param)
// 	case "ProveCommitSector": //c2 单个扇区 Method=7
// 		// p.BaseUpC2(addr, param)
// 		// p.SignEfficient(addr, cid, 7, param)
// 	case "PreCommitSectorBatch": //p2 集合Method=25
// 		// p.BaseUpP2(addr)
// 		// p.SignEfficient(addr, cid, 25, param)
// 	case "ProveCommitAggregate": //C2 集合Method=26
// 		// p.BaseUpC2(addr, param)
// 		// p.SignEfficient(addr, cid, 26, param)
// 	default:
// 		utils.Log.Traceln("其他消息类型", tp)

// 	}
// }
// func (p *InterfaceParam) DeclareFaultsUnmarshalParam(addr, cid string, param []byte) error {
// 	p.Faults = new(miner.DeclareFaultsParams)
// 	if err := json.Unmarshal(param, p.Faults); err != nil {
// 		utils.Log.Errorln(err)
// 		return err
// 	}
// 	if len(p.Faults.Faults) != 0 {
// 		//出错了
// 		utils.Log.Tracef("%v", string(param))
// 	}
// 	//发消息
// 	//notify.SendQyMessage(4, param, "", 0)
// 	return nil
// }

// func (p *InterfaceParam) SubmitWindowedPoStUnmarshalParam(addr, cid string, param []byte) error {
// 	p.IsSend = true
// 	p.winPost = new(miner.SubmitWindowedPoStParams)
// 	if err := json.Unmarshal(param, p.winPost); err != nil {
// 		utils.Log.Errorln(err)
// 		return err
// 	}

// 	if len(p.winPost.Partitions) != 0 {
// 		count := uint64(0)
// 		var err error

// 		for _, v := range p.winPost.Partitions {
// 			count, err = v.Skipped.Count()
// 			if err != nil {
// 				utils.Log.Errorln(err)
// 				continue
// 			}
// 		}

// 		if count != 0 {
// 			//消息先发
// 			utils.Log.Tracef("扇区错误消息提示cid=%s addr=%s %+v", cid, addr, p.winPost)
// 			fauls := p.CheckSectorsState(addr, p.winPost.Deadline)
// 			strTime := time.Unix(BeginTime(int64(p.winPost.Deadline), addr), 0).Format(utils.TimeString)
// 			addrOwner := utils.GetAddrOwner(addr)
// 			p.SectorErr(addr, addrOwner, fmt.Sprintf("%d", p.winPost.Deadline), fmt.Sprintf("%d", fauls), strTime)
// 		}

// 	}

// 	//需要监听下一个窗口期 在开始后 5分钟内没有消息上来就报警
// 	currLine := int64(p.winPost.Deadline)
// 	//TODO
// 	//时空证明消息监听
// 	if err := p.WinPoStSetKekEven(currLine, addr); err != nil {
// 		if err.Error() == "errorSec" {
// 			utils.Log.Errorf("%+v", p.winPost)
// 		}
// 	}
// 	//需要发送时空抽查告警
// 	//消息上链发送通知
// 	//并且是有过扇区恢复的消息上链的 才需要通知
// 	key := fmt.Sprintf(utils.RedisKeyWindowPostMark, addr, currLine)
// 	if redisdb.IsExistDB1(key) == 1 {
// 		utils.Log.Traceln("消息上链发送通知 addr=", addr)
// 		addrOwner := utils.GetAddrOwner(addr)
// 		p.WindowedPoSt(addr, addrOwner, int64(p.winPost.Deadline), true)
// 	}

// 	//检查数据库中有没有告警的
// 	minerInfo := new(module.MinerDeadLine)
// 	minerInfo.Miner = addr
// 	minerInfo.Deadline = p.winPost.Deadline
// 	minerInfo.ReMove()
// 	return nil
// }

// func (p *InterfaceParam) WinPoStSetKekEven(line int64, addr string) error {
// 	NextLine := int64(line + 1)
// 	if NextLine > 47 {
// 		NextLine = 0
// 	}
// 	sec := TimeAfter(NextLine, addr)
// 	if sec <= 0 {
// 		utils.Log.Errorf("有效时间小于0 add=%s line=%d", addr, NextLine)
// 		return fmt.Errorf("errorSec")
// 	}
// 	sec += 300
// 	keyNext := fmt.Sprintf(utils.RedisKeySectorErr, addr, NextLine)
// 	keyCurr := fmt.Sprintf(utils.RedisKeySectorErr, addr, line)
// 	redisdb.DeleteKeyDB0(keyCurr) //有消息上来需要删除redis key
// 	redisdb.Publish(keyNext, sec) //监控下一个上链消息 需要判断下一窗口的的扇区总数
// 	return nil
// }

// //第 line 个窗口 抽查的开始时间 如果需要计算下一次该窗口的抽查时间需要判断日期
// func TimeAfter(line int64, addr string) int64 {
// 	if line > 47 {
// 		line = 0
// 	}
// 	begin := BeginTime(line, addr)
// 	//utils.Log.Tracef("第%d个窗口的开始时间为%s", line, time.Unix(begin, 0).Format(utils.TimeString))
// 	//second := begin - int64(time.Now().Unix())
// 	second := begin + 24*3600
// 	//utils.Log.Tracef("下一次 第%d个窗口的开始时间为%s", line, time.Unix(second, 0).Format(utils.TimeString))
// 	if second < 10*60 {
// 		utils.Log.Errorf("%s 的第 %d 窗口的抽查时间计算错误", addr, line)
// 	}
// 	//
// 	if second <= 0 {
// 		utils.Log.Errorf("%s 的第 %d 窗口的抽查时间计算错误 时间为%d", addr, line, second)
// 	}

// 	return second
// }
func BeginTime(line int64, addr string) int64 {
	line0 := DeadLineOpenTien(addr)
	begin := line0 + line*1800
	return begin
}

func DeadLineOpenTien(addr string) int64 {
	ctx := context.TODO()
	a, err := address.NewFromString(addr)
	if err != nil {
		utils.Log.Errorln(err, "   "+addr)
		return 0
	}
	di, err := utils.LotusApi.StateMinerProvingDeadline(ctx, a, types.EmptyTSK)
	if err != nil {
		utils.Log.Errorln(err, "   "+addr)
		return 0
	}
	//计算第0个抽查窗口
	DeadlineInde := int64(di.Index * 1800)
	DeadlineOpen := EpochTime(di.CurrentEpoch, di.Open)
	sec := DeadlineOpen - DeadlineInde
	if sec <= 0 {
		utils.Log.Errorf("addr=%s,第0个抽查窗口的时间为%s, 当前窗口为第%d窗口的时间为%s ", addr, time.Unix(sec, 0).Format(utils.TimeString), di.Index, time.Unix(DeadlineOpen, 0).Format(utils.TimeString))
	}
	//utils.Log.Errorf("addr=%s,第0个抽查窗口的时间为%s, 当前窗口为第%d窗口的时间为%s ", addr, time.Unix(sec, 0).Format(utils.TimeString), di.Index, time.Unix(DeadlineOpen, 0).Format(utils.TimeString))
	return sec
}

func EpochTime(curr, e abi.ChainEpoch) int64 {
	unix := time.Now().Unix()
	switch {
	case curr > e:
		durafmt.Parse(time.Second * time.Duration(int64(build.BlockDelaySecs)*int64(e-curr))).Duration()
		//utils.Log.Tracef("datetime=%s total=%s", time.Unix(unix, 0).Format(utils.TimeString), time.Unix(unix-int64(temp.Seconds()), 0).Format(utils.TimeString))
		//	utils.Log.Traceln(temp)
		return unix - int64(durafmt.Parse(time.Second*time.Duration(int64(build.BlockDelaySecs)*int64(curr-e))).Duration().Seconds())
	case curr == e:
		//utils.Log.Traceln(time.Unix(unix, 0).Format(utils.TimeString))
		return unix
	case curr < e:
		//return unix + int64(durafmt.Parse(time.Second*time.Duration(int64(build.BlockDelaySecs)*int64(e-curr))).LimitFirstN(2).Duration().Seconds())

		durafmt.Parse(time.Second * time.Duration(int64(build.BlockDelaySecs)*int64(e-curr))).Duration()
		//utils.Log.Tracef("datetime=%s total=%s", time.Unix(unix, 0).Format(utils.TimeString), time.Unix(int64(temp.Seconds())+unix, 0).Format(utils.TimeString))
		//utils.Log.Traceln(temp)
		return unix + int64(durafmt.Parse(time.Second*time.Duration(int64(build.BlockDelaySecs)*int64(e-curr))).Duration().Seconds())
	}
	return 0
}

// func (p *InterfaceParam) TerminateSectorsUnmarshalParam(addr, cid string, param []byte) error {
// 	p.Term = new(miner.TerminateSectorsParams)
// 	if err := json.Unmarshal(param, p.Term); err != nil {
// 		utils.Log.Errorln(err)
// 		return err
// 	}
// 	//不知道能不能解析出来
// 	//utils.Log.Tracef("%+v", p.Term)
// 	return nil
// }

//所有的该消息上链都要发
//因为有错误了才会有该消息
// func (p *InterfaceParam) DeclareFaultsRecoveredUnmarshalParam(addr, cid string, param []byte) error {
// 	p.Recovered = new(miner.DeclareFaultsRecoveredParams)
// 	if err := json.Unmarshal(param, p.Recovered); err != nil {
// 		utils.Log.Errorln(err)
// 		return err
// 	}
// 	p.IsSend = true
// 	for _, v := range p.Recovered.Recoveries {
// 		line := int64(v.Deadline)
// 		sectors, err := v.Sectors.Count()
// 		if err != nil {
// 			utils.Log.Errorln(err)
// 		}
// 		lines := fmt.Sprintf("%d", line)
// 		key := fmt.Sprintf(utils.RedisSectorErr20s, addr, lines)
// 		redisdb.SetNXDB1(key, "10000", 0)
// 		if err = redisdb.Expire(key, 3600*24); err != nil {
// 			utils.Log.Errorln(err, "  ", addr)
// 		}
// 		keyTime := fmt.Sprintf(utils.RedisSectorErrFaultsTime, addr, lines, sectors)
// 		if err = redisdb.Expire(keyTime, 3600*24); err != nil {
// 			utils.Log.Errorln(err, "  ", addr)
// 		}
// 		//设置时空证明消息 是否发送的标记
// 		keydecover := fmt.Sprintf(utils.RedisKeyWindowPostMark, addr, line)
// 		redisdb.SetNXDB1(keydecover, "是否发送时空证明上链消息", 24*3600)
// 		addrOwner := utils.GetAddrOwner(addr)
// 		p.DeclareFaultsRecoveredNotify(line, int64(sectors), addr, addrOwner, true)
// 		//删除RedisKey 删除监听事件 如果不删除 就会发送消息未上链 告警
// 		keyDeclare := fmt.Sprintf(utils.RedisKeyDeclareFaultsRecovered, addr, v.Deadline)
// 		redisdb.DeleteKeyDB0(keyDeclare)
// 		//监控 时空抽查消息上链

// 		//if err = p.WinPoStSetKekEven(int64(v.Deadline), addr); err != nil {
// 		//	if err.Error() == "errorSec" {
// 		//		utils.Log.Errorln("%+v", p.Recovered)
// 		//	}
// 		//
// 		//}
// 	}

// 	//不知道能不能解析出来
// 	//	utils.Log.Tracef("%+v", p.Recovered)
// 	return nil
// }

// func (p *InterfaceParam) BuildDate(addr string, line int) {

// }

//下一次该窗口的消息上链事件 监听
func (fs *Filscaner) DeclareFaultsRecovered(deadLine int64, addr string) {
	//设置扇区恢复消息事件监听
	//一天一次窗口检测，如果是要设置多次报警 在设置 默认 15 分钟内三次
	openTime := BeginTime(deadLine, addr)
	//下一次该窗口抽查前半5分钟触发
	openTime = openTime + 24*3600 - 5*60
	//需要在多少秒后告警
	sec := openTime - time.Now().Unix()
	utils.Log.Tracef(" addr=%s deadline=%d 下一次该窗口 扇区恢复消息上链时间 计算得出 %s redis 过期时间为%d", addr, deadLine, time.Unix(openTime, 0).Format(utils.TimeString), sec)
	key := fmt.Sprintf(utils.RedisKeyDeclareFaultsRecovered, addr, deadLine)
	_ = key
	//订阅事件
	// if redisdb.SetNXDB0(key, sec) {
	// 	unix := time.Now().Unix() + sec
	// 	utils.Log.Traceln("事件触发时间点", time.Unix(unix, 0).Format(utils.TimeString), key)
	// }

}

type SectorErr struct {
	Tittle       string `json:"tittle"`
	Miner        string `json:"miner"`
	SectorsFault uint64 `json:"sectors"`
	Deadline     uint64 `json:"deadline"`
	DeadlineTime string `json:"deadline_time"`
}

//定时任务 5 分钟发一次

func getTime(deadline, sex int64) string {
	unix := sex + deadline*1800
	return time.Unix(unix, 0).Format(utils.TimeString)
}

//lotus 同步监听
func (fs *Filscaner) LotusSyncWait() {
	ctx := context.TODO()
	state, err := fs.api.SyncState(ctx)
	if err != nil {
		utils.Log.Errorln(err)
		return
	}
	height := 0
	htdiff := 0
	for _, ss := range state.ActiveSyncs {
		var base, target []cid2.Cid
		var heightDiff int64
		var theight abi.ChainEpoch
		if ss.Base != nil {
			base = ss.Base.Cids()
			heightDiff = int64(ss.Base.Height())
		}
		if ss.Target != nil {
			target = ss.Target.Cids()
			heightDiff = int64(ss.Target.Height()) - heightDiff
			theight = ss.Target.Height()
		} else {
			heightDiff = 0
		}

		if height < int(ss.Height) {
			height = int(ss.Height)
			htdiff = int(heightDiff)
		}
		//utils.Log.Tracef("\tBase:\t%s\n", base)
		//utils.Log.Tracef("\tTarget:\t%s (%d)\n", target, theight)
		//utils.Log.Tracef("\tHeight diff:\t%d\n", heightDiff)
		//utils.Log.Tracef("\tStage: %s\n", ss.Stage)
		//utils.Log.Tracef("\tHeight: %d\n", ss.Height)
		if ss.End.IsZero() {
			if !ss.Start.IsZero() {
				//utils.Log.Tracef("\tElapsed: %s\n", time.Since(ss.Start))
			}
		} else {
			//utils.Log.Tracef("\tElapsed: %s\n", ss.End.Sub(ss.Start))
		}
		if ss.Stage == api.StageSyncErrored {
			utils.Log.Tracef("\tError: %s\n", ss.Message)
		}
		_ = target
		_ = base
		_ = theight
		_ = htdiff

	}

	// if htdiff > 3 {
	// 	utils.TipMark = false
	// 	redisdb.SetNXDB1(utils.RedisLotusSyncWait, "lotus sync staus", 30*24*3600)
	// 	fs.LotusSyncWaitSendMessage(height, htdiff)
	// } else {
	// 	utils.TipMark = true
	// 	if redisdb.IsExistDB1(utils.RedisLotusSyncWait) == 1 {
	// 		redisdb.DeleteKeyDB1(utils.RedisLotusSyncWait)
	// 		fs.LotusSyncWaitSendMessage(height, htdiff)
	// 	}
	// }

	return
}

func (fs *Filscaner) tipSetHeight() int64 {
	result := struct {
		Code    int    `json:"code"`
		Message string `json:"msg"`
		Result  []struct {
			LatestHeight int64 `json:"latest_height"`
			TipSetHeight int64 `json:"tipSetHeight"`
		} `json:"result"`
	}{}
	url := "https://api.fgas.io/api/v1/fil?type=64G"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		log.Println(err)
		return 0
	}
	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return 0
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return 0
	}
	if err = json.Unmarshal(body, &result); err != nil {
		log.Println(err)
		return 0
	}
	if result.Code == 0 {
		for _, v := range result.Result {
			return v.LatestHeight
		}

	}
	return 0
}
