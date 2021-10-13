package module

import (
	"encoding/json"
	"lotus_data_sync/utils"
	"github.com/globalsign/mgo"
	"gopkg.in/mgo.v2/bson"
)

// Include  StateListActors interface data（Miner） 、produce Msg wallet
type Tag struct {
	Address string `bson:"address" json:"address"`
	Name    string `bson:"name" json:"name"`
	Signed  bool   `bson:"signed" json:"signed"`
	//todo
	GmtCreate   int64 `bson:"gmt_create" json:"gmt_create"`
	GmtModified int64 `bson:"gmt_modified" json:"gmt_modified"`
}
type TagRes struct {
	Address string `bson:"address" json:"address"`
	Name    string `bson:"name" json:"name"`
	Signed  bool   `bson:"signed" json:"signed"`
	//todo
	GmtCreate   int64 `bson:"gmt_create" json:"gmt_create"`
	GmtModified int64 `bson:"gmt_modified" json:"gmt_modified"`
}

//	"height": 611739,
//	"timestamp": 1616658570,
//	"totalCount": 1683,
//	"miners": [{
//		"address": "f0127595",
//		"tag": {
//			"name": "时空云",
//			"signed": true
//		}

type TagResp struct {
	Height     string   `json:"height"`
	Timestamp  string   ` json:"timestamp"`
	TotalCount int      ` json:"totalCount"`
	Miners     []*miner ` json:"miners"`
}
type miner struct {
	Address string  `json:"address"`
	FilTag  *FilTag ` json:"tag"`
}
type FilTag struct {
	Name   string ` json:"name"`
	Signed bool   `json:"signed"`
}

type FilscanTagResult struct {
	Address string `bson:"address" json:"address"`
	Name    string `bson:"name" json:"name"`
}

const (
	TagCollection = "tag"
)

func CreateTagIndex() {
	ms, c := Connect(TagCollection)
	defer ms.Close()
	ms.SetMode(mgo.Monotonic, true)

	indexs := []mgo.Index{
		{Key: []string{"address"}, Unique: true, Background: true},
	}
	for _, index := range indexs {
		if err := c.EnsureIndex(index); err != nil {
			panic(err)
		}
	}
}

func GetTagByAddress(address string) (res FilscanTagResult, err error) {
	q := bson.M{"address": address}
	//utils.Log.Traceln(q)
	err = FindOne(TagCollection, q, nil, &res)
	return
}
func InsertTag(tag *Tag) (err error) {
	tag.GmtCreate = TimeNow
	tag.GmtModified = TimeNow
	tbyte, _ := json.Marshal(tag)
	var p interface{}
	err = json.Unmarshal(tbyte, &p)
	if err != nil {
		return err
	}
	err = Insert(TagCollection, p)
	return
}

func UpdateTag(tag *TagRes) (err error) {
	tag.GmtCreate = tag.GmtCreate
	tag.GmtModified = TimeNow
	tbyte, _ := json.Marshal(tag)
	var p interface{}
	err = json.Unmarshal(tbyte, &p)
	if err != nil {
		return err
	}
	selector := bson.M{"address": tag.Address}
	//	utils.Log.Traceln(selector)
	_, err = Upsert(TagCollection, selector, p) //== update
	return
}
func UpdateTagGmtModifiedByAddress(address string) error {
	q := bson.M{"addresds": address}
	u := bson.M{"$set": bson.M{"gmt_modified": TimeNow}}
	utils.Log.Traceln(q, u)
	return Update(TagCollection, q, u)
}

func InsertTagMulti(tags []Tag) (err error) {
	if len(tags) == 0 {
		return
	}
	var docs []interface{}
	for _, value := range tags {
		value.GmtModified = TimeNow
		value.GmtCreate = TimeNow
		docs = append(docs, value)
	}
	err = Insert(TagCollection, docs...)
	return err
}

func UpsertTagArr(Tags []*Tag) (err error) {
	if len(Tags) == 0 {
		return
	}
	for _, Tag := range Tags {
		q := bson.M{"address": Tag.Address}
		var res []*TagRes
		//	utils.Log.Traceln(q)
		err = FindAll(TagCollection, q, nil, &res)
		if err != nil {
			return
		}
		if len(res) < 1 {
			err = InsertTag(Tag)
			if err != nil {
				return
			}
		} else {
			if Tag.Address == "" {
				continue
			}
			res[0].Address = Tag.Address
			res[0].Name = Tag.Name
			res[0].Signed = Tag.Signed
			err = UpdateTag(res[0])
			if err != nil {
				return
			}
		}
	}
	return
}

func GetDistinctFromAddressByTimeInTag(startTime, endTime int64) (res []string, err error) {
	q := bson.M{"gmt_create": bson.M{"$gte": startTime, "$lte": endTime}}
	err = Distinct(TagCollection, "address", q, &res)
	return
}
