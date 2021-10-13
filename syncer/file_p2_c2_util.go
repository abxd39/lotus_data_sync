package syncer

// import (
// 	"encoding/json"
// 	"filscan_lotus/models"
// 	"filscan_lotus/notify"
// 	"filscan_lotus/redisdb"
// 	"filscan_lotus/utils"
// 	"fmt"
// 	"github.com/filecoin-project/specs-actors/v5/actors/builtin/miner"
// 	"strconv"
// 	"strings"
// 	"time"
// )

// func (p *InterfaceParam) PreCommitSectorUnmarshalParams(addr, cid string, param []byte) {
// 	p.IsSend = true

// }
// func (p *InterfaceParam) Is32or64(addr string) int64 {
// 	if v, ok := utils.Map32Or64[addr]; ok {
// 		return v
// 	}
// 	return 0
// }

// func (p *InterfaceParam) IsSet(addr string) bool {
// 	if v, ok := utils.MapSet[addr]; ok {
// 		return v
// 	}
// 	return false
// }

// func (p *InterfaceParam) C2(addr string) {
// 	//if 32 or 64
// }
// func (p *InterfaceParam) BaseUpC2(addr string, param []byte) {
// 	//utils.Log.Tracef("c2 baseUp64")
// 	switch p.IsSet(addr) {
// 	case true:
// 		//
// 		p.SetC2(addr, param)
// 	case false:
// 		//单台效率
// 		p.NoSetC2(addr)
// 	default:
// 	}
// }
// func (p *InterfaceParam) BaseUpP2(addr string) {
// 	//utils.Log.Tracef("P2 baseUp64")
// 	switch p.IsSet(addr) {
// 	case true:
// 		//
// 		p.SetP2(addr)
// 	case false:
// 		//
// 		p.NoSetP2(addr)
// 	default:
// 	}
// }
// func (p *InterfaceParam) SetP2(addr string) {

// }
// func (p *InterfaceParam) NoSetP2(addr string) {
// 	//utils.Log.Traceln("NoSetP2")
// 	//不区分32还是64
// 	//删除上一次设置的key_even
// 	key := fmt.Sprintf(utils.P2_PreExpire, addr)
// 	redisdb.DeleteKeyDB0(key)
// 	//设置下一次key_even
// 	redisdb.SetNXDB0(key, 2.5*3600)

// }

// func (p *InterfaceParam) SetC2(addr string, param []byte) {
// 	//utils.Log.Tracef("c2 Set 消息监控")
// 	p.ProveCommitC2Aggregate = new(miner.ProveCommitAggregateParams)
// 	if err := json.Unmarshal(param, p.ProveCommitC2Aggregate); err != nil {
// 		utils.Log.Errorln(err)
// 		return
// 	}
// 	count, err := p.ProveCommitC2Aggregate.SectorNumbers.Count()
// 	if err != nil {
// 		utils.Log.Errorln(err)
// 		return
// 	}
// 	confKey := addr + "Machine"
// 	MachineCount, err := utils.Initconf.Int64(confKey)
// 	if err != nil {
// 		utils.Log.Traceln(err, "---------", addr)
// 		return
// 	}
// 	//在此时间内会有消息上链
// 	h := 24*3600/(float64(MachineCount*56)/float64(count)) + 1800
// 	key := fmt.Sprintf(utils.C2SETExpire, addr, int(h))
// 	redisdb.DeleteKeyDB0(key)
// 	redisdb.SetNXDB0(key, int64(h))
// 	utils.Log.Tracef("%s ProveCommitAggregate 消息上链时间计算的到%.4f 秒 为%d 分钟", addr, h, int(h/60))

// }
// func (p *InterfaceParam) NoSetC2(addr string) {
// 	//utils.Log.Tracef("c2 Noset 消息监控")
// 	//消息上链的时间
// 	Threshold := int64(0)
// 	confKey := addr + "Machine"
// 	count, err := utils.Initconf.Int64(confKey)
// 	if err != nil {
// 		//utils.Log.Traceln(err, "---------", addr)
// 		return
// 	}
// 	if count == 0 {
// 		utils.Log.Tracef("%s 的C2上链情况不再监控节点内。。", addr)
// 		return
// 	}
// 	switch p.Is32or64(addr) {
// 	case 64:
// 		FloatThreshold := (24 * 60) / float64(count*56) //每分钟的消息上链数

