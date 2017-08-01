package goflow

import "fmt"

//XML分叉节点
type ForkModel struct {
	NodeModel
}

//执行
func (p *ForkModel) Exec(execution *Execution) {
	fmt.Println("ForkModel 执行")
	p.RunOutTransition(execution)
}
