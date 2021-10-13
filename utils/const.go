package utils

import "time"

var TipMark bool

const (
	BlockTimeTag = "d69a07f4-ba5f-4bb5-aa12-94e2432a38ed"
)

const PrecisionDefault = 8
const MonitorTime = 15 //C2 P2 告警时间  单位为分钟

const (
	TimeString = "2006-01-02 15:04:05"
	TimeDate   = "2006-01-02"
)

const (
	RedisKeyAccessToken = "access_token:1000004" //并发安全
)

const Tokenfail = "access_token"

const ChatName = "问题反馈"

const (
	//错误扇区消息监听事件
	RedisKeyDeclareFaultsRecovered = "DeclareFaultsRecovered_%s_%d"
	//有消息上链 然后需要告警时空抽查上链
	RedisKeyWindowPostMark   = "Declare_wind_post:%s_%d"
	RedisKeySectorErr        = "SubmitWindowedPoSt_%s_%d"
	RedisOnceADaySectorErr   = "SectorOnceErr_%s_%s"
	RedisSectorErr20s        = "SectorErr20s_%s_%s"
	RedisSectorErrFaultsTime = "SectorErrTime_%s_%s_%d"
	RedisLotusSyncWait       = "lotus_sync_once"
	RedisPhone               = "callee_%s_11111"
)

//消息计数
const (
	C2_64NoSet       = "C2_64noSet_%s"       //5分钟内的消息数量
	C2_64NoSetDuring = "C2_64noSetDuring_%s" //5分钟内上链的消息数
	C2_64NoSetTotal  = "C2_64noSetTotal_%s"  //单位时间内的消息总数
	//检测时间点事件key
	C2_64Expire  = "C2_64Expire_%s"  //过期时间
	P2_PreExpire = "P2_PreExpire_%s" //设置P2上链的时间区间

	C2SETExpire = "C2SET_Expire_%s_%d"
)

//单台效率
const (
	//非集合64
	Efficient64NoSetPrevious = "Efficient_64NoSetP2Previous_%s_%d"
	//非集合32
	Efficient32NoSetPrevious = "Efficient_32NoSetC2Previous_%s_%d"
	//集合64C2
	Efficient64SetPrevious = "Efficient_64SetC2Previous_%s_%d"

)

const (
	Balance48 = "balance48_%s"
	Balance24 = "balance24_%s"
)

var LocalMinerAddr = map[string]string{
	"f010936":  "测试1", //测试网络
	"f03112":   "测试2", //测试网络
	"f06806":   "测试3",
	"f012400":  "测试4",
	"f01247":   "测试5",
	"f01248":   "测试6",
	"f03149":   "测试7",
	"f012809":  "测试8",
	"f014533":  "测试9",
	"f0154294": "测试错误扇区恢复",
}

var MinerAddr = map[string]string{
	"f0469055":  "西部1号",
	"f0428661":  "西部2号",
	"f0716775":  "西部3号",
	"f0717289":  "西部4号",
	"f0119976":  "孙总1号中商云",
	"f0494733":  "孙总2号智力云CH1",
	"f0686816":  "孙总3号",
	"f0120057":  "西部世界005",
	"f0113735":  "算力果1号",
	"f01218989": "算力果2号",
	//"f0730670":  "",//这两个节点说是交付出去了，不用监控了
	//"f0806904":  "",
	"f0748179":  "今泉",
	"f0822441":  "孙总4号",
	"f0822818":  "孙总5号",
	"f0699021":  "沈总郑总",
	"f01021773": "ZKYK&龙池",
	"f01038389": "牧牛",
	"f01177590": "余总",
	"f01089422": "测试号",
	"f01136428": "西部5号三体科技",
	"f01179295": "西部券商", //西部券商
	"f01215328": "罗总",   // 罗总
	"f01209020": "张总",   // 张总
	"f01227383": "迪拜",   // 迪拜
	"f01264518": "西部华芯",
	"f01224142": "天一星际",
	"f01273431": "长沙",
	//"f0123261":"其他",
}