// 		Threshold = int64(utils.MonitorTime / FloatThreshold)
// 		//utils.Log.Tracef("c2-64-noSet-------%v   (%d)分钟内的消息数量为%v,大约在%.2f分钟内有一条消息上链", addr, utils.MonitorTime, Threshold, FloatThreshold)
// 	case 32:
// 		//
// 		FloatThreshold := (24 * 60) / float64(count*176) //每分钟的消息上链数
// 		Threshold = int64(utils.MonitorTime / FloatThreshold)
// 		//utils.Log.Tracef("c2-32-noSet-------%v   (%d)分钟内的消息数量为%v,大约在%.2f分钟内有一条消息上链", addr, utils.MonitorTime, Threshold, FloatThreshold)
// 	default:
// 		return
// 	}

// 	key := fmt.Sprintf(utils.C2_64NoSetTotal, addr)
// 	//utils.Log.Tracef("addr=%s key=%s", addr, key)
// 	redisdb.IncrDB1(key, 7*24*3600) //加一操作
// 	key1 := fmt.Sprintf(utils.C2_64Expire, addr)
// 	//utils.Log.Tracef("addr=%s key=%s key1=%s", addr, key, fmt.Sprintf(utils.C2_64NoSetDuring, addr))
// 	redisdb.SetNXDB0(key1, utils.MonitorTime*60) //过期了就设置一次
// 	key = fmt.Sprintf(utils.C2_64NoSet, addr)
// 	redisdb.GetSetDB1(key, int(Threshold))
	
// }

// //redis 事件触发
// func (p *InterfaceParam) OpMessageC2NotSet(v string) {
// 	KeyType := strings.Split(v, "_")
// 	if len(KeyType) < 3 {
// 		utils.Log.Errorln("未知类型", v)
// 		return
// 	}
// 	addr := KeyType[2]
// 	Key := fmt.Sprintf(utils.C2_64NoSetTotal, addr) //目前未知的消息总数
// 	max := redisdb.GetValueIntDB1(Key)
// 	Key = fmt.Sprintf(utils.C2_64NoSet, addr)
// 	Threshold := redisdb.GetValueIntDB1(Key)
// 	if Threshold == 0 {
// 		utils.Log.Errorf("add=%s逻辑错误", addr)
// 		return
// 	}
// 	Key = fmt.Sprintf(utils.C2_64NoSetDuring, addr) //截止上一次的数量
// 	min := redisdb.GetSetDB1(Key, int(max))
// 	if max == 0 {
// 		// 忽略本次
// 		return
// 	}
// 	diff := max - int64(min)
// 	//utils.Log.Tracef("C2 事件触发 %s 单位时间内的消息上链数量为%d max=%d min=%d 单位时间内应上链的消息数量为Threshold=%d", addr, diff, max, min, Threshold)
// 	if diff < Threshold {
// 		//需要告警
// 		un := time.Now().Unix()
// 		un5mintuAgo := un - utils.MonitorTime*60
// 		un5mintuAgeTime := time.Unix(un5mintuAgo, 0).Format(utils.TimeString)
// 		unStr := time.Unix(un, 0).Format(utils.TimeString)
// 		str := fmt.Sprintf("%s------%s", strings.ReplaceAll(un5mintuAgeTime, "-", "."), strings.ReplaceAll(unStr, "-", "."))
// 		//p.Notify(addr, str, "C2")

// 		addrOwn := utils.GetAddrOwner(addr)
// 		admin := utils.Getadmin(addr)
// 		mss := fmt.Sprintf("%s %s （非集合消息）\n%s\n应上链消息数量：%d\n实上链消息数量：%d \nC2消息堵住了！！快看看机器有没有什么问题", addr, addrOwn, str, Threshold, diff)
// 		if utils.Initconf.String("Local") == "local" {
// 			notify.SendQyMessage(5, []byte(mss), "", 0)
// 		} else {
// 			notify.SendQyMessage(4, []byte(mss), admin, 0)
// 		}
// 	}

