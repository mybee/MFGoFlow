package goflow

import "fmt"

//XML开始节点元素
type StartModel struct {
	NodeModel
}

//执行
func (p *StartModel) Exec(execution *Execution) {
	fmt.Println("StartModel 执行")
	p.RunOutTransition(execution)
}
