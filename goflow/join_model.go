package goflow

import "fmt"

//XML合并节点
type JoinModel struct {
	NodeModel
}

//合并分叉节点
func (p *JoinModel) MergeBranchHandle(execution *Execution) {
	fmt.Println("合并分叉节点")
	activeNodes := FindActiveNodes(p)
	MergeHandle(execution, activeNodes)
}

//执行
func (p *JoinModel) Exec(execution *Execution) {
	fmt.Println("JoinModel 执行")
	p.MergeBranchHandle(execution)
	if execution.IsMerged {
		p.RunOutTransition(execution)
	}
}

//递归查找分叉节点
func FindForkTaskNames(node INodeModel) []string {
	fmt.Println("递归查找分叉节点")
	ret := make([]string, 0)
	switch node.(type) {
	case *ForkModel:
	default:
		for _, tm := range node.GetInputs() {
			switch tm.Source.(type) {
			case *SubProcessModel:
				ret = append(ret, tm.Source.(*SubProcessModel).Name)
			case *TaskModel:
				ret = append(ret, tm.Source.(*TaskModel).Name)
			default:
				ret = append(ret, FindForkTaskNames(tm.Source)...)
			}
		}
	}
	return ret
}

//查找分叉节点
func FindActiveNodes(node INodeModel) []string {
	fmt.Println("查找分叉节点")
	return FindForkTaskNames(node)
}