// }

// func (p *InterfaceParam) OpMessageP2NotSet(v string) {
// 	KeyType := strings.Split(v, "_")
// 	if len(KeyType) < 3 {
// 		utils.Log.Errorln("未知类型", v)
// 		return
// 	}
// 	addr := KeyType[2]
// 	//需要告警
// 	un := time.Now().Unix()
// 	un5mintuAgo := un - 9000
// 	un5mintuAgeTime := time.Unix(un5mintuAgo, 0).Format(utils.TimeString)
// 	unStr := time.Unix(un, 0).Format(utils.TimeString)
// 	str := fmt.Sprintf("%s------%s", strings.ReplaceAll(un5mintuAgeTime, "-", "."), strings.ReplaceAll(unStr, "-", "."))
// 	p.Notify(addr, str, "P2")

// }

// func (p *InterfaceParam) OpMessageC2Set(v string) {
// 	KeyType := strings.Split(v, "_")
// 	if len(KeyType) < 3 {
// 		utils.Log.Errorln("未知类型", v)
// 		return
// 	}
// 	addr := KeyType[2]
// 	//需要告警
// 	t := KeyType[3]
// 	ago, err := strconv.Atoi(t)
// 	if err != nil {
// 		utils.Log.Errorln(err)
// 		return
// 	}
// 	un := time.Now().Unix()
// 	un5mintuAgo := un - int64(ago)
// 	un5mintuAgeTime := time.Unix(un5mintuAgo, 0).Format(utils.TimeString)
// 	unStr := time.Unix(un, 0).Format(utils.TimeString)
// 	str := fmt.Sprintf("%s------%s", strings.ReplaceAll(un5mintuAgeTime, "-", "."), strings.ReplaceAll(unStr, "-", "."))
// 	p.Notify(addr, str, "C2")

// }

// func (p *InterfaceParam) Notify(addr, str, c2orp2 string) {
// 	admin := utils.Getadmin(addr)
// 	addrOwn := utils.GetAddrOwner(addr)
// 	mss := fmt.Sprintf("%s %s \n%s  \n%s消息堵住了！！快看看机器有没有什么问题", addr, addrOwn, str, c2orp2)
// 	if utils.Initconf.String("Local") == "local" {
// 		notify.SendQyMessage(5, []byte(mss), "", 0)
// 	} else {
// 		notify.SendQyMessage(4, []byte(mss), admin, 0)
// 	}
// }

// //###########################################################单台效率告警 统计算法入口 ##########################################################
// func (p *InterfaceParam) SignEfficient(addr, cid string, method int, ggregate []byte) {
// 	//
// 	confKey := addr + "Machine"
// 	count, err := utils.Initconf.Int64(confKey)
// 	if err != nil {
// 		//utils.Log.Traceln(err, "---------", addr)
// 		return
// 	}
// 	if count == 0 {
// 		utils.Log.Tracef("%s 的C2上链情况不再监控节点内。。", addr)
// 		return
// 	}

// 	param := models.AboutMachineEfficient{}
// 	if method == 26 { //C2 集合
// 		p.ProveCommitC2Aggregate = new(miner.ProveCommitAggregateParams)
// 		if err := json.Unmarshal(ggregate, p.ProveCommitC2Aggregate); err != nil {
// 			utils.Log.Errorln(err)
// 			return
// 		}
// 		count, err := p.ProveCommitC2Aggregate.SectorNumbers.Count()
// 		if err != nil {
// 			utils.Log.Errorln(err)
// 			return
// 		}
// 		param.SectorCount = int(count)

