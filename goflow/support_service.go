package goflow

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

//根据OrderID得到活动流程
func GetActiveTasksByOrderId(orderId string) []*Task {
	task := &Task{}
	tasks := task.GetActiveTasksByOrderId(orderId)
	return tasks
}

//得到任务的角色
func GetTaskActors(taskModel *TaskModel, execution *Execution) []string {
	assignee := taskModel.Assignee
	if assignee != "" {
		assigneeInf := execution.Args[taskModel.Assignee]
		if assigneeInf == nil {
			assigneeInf = taskModel.Assignee
		}
		switch assigneeInf.(type) {
		case string:
			return strings.Split(assigneeInf.(string), ",")
		case []string:
			return assigneeInf.([]string)
		case int:
			return []string{IntToStr(assigneeInf.(int))}
		default:
		}
	}
	return nil
}

//创建task，并根据model类型决定是否分配参与者
func CreateTask(taskModel *TaskModel, execution *Execution) []*Task {
	// TODO 循环
	actors := GetTaskActors(taskModel, execution)
	args := execution.Args
	args[DEFAULT_KEY_ACTOR] = actors

	task := &Task{
		OrderId:     execution.Order.Id,
		TaskName:    taskModel.Name,
		DisplayName: taskModel.DisplayName,
		CreateTime:  time.Now(),
		TaskType:    ProcessTaskType(taskModel.TaskType),
		Model:       taskModel,
		ExpireTime:  ProcessTime(args, taskModel.ExpireTime),
		Variable:    MapToJson(args),
	}
	if execution.Task == nil {
		task.ParentTaskId = DEFAULT_START_ID
	} else {
		task.ParentTaskId = execution.Task.Id
	}

	action := args[taskModel.Action]
	if action == nil {
		task.Action = taskModel.Action
	} else {
		task.Action = action.(string)
	}

	tasks := make([]*Task, 0)

	if ProcessPerformType(taskModel.PerformType) == PO_ANY {
		SaveTask(task, actors...)
		tasks = append(tasks, task)
	} else {
		for _, actor := range actors {
			singleTask := *task
			pSingleTask := &singleTask
			SaveTask(pSingleTask, actor)
			tasks = append(tasks, pSingleTask)
		}
	}
	return tasks
}

//保存任务
func SaveTask(task *Task, actors ...string) {
	task.Id = NewUUID()
	task.PerformType = PO_ANY
	Save(task, task.Id)
	AssignTask(task.Id, actors...)
}

//根据已有任务、任务类型、参与者创建新的任务，适用于转派，动态协办处理
func CreateNewTask(taskId string, taskType TASK_ORDER, actors ...string) {
	fmt.Println("根据已有任务、任务类型、参与者创建新的任务，适用于转派，动态协办处理")
	task := &Task{}
	if task.GetTaskById(taskId) {
		newTask := *task
		pNewTask := &newTask
		pNewTask.TaskType = taskType
		pNewTask.CreateTime = time.Now()
		pNewTask.ParentTaskId = taskId
		SaveTask(pNewTask, actors...)
	}
}

//驳回任务
func RejectTask(processModel *ProcessModel, currTask *Task) *Task {
	fmt.Println("驳回任务")
	parentTaskId := currTask.ParentTaskId
	if parentTaskId == "" || parentTaskId == DEFAULT_START_ID {
		return nil
	}
	currentNode := processModel.GetNode(currTask.TaskName)
	historyTask := &HistoryTask{}
	if historyTask.GetHistoryTaskById(parentTaskId) {
		parentNode := processModel.GetNode(historyTask.TaskName)
		if CanRejected(currentNode, parentNode) {
			task := historyTask.Undo()
			task.Id = NewUUID()
			task.CreateTime = time.Now()
			Save(task, task.Id)
			AssignTask(task.Id, task.Operator)
			return task
		}
	}
	return nil
}

//撤销任务
func WithdrawTask(taskId string, operator string) *Task {
	fmt.Println("撤销任务")
	historyTask := &HistoryTask{}
	if historyTask.GetHistoryTaskById(taskId) {
		var tasks []*Task
		if historyTask.PerformType == PO_ANY {
			tasks = GetNextAnyActiveTasks(historyTask.Id)
		} else {
			tasks = GetNextAllActiveTasks(historyTask.OrderId, historyTask.TaskName, historyTask.ParentTaskId)
		}
		for _, task := range tasks {
			Delete(task, task.Id)
		}

		task := historyTask.Undo()
		task.Id = NewUUID()
		task.CreateTime = time.Now()
		Save(task, task.Id)
		AssignTask(task.Id, task.Operator)
		return task
	} else {
		return nil
	}
}

