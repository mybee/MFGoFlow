# MFGoFlow

 ![image](https://github.com/mybee/MFGoFlow/tree/master/img/Snip2.png)

 马氏工作流

```var engine *Engine
var order *Order

func main()  {

	bytes := LoadXML("/Users/deer_mac/Desktop/VERYIMPORTANT/Go/项目/工作流/main/res/qingjia.xml")
	engine = NewEngineByConfig("/Users/deer_mac/Desktop/VERYIMPORTANT/Go/项目/工作流/goflow/conf/app.conf")
	processId := engine.Deploy(bytes, "")
	fmt.Println("processId:", processId)
	args := map[string]interface{}{
		"task1.operator": []string{"1"},
		"task2.operator": []string{"3"},
		"task3.operator": []string{"4"},
		"task4.operator": []string{"5"},
		"task5.operator": []string{"6"},
		"name": []string{"mafeng"},
		"do": []string{"呵呵哈哈哈"},
	}
	order = engine.StartInstanceById(processId, "2", args)
	fmt.Println("order:", order)
	fmt.Printf("OrderId %s", order.Id)
	tasks := GetActiveTasksByOrderId(order.Id)

	for _, task := range tasks {
		fmt.Println("task:", task)
		engine.ExecuteTask(task.Id, "1", args)
		//engine.ExecuteTask(task.Id, "1", args)
	}

	http.HandleFunc("/exe", exe)
	http.ListenAndServe(":9000", nil)
}

 func exe(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	name, found1 := req.Form["name"]
	do, found2 := req.Form["do"]
	if !(found1 && found2) {
		fmt.Fprint(w, "请勿非法访问")
		return
	}
	tasks := GetActiveTasksByOrderId(order.Id)
	if len(tasks) < 1 {
		w.Write([]byte("任务已经完成"))
	}
	args := map[string]interface{}{
		"task1.operator": []string{name[0]},
		"task2.operator": []string{name[0]},
		"task3.operator": []string{name[0]},
		"task4.operator": []string{name[0]},
		"task5.operator": []string{name[0]},
		"task6.operator": []string{name[0]},
		"task7.operator": []string{name[0]},
		"task8.operator": []string{name[0]},
		"name": []string{name[0]},
		"do": []string{do[0]},
		"content":        250.0,
	}
	for _, task := range tasks {
		fmt.Println("task:", task)
		engine.ExecuteTask(task.Id, "1", args)
		//engine.ExecuteTask(task.Id, "1", args)
	}

}
```
 ![image](https://github.com/mybee/MFGoFlow/tree/master/img/Snip20170802_13.png)