// 	}
// 	param.Timestamp = time.Now().Unix()
// 	param.Cid = cid
// 	param.Addr = addr
// 	param.Method = method
// 	// if err := models.AboutMachineEfficientUpsert(param); err != nil {
// 	// 	utils.Log.Errorln(err)
// 	// }
// 	sqlParam := new(models.Efficient)
// 	sqlParam.Addr = param.Addr
// 	sqlParam.Cid = param.Cid
// 	sqlParam.Created = int(param.Timestamp)
// 	sqlParam.Method = param.Method
// 	sqlParam.SectorCount = param.SectorCount
// 	if err = new(models.Efficient).Insert(*sqlParam); err != nil {
// 		utils.Log.Errorln(err)
// 	}
// }

// func (i *InterfaceParam) public64NoSet(addr string, count int64, param models.AboutMachineEfficient) {
// 	//mcount, _ := models.GetAboutMachineEfficientCount(param)
// 	sqlParam := new(models.Efficient)
// 	sqlParam.Addr = param.Addr
// 	sqlParam.Created = int(param.Timestamp)
// 	sqlParam.Method = param.Method
// 	mcount := sqlParam.FindCount(*sqlParam)
// 	key := fmt.Sprintf(utils.Efficient64NoSetPrevious, addr, param.Method)
// 	temp := float32(mcount) / 4.0 / float32(count)
// 	preTempStr := redisdb.GetSetDB1String(key, fmt.Sprintf("%.3f", temp))
// 	preTempFloat32, err := strconv.ParseFloat(preTempStr, 32)
// 	if err != nil {
// 		utils.Log.Errorln(err)
// 		return
// 	}
// 	op := ""
// 	if param.Method == 6 {
// 		op = "P2"
// 	}
// 	if param.Method == 7 {
// 		op = "C2"
// 	}
// 	t := time.Now()
// 	befor := t.Unix() - 6*3600
// 	bs := time.Unix(befor, 0).Format(utils.TimeString)
// 	ts := fmt.Sprintf("%s------%s", strings.ReplaceAll(bs, "-", "."), strings.ReplaceAll(t.Format(utils.TimeString), "-", "."))
// 	owenaddr := utils.GetAddrOwner(addr)
// 	if temp < 3.5 {
// 		//告警

// 		mss := fmt.Sprintf("%s %s （非集合消息）\n%s\n当天预测：%.2f    单台效率：%.2f \n\n当前%s单台效率低于3.5！！！快看看机器有没有什么问题", addr, owenaddr, ts, float32(mcount/4.0), temp, op)
// 		i.Less3to5NotifyNoSet(addr, mss)
// 		return
// 	}
// 	sub := float32(preTempFloat32) - temp
// 	if 0.1 < sub {
// 		//告警
// 		mss := fmt.Sprintf(" %s %s （非集合消息）\n上一次预测\n当天预测：%.2f         单台效率：%.2f\n\n%s\n当天预测：%.2f         单"+
// 			"台效率：%.2f\n\n当前%s单台效率降速大于0.1！！！快看看机器有没有什么问题", addr, owenaddr, float32(mcount/4.0), float32(preTempFloat32), ts, float32(mcount/4.0), temp, op)
// 		i.LessZeroTo1NotifyNoSet(addr, mss)
// 	}
// 	utils.Log.Tracef("----------------------单台效率告警两小时跑一次----------------------%s %s", param.Addr, "一切正常")

// }
// func (i *InterfaceParam) public32NoSet(addr string, count int64, param models.AboutMachineEfficient) {