//加任务角色
func AddTaskActor(taskId string, performType PERFORM_ORDER, actors ...string) {
	fmt.Println("加任务角色")
	task := &Task{}
	if task.GetTaskById(taskId) {
		if performType == PO_ANY {
			AssignTask(taskId, actors...)
			v := JsonToMap(task.Variable)
			oldActor := v[DEFAULT_KEY_ACTOR].(string)
			v[DEFAULT_KEY_ACTOR] = oldActor + "," + strings.Join(actors, ",")
			task.Variable = MapToJson(v)
			Update(task, task.Id)
		} else {
			for _, actor := range actors {
				newTask := *task
				pNewTask := &newTask
				pNewTask.Id = NewUUID()
				pNewTask.CreateTime = time.Now()
				pNewTask.Operator = actor
				v := JsonToMap(task.Variable)
				v[DEFAULT_KEY_ACTOR] = actor
				task.Variable = MapToJson(v)
				Save(pNewTask, pNewTask.Id)
				AssignTask(pNewTask.Id, actor)
			}
		}
	}
}

//删除任务角色
func RemoveTaskActor(taskId string, actors ...string) {
	fmt.Println("删除任务角色")
	task := &Task{}
	if task.GetTaskById(taskId) {
		if len(actors) > 0 && task.TaskType == TO_MAJOR {
			for _, actorId := range actors {
				taskActor := &TaskActor{
					TaskId:  taskId,
					ActorId: actorId,
				}
				DeleteObj(taskActor)
			}
			v := JsonToMap(task.Variable)
			oldActors := strings.Split(v[DEFAULT_KEY_ACTOR].(string), ",")
			for _, actor := range actors {
				for k, s := range oldActors {
					if strings.ToUpper(s) == strings.ToUpper(actor) {
						oldActors = StringsRemoveAtIndex(oldActors, k)
						break
					}
				}
			}
			v[DEFAULT_KEY_ACTOR] = oldActors
			task.Variable = MapToJson(v)
			Update(task, task.Id)
		}
	}
}

//结束并且提取任务
func TakeTask(taskId string, operator string) *Task {
	fmt.Println("结束并且提取任务")
	task := &Task{}
	success := task.GetTaskById(taskId)

	if success {
		if !IsAllowed(task, operator) {
			return nil
		}
		task.Operator = operator
		task.FinishTime = time.Now()
		Update(task, task.Id)
		return task
	} else {
		return nil
	}
}

//对指定的任务分配参与者。参与者可以为用户、部门、角色
func AssignTask(taskId string, actors ...string) {
	fmt.Println("对指定的任务分配参与者。参与者可以为用户、部门、角色")
	if len(actors) == 0 {
		return
	} else {
		for _, actorId := range actors {
			if actorId != "" {
				taskActor := &TaskActor{
					Id:      NewUUID(),
					TaskId:  taskId,
					ActorId: actorId,
				}
				Save(taskActor, taskActor.Id)
			}
		}
	}
}

//是否被授权执行任务
func IsAllowed(task *Task, operator string) bool {
	fmt.Println("是否被授权执行任务")
	if strings.ToUpper(operator) == string(ER_ADMIN) ||
		strings.ToUpper(operator) == string(ER_AUTO) ||
		(task.Operator != "" && strings.ToUpper(task.Operator) == strings.ToUpper(operator)) {
		return true
	} else {
		taskActors := task.GetTaskActors()
		return len(taskActors) == 0
	}
}

//完成任务
func CompleteTask(taskId string, operator string, args map[string]interface{}) *Task {
	fmt.Println("完成任务")
	task := &Task{}
	if task.GetTaskById(taskId) {
		task.Variable = MapToJson(args)
		if IsAllowed(task, operator) {
			historyTask := &HistoryTask{
				Id:           task.Id,
				OrderId:      task.OrderId,
				CreateTime:   task.CreateTime,
				DisplayName:  task.DisplayName,
				TaskName:     task.TaskName,
				TaskType:     task.TaskType,
				ExpireTime:   task.ExpireTime,
				Action:       task.Action,
				ParentTaskId: task.ParentTaskId,
				Variable:     task.Variable,
				PerformType:  task.PerformType,
				FinishTime:   time.Now(),
				Operator:     operator,
				TaskState:    FS_FINISH,
			}
			Save(historyTask, historyTask.Id)
			Delete(task, task.Id)

			taskActors := GetTaskActorsByTaskId(historyTask.Id)
			for _, taskActor := range taskActors {
				historyTaskActor := &HistoryTaskActor{
					Id:      taskActor.Id,
					TaskId:  taskActor.TaskId,
					ActorId: taskActor.ActorId,
				}
				Save(historyTaskActor, historyTaskActor.Id)
				Delete(taskActor, taskActor.Id)
			}
		}
		return task
	}
	return nil
}

