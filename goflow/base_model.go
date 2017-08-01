package goflow

import "fmt"

//XML节点基本模型
type BaseModel struct {
	Name        string `xml:"name,attr"`        //节点名称
	DisplayName string `xml:"displayName,attr"` //节点显示名称
}

func (p *BaseModel) GetName() string {
	fmt.Println("BaseModel 获取名字")
	return p.Name
}