// 	//mcount, _ := models.GetAboutMachineEfficientCount(param)
// 	sqlParam := new(models.Efficient)
// 	sqlParam.Addr = param.Addr
// 	sqlParam.Created = int(param.Timestamp)
// 	sqlParam.Method = param.Method
// 	mcount := sqlParam.FindCount(*sqlParam)
// 	key := fmt.Sprintf(utils.Efficient32NoSetPrevious, addr, param.Method)
// 	temp := float32(mcount) / 8.0 / float32(count)
// 	preTempStr := redisdb.GetSetDB1String(key, fmt.Sprintf("%.3f", temp))
// 	preTempFloat32, err := strconv.ParseFloat(preTempStr, 32)
// 	if err != nil {
// 		utils.Log.Errorln(err)
// 		return
// 	}
// 	owenaddr := utils.GetAddrOwner(addr)
// 	op := ""
// 	if param.Method == 6 {
// 		op = "P2"
// 	}
// 	if param.Method == 7 {
// 		op = "C2"
// 	}
// 	t := time.Now()
// 	befor := t.Unix() - 6*3600
// 	bs := time.Unix(befor, 0).Format(utils.TimeString)
// 	ts := fmt.Sprintf("%s------%s", strings.ReplaceAll(bs, "-", "."), strings.ReplaceAll(t.Format(utils.TimeString), "-", "."))
// 	if temp < 5.5 {
// 		//告警
// 		mss := fmt.Sprintf("%s %s （非集合消息）\n%s\n当天预测：%.2f    单台效率：%.2f \n\n当前%s单台效率低于5.5！！！快看看机器有没有什么问题", addr, owenaddr, ts, float32(mcount/8.0), temp, op)
// 		i.Less3to5NotifyNoSet(addr, mss)
// 		return
// 	}
// 	sub := float32(preTempFloat32) - temp
// 	if 0.1 < sub {
// 		//告警
// 		mss := fmt.Sprintf(" %s %s （非集合消息）\n上一次预测\n当天预测：%.2f         单台效率：%.2f\n\n%s\n当天预测：%.2f         单"+
// 			"台效率：%.2f\n\n当前%s单台效率降速大于0.1！！！快看看机器有没有什么问题", addr, owenaddr, float32(mcount/8.0), float32(preTempFloat32), ts, float32(mcount/8.0), temp, op)
// 		i.LessZeroTo1NotifyNoSet(addr, mss)
// 	}
// 	utils.Log.Tracef("----------------------单台效率告警两小时跑一次----------------------%s %s", param.Addr, "一切正常")
// }

// //单台效率告警定时任务入口
// func CronEfficient() {
// 	p := new(InterfaceParam)
// 	Unix := time.Now().Unix()
// 	for _, v := range utils.Array6 {
// 		confKey := v + "Machine"
// 		count, err := utils.Initconf.Int64(confKey)
// 		if err != nil {
// 			utils.Log.Traceln(err, "---------", v)
// 			continue
// 		}
// 		param := models.AboutMachineEfficient{}
// 		param.Timestamp = Unix
// 		param.Method = 6 //P2
// 		param.Addr = v
// 		utils.Log.Tracef("----------------------单台效率告警两小时跑一次  P2----------------------%s", param.Addr)
// 		Size32Or64 := p.Is32or64(v)
// 		if Size32Or64 == 32 {
// 			p.public32NoSet(v, count, param)
// 		}
// 		if Size32Or64 == 64 {
// 			p.public64NoSet(v, count, param)
// 		}

// 	}

// 	for _, v := range utils.Array7 { //C2

