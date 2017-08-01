package goflow

import (
	"time"
	"fmt"
)

//流程定义实体类
type Process struct {
	Id             string        `xorm:"varchar(36) pk notnull"` //主键ID
	Version        int           `xorm:"tinyint"`                //版本
	Name           string        `xorm:"varchar(100) index"`     //流程定义名称
	DisplayName    string        `xorm:"varchar(200)"`           //流程定义显示名称
	InstanceAction string        `xorm:"varchar(200)"`           //当前流程的实例Action,(Web为URL,一般为流程第一步的URL;APP需要自定义),该字段可以直接打开流程申请的表单
	State          FLOW_STATUS   `xorm:"tinyint"`                //状态
	CreateTime     time.Time     `xorm:"datetime"`               //创建时间
	Creator        string        `xorm:"varchar(36)"`            //创建人
	Content        string        `xorm:"text"`                   //流程定义XML
	Model          *ProcessModel `xorm:"-"`                      //Model对象
}

//根据ID得到Process
func (p *Process) GetProcessById(id string) bool {
	fmt.Println("根据ID得到Process")
	p.Id = id
	success, err := orm.Get(p)
	PanicIf(err, "fail to GetProcessById")
	return success
}

//根据Process本身条件得到Process
func (p *Process) GetProcess() bool {
	fmt.Println("根据Process本身条件得到Process")
	success, err := orm.Get(p)
	PanicIf(err, "fail to GetProcess")
	return success
}

//设定Model对象
func (p *Process) SetModel(model *ProcessModel) {
	fmt.Println("设定Model对象")
	p.Model = model
	p.Name = model.Name
	p.DisplayName = model.DisplayName
	p.InstanceAction = model.InstanceAction
}

//得到最新的Process
func GetLatestProcess(name string) *Process {
	fmt.Println("得到最新的Process")
	process := &Process{
		Name: name,
	}
	processes := make([]*Process, 0)
	err := orm.Desc("Version").Find(&processes, process)
	PanicIf(err, "fail to GetLatestProcess")
	if len(processes) > 0 {
		return processes[0]
	} else {
		return nil
	}
}
