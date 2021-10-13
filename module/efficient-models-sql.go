package module

import (
	"lotus_data_sync/utils"
)

type Efficient struct {
	Id          int    `xorm:"not null pk autoincr unique INT(11)"`
	Cid         string `xorm:"not null unique VARCHAR(100)"`
	Addr        string `xorm:"VARCHAR(45)"`
	Method      int    `xorm:"INT(11)"`
	SectorCount int    `xorm:"comment('聚合的扇区数量') INT(11)"`
	Created     int    `xorm:"INT(11)"`
}

func (e *Efficient) TableName() string {
	return "efficient"
}
func (e *Efficient) Insert(param Efficient) error {
	if _, err := utils.DB.InsertOne(&param); err != nil {
		utils.Log.Errorln(err)
		return err
	}
	return nil
}

func (e *Efficient) FindCount(param Efficient) int64 {
	query := utils.DB.Table("efficient")
	if param.Addr != "" {
		query.Where("addr=?", param.Addr)
	}
	if param.Method != 0 {
		query.Where("method=?", param.Method)
	}
	if param.Created != 0 {
		query.Where("created<=? and created>?", param.Created, param.Created-6*3600)
	}
	if count, err := query.Count(); err != nil {
		utils.Log.Errorln(err)
		return 0
	} else {
		return count
	}
}
func (e *Efficient) FindCountAggreagte(param Efficient) int64 {
	query := utils.DB.Table("efficient")
	if param.Addr != "" {
		query.Where("addr=?", param.Addr)
	}
	if param.Method != 0 {
		query.Where("method=?", param.Method)
	}
	if param.Created != 0 {
		query.Where("created<=? and created>?", param.Created, param.Created-6*3600)
	}
	if count, err := query.SumInt(&Efficient{}, "sector_count"); err != nil {
		utils.Log.Errorln(err)
		return 0
	} else {
		return count
	}
}