// 		confKey := v + "Machine"
// 		count, err := utils.Initconf.Int64(confKey)
// 		if err != nil {
// 			utils.Log.Traceln(err, "---------", v)
// 			continue
// 		}
// 		param := models.AboutMachineEfficient{}
// 		param.Timestamp = Unix
// 		param.Addr = v
// 		param.Method = 7 //C2
// 		utils.Log.Tracef("----------------------单台效率告警两小时跑一次 C2----------------------%s", param.Addr)
// 		Size32Or64 := p.Is32or64(v)
// 		if Size32Or64 == 32 {
// 			p.public32NoSet(v, count, param)
// 		}
// 		if Size32Or64 == 64 {
// 			p.public64NoSet(v, count, param)
// 		}
// 	}
// 	for _, v := range utils.Array25 {
// 		confKey := v + "Machine"
// 		count, err := utils.Initconf.Int64(confKey)
// 		if err != nil {
// 			utils.Log.Traceln(err, "---------", v)
// 			continue
// 		}
// 		utils.Log.Traceln("单台效率告警两小时跑一次 25", v, count)
// 	}
// 	for _, v := range utils.Array26 {
// 		confKey := v + "Machine"
// 		count, err := utils.Initconf.Int64(confKey)
// 		if err != nil {
// 			utils.Log.Traceln(err, "---------", v)
// 			continue
// 		}
// 		// param := models.AboutMachineEfficient{}
// 		// param.Timestamp = Unix
// 		// param.Method = 26
// 		// param.Addr = v
// 		//utils.Log.Tracef("----------------------单台效率告警两小时跑一次   ---------------------- %s", param.Addr)
// 		//_, AggregateCount := models.GetAboutMachineEfficientCount(param)
// 		sqlParam := new(models.Efficient)
// 		sqlParam.Addr = v
// 		sqlParam.Created = int(Unix)
// 		sqlParam.Method = 26
// 		AggregateCount := sqlParam.FindCountAggreagte(*sqlParam)
// 		//C2
// 		key := fmt.Sprintf(utils.Efficient64SetPrevious, v, sqlParam.Method)
// 		temp := float32(AggregateCount) / 4.0 / float32(count)
// 		preTempStr := redisdb.GetSetDB1String(key, fmt.Sprintf("%.3f", temp))
// 		preTempFloat32, err := strconv.ParseFloat(preTempStr, 32)
// 		if err != nil {
// 			utils.Log.Errorln(err)
// 			continue
// 		}
// 		op := ""
// 		if sqlParam.Method == 26 {
// 			op = "C2"
// 		}
// 		t := time.Now()
// 		befor := t.Unix() - 6*3600
// 		bs := time.Unix(befor, 0).Format(utils.TimeString)
// 		ts := fmt.Sprintf("%s------%s", strings.ReplaceAll(bs, "-", "."), strings.ReplaceAll(t.Format(utils.TimeString), "-", "."))
// 		owenaddr := utils.GetAddrOwner(sqlParam.Addr)
// 		if temp < 3.5 {
// 			//告警
// 			mss := fmt.Sprintf("%s %s （集合消息）\n%s\n当天预测：%.2f    单台效率：%.2f \n\n当前%s单台效率低于3.5！！！快看看机器有没有什么问题", sqlParam.Addr, owenaddr, ts, float32(AggregateCount/4.0), temp, op)
// 			p.Less3to5NotifyNoSet(sqlParam.Addr, mss)
// 			continue
// 		}
// 		sub := float32(preTempFloat32) - temp
// 		if 0.1 < sub {
// 			//告警
// 			mss := fmt.Sprintf(" %s %s （集合消息）\n上一次预测\n当天预测：%.2f         单台效率：%.2f\n\n%s\n当天预测：%.2f         单台效率：%.2f\n\n当前%s"+
// 				"单台效率降速大于0.1！！！快看看机器有没有什么问题", sqlParam.Addr, owenaddr, float32(AggregateCount/4.0), float32(preTempFloat32), ts, float32(AggregateCount/4.0), temp, op)
// 			p.LessZeroTo1NotifyNoSet(sqlParam.Addr, mss)
// 			continue
// 		}
// 	}

// }

// func (i *InterfaceParam) Less3to5NotifyNoSet(addr, mss string) {
// 	admin := utils.Getadmin(addr)
// 	if utils.Initconf.String("Local") == "local" {
// 		notify.SendQyMessage(5, []byte(mss), "", 0)
// 	} else {
// 		notify.SendQyMessage(6, []byte(mss), admin, 0)
// 	}
// }

// func (i *InterfaceParam) LessZeroTo1NotifyNoSet(addr, mss string) {
// 	admin := utils.Getadmin(addr)

// 	if utils.Initconf.String("Local") == "local" {
// 		notify.SendQyMessage(5, []byte(mss), "", 0)
// 	} else {
// 		notify.SendQyMessage(6, []byte(mss), admin, 0)
// 	}
// }

// //每天清除一次24小时前的数据
// func DrupDataForEfficient() {
// 	end := time.Now().Unix() - 24*3600
// 	if _, err := utils.DB.Table("efficient").Where("created<?", end).Delete(&models.Efficient{}); err != nil {
// 		utils.Log.Errorln(err)
// 		return
// 	}

// }
