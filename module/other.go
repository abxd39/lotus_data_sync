package module

//"payment_32":0.3138,"payment_1T":10.0416,"payment_1P":10282.5984,"preGas_32":0.03444047237278913,"preGas_1T":1.1020951159292522,"preGas_1P":1128.5453987115543,"total_32":0.34824047237278916,"total_1T":11.143695115929253,"total_1P":11411.143798711555,"sector_type":"32G","timestamp":1617250701097,"latest_height":631477,
type S32 struct {
	Payment32    float64 `json:"payment_32"`
	Payment1t    float64 `json:"payment_1T"`
	Payment1p    float64 `json:"payment_1P"`
	PreGas32     float64 `json:"preGas_32"`
	PreGas1t     float64 `json:"preGas_1T"`
	PreGas1p     float64 `json:"preGas_1P"`
	Total32      float64 `json:"total_32"`
	Total1t      float64 `json:"total_1T"`
	Total1p      float64 `json:"total_1P"`
	Timestamp    uint64  `json:"timestamp"`
	LatestHeight uint64  `json:"latest_height"`
}
type S64 struct {
	Data struct {
		GasIn64GB float64 `json:"gasIn64GB"`
		Total1t64 float64 `json:"newlyPowerCostIn64GB"`
	} `json:"data"`
}

type Info struct {
	Code   string `json:"code"`
	Result struct {
		NewlyPrice       float64 `json:"newlyPrice"`
		TotalFil         uint64  `json:"totalFil"`
		CurrentFil       uint64  `json:"currentFil"`
		TotalBurnUp      uint64  `json:"totalBurnUp"`
		FlowRate         float64 `json:"flowRate"`
		NewlyFilIn24h    float64 `json:"newlyFilIn24h"`
		PowerIn24H       string  `json:"PowerIn24H"`
		AvgBlocksTipSet  float64 `json:"avgBlocksTipSet"`
		AvgGasPremium    float64 `json:"avgGasPremium"`
		BlockRewardIn24h float64 `json:"blockRewardIn24h"`
		FlowTotal        string  `json:"flowTotal"`
		OneDayMessages   int64   `json:"oneDayMessages"`
		TransferIn24H    int64   `json:"transferIn24H"`
		TotalAccounts    uint64  `json:"totalAccounts"`
		PledgeCollateral float64 `json:"pledgeCollateral"`
	} `json:"data"`
}

type GetPowerInR struct {
	C      string `json:"code"`
	Result struct {
		Datetime []string `json:"datetime"`
		Powers   []string `json:"powers"`
	} `json:"data"`
	Unit string `json:"unit"`
}

// type GasRsp struct {
// 	C      string `json:"code"`
// 	Result struct {
// 		BaseFee  *tfspt.BaseFee  `json:"baseFee"`
// 		GasFee64 *tfspt.GasFee64 `json:"gasFee64"`
// 		GasFee32 *tfspt.GasFee32 `json:"gasFee32"`

// 		Height   []int64  `json:"height"`
// 		TimeList []string `json:"timeList"`
// 	} `json:"data"`
// }

// type MarketPairsRes struct {
// 	Data struct {
// 		Name        string              `json:"name"`
// 		MarketPairs []*tfspt.MarketPair `json:"marketPairs"`
// 	} `json:"data"`
// 	Rate float64
// }

// type DaedLinesResp struct {
// 	Data struct {
// 		MinerID   string
// 		DeadLines []*tfspt.DeadLine `json:"dead_lines"`
// 	} `json:"data"`
// }

type DeadLine struct {
	DeadLineId       int64  `json:"DeadLineId"`
	Partitions       int64  `json:"Partitions"`
	Sectors          uint64 `json:"Sectors"`
	Fault            uint64 `json:"Fault"`
	Recovery         uint64 `json:"Recovery"`
	ProvenPartitions string `json:"ProvenPartitions"`
	Current          string `json:"Current"`
}
type DeadInfo struct {
	DeadlineInde        uint64 `json:"deadlineinde"`
	DeadlineOpen        string `protobuf:"bytes,2,opt,name=DeadlineOpen,proto3" json:"DeadlineOpen,omitempty"`
	DeadlineClose       string `protobuf:"bytes,3,opt,name=DeadlineClose,proto3" json:"DeadlineClose,omitempty"`
	DeadlineChallenge   string `protobuf:"bytes,4,opt,name=DeadlineChallenge,proto3" json:"DeadlineChallenge,omitempty"`
	DeadlineFaultCutoff string `protobuf:"bytes,5,opt,name=DeadlineFaultCutoff,proto3" json:"DeadlineFaultCutoff,omitempty"`
}
