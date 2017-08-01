package goflow

import (
	"strings"
	"time"
)

//抄送实例表
type CCOrder struct {
	Id         string      `xorm:"varchar(36) pk notnull"` //主键ID
	OrderId    string      `xorm:"varchar(36) index"`      //流程实例ID
	ActorId    string      `xorm:"varchar(36)"`            //操作者ID
	Creator    string      `xorm:"varchar(36)"`            //流程实例创建者ID
	CreateTime time.Time   `xorm:"datetime"`               //流程实例创建时间
	FinishTime time.Time   `xorm:"datetime"`               //流程实例完成时间
	State      FLOW_STATUS `xorm:"tinyint"`                //流程实例状态
}

func GetCCOrder(orderId string, actorIds ...string) []*CCOrder {
	ccorders := make([]*CCOrder, 0)
	err := orm.Where(`"OrderId" = ? and "ActorId" in (?)`, orderId, strings.Join(actorIds, ",")).Find(&ccorders)
	PanicIf(err, "fail to GetCCOrder")
	return ccorders
}