//创建Order
func CreateOrder(process *Process, operator string, args map[string]interface{},
	parentId string, parentNodeName string) *Order {
	fmt.Println("创建Order")
	now := time.Now()
	order := &Order{
		Id:             NewUUID(),
		ParentId:       parentId,
		ParentNodeName: parentNodeName,
		ProcessId:      process.Id,
		Creator:        operator,
		CreateTime:     now,
		LastUpdateTime: now,
		LastUpdator:    operator,
		Variable:       MapToJson(args),
		OrderNo:        GenerateNo(),
	}
	orderNo := args[string(ER_ORDERNO)]
	if orderNo != nil && orderNo.(string) != "" {
		order.OrderNo = orderNo.(string)
	}
	model := process.Model
	if model != nil {
		order.ExpireTime = ProcessTime(args, model.ExpireTime)
	}
	SaveOrder(order)
	return order
}

//生成OrderNo
func GenerateNo() string {
	fmt.Println("生成OrderNo")
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("%s_%04d", FormatTime(time.Now(), ORDERNO_TIME_LAYOUT), r.Intn(1000))
}

//保存Order
func SaveOrder(order *Order) {
	fmt.Println("保存Order")
	historyOrder := &HistoryOrder{}
	historyOrder.DataByOrder(order)

	historyOrder.OrderState = FS_ACTIVITY
	Save(order, order.Id)
	Save(historyOrder, historyOrder.Id)
}

//完成Order
func CompleteOrder(id string) {
	fmt.Println("完成Order")
	order := &Order{}
	if order.GetOrderById(id) {

		historyOrder := &HistoryOrder{}
		if historyOrder.GetHistoryOrderById(id) {
			historyOrder.OrderState = FS_FINISH

			Update(historyOrder, historyOrder.Id)
			Delete(order, order.Id)
		}
	}
}

//唤醒Order
func ResumeOrder(id string) {
	fmt.Println("唤醒Order")
	historyOrder := &HistoryOrder{}
	if historyOrder.GetHistoryOrderById(id) {
		historyOrder.OrderState = FS_ACTIVITY
		order := historyOrder.Undo()

		Save(order, order.Id)
		Save(historyOrder, historyOrder.Id)
	}

}

//终止Order
func TerminateOrder(id string, operator string) {
	fmt.Println("终止Order")
	tasks := GetActiveTasksByOrderId(id)
	for _, task := range tasks {
		CompleteTask(task.Id, operator, nil)
	}

	order := &Order{}
	if order.GetOrderById(id) {
		historyOrder := &HistoryOrder{}
		historyOrder.DataByOrder(order)
		historyOrder.OrderState = FS_TERMINATION
		historyOrder.FinishTime = time.Now()

		Update(historyOrder, historyOrder.Id)
		Delete(order, order.Id)
	}
}

//得到代理人
func GetSurrogate(operator string, processName string) string {
	fmt.Println("得到代理人")
	var result []string
	now := time.Now()
	surrogates := GetSurrogateSQL(`"State" = ? and "StartTime" =< ?  and "EndTime" >= ? and "Operator" in (?) and "ProcessName" in (?)`, SS_ENABLE, now, now, operator, processName)
	for _, surrogate := range surrogates {
		result = append(result, surrogate.Surrogate)
	}
	return strings.Join(result, ",")
}

//创建抄送
func CreateCCOrder(orderId string, creator string, actorIds ...string) {
	fmt.Println("创建抄送")
	for _, actorId := range actorIds {
		ccorder := &CCOrder{
			Id:         NewUUID(),
			OrderId:    orderId,
			ActorId:    actorId,
			Creator:    creator,
			State:      FS_ACTIVITY,
			CreateTime: time.Now(),
		}
		Save(ccorder, ccorder.Id)
	}
}

//更新抄送记录状态为已阅
func UpdateCCStatus(orderId string, actorIds ...string) {
	fmt.Println("更新抄送记录状态为已阅")
	ccorders := GetCCOrder(orderId, actorIds...)
	for _, ccorder := range ccorders {
		ccorder.State = FS_FINISH
		ccorder.FinishTime = time.Now()
		Update(ccorder, ccorder.Id)
	}
}

//删除指定的抄送记录
func DeleteCCOrder(orderId string, actorId string) {
	fmt.Println("删除指定的抄送记录")
	ccorders := GetCCOrder(orderId, actorId)
	for _, ccorder := range ccorders {
		Delete(ccorder, ccorder.Id)
	}
}