func GetAddrOwner(key string) string {
	if Initconf.String("Local") == "local" {
		//if v, ok := LocalMinerAddr[key]; ok {
		if v, ok := MinerAddr[key]; ok {
			return v
		}
		return "测试网络节点"
	} else {
		if v, ok := MinerAddr[key]; ok {
			return v
		}
		return ""
	}
}

const MinerStr = `f0469055 西部1号
f0428661 西部2号
f0716775 西部3号
f0717289 西部4号
f0119976 孙总1号中商云
f0494733 孙总2号智力云CH1
f0686816 孙总3号
f0120057 西部世界005
f0113735 算力果
f0748179 今泉
f0822441 孙总4号
f0822818 孙总5号
f0699021 沈总郑总
f01021773 ZKYK&龙池
f01038389 牧牛
f01177590 余总
f01089422 测试号
f01136428 西部5号三体科技
f01215328 罗总
f01218989 算力果
f01209020 张总
f01227383 迪拜
f01179295 西部券商`

const Notify0_5 = `"@all"`

const Notify5_15 = `"changyuankui","mingxiaokai","chengzh","haer","wulingeng"`

const Notify15_20 = `"changyuankui"`

const PhoneList = "15920038315,13538276476,17301986143,18565162701,13143240516,15622956746"

var MapMember = map[string]string{
	//"王应文": "15920038315",
	//"红莲":  "17301986143",
	//"王大发": "13058001431",
	"常元魁": "13652351702",
	"程振华": "15622956746",
	"郑立凯": "18565162701",
	"吴战飞": "13714330963",
	"蒋老师": "18813952608",
	"联陈发": "15814784737",
	"杜舸":  "13534031345",
	//"钟佳林": "17602314504",
	//"吕杨开": "13296545968",
	"柴仁美": "15361428846",
	"唐伟钦": "18574382151",
	"张虎军": "13143240516",
	//"唐景":  "13510220779",
	"林湘桃":"18038129151",
}

func StringToTime(t string) time.Time {
	the_time, err := time.ParseInLocation("2006-01-02 15:04:05", t, time.Local)
	if err != nil {
		Log.Errorln(err)
		return time.Now()
	}
	return the_time
}

/*
f0119976 孙总1号中商云 32GiB
f0748179 今泉 64GiB
f0822441  孙总4号 64GiB
f0822818 孙总5号 64GiB
f01038389 牧牛 64GiB
f01177590 余总 64GiB
f01136428  西部5号三体科技 64GiB
f01215328  罗总 32GiB
f01218989  算力果 32GiB
f01209020  张总 64GiB
f01227383  迪拜 32GiB
f01179295  西部券商 64GiB
*/
var Map32Or64 = map[string]int64{
	"f0119976":  32,
	"f0748179":  64,
	"f0822441":  64,
	"f0822818":  64,
	"f01038389": 64,
	"f01177590": 64,
	//"f01136428": 64,
	"f01215328": 32,
	"f01218989": 32,
	"f01209020": 64,
	"f01227383": 32,
	"f01179295": 64,
	"f01264518": 64,
}

//单台效率告警 新增流程 第一步 加入下列 数组，和上面的map 然后加配置文件 设备数量
var Array6 = []string{"f01227383", "f01179295", "f01264518", "f01038389"} //6
var Array7 = []string{"f01227383", "f01179295"}                           //7
var Array25 = []string{}                                                  //25
var Array26 = []string{"f01038389", "f01264518"}                          //26

//跑集合的节点
var MapSet = map[string]bool{
	"f0822441":  true,
	"f01038389": true,
	"f01264518": true,
}

func Getadmin(addr string) string {
	switch addr {
	case "f0136428", "f01177590", "f0822818", "f01021773", "f01215328", "f0716775":
		return `"550180641@qq.com"` //杜舸
	case "f0822441", "f0469055", "f01038389", "f01224142":
		return `"ZHJ9702@163.com"` //张虎军
	case "f0494733", "f01218989", "f0717289":
		return `"2596866287@qq.com"` //练称发
	case "f01179295":
		return `"1781592191@qq.com"` //唐伟钦
	case "f01264518":
		return `"329785167@qq.com"` //柴仁美
	case "f01227383":
		return `"chengzh@zhiannet.com","wzf@staron.io"`
	default:
		return Notify15_20

	}
}
