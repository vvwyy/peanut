# Peanut
> peanut is for gopher

## concurrent package

- executor 执行任务管理器
- future 及其相关接口，是对任务的抽象，提供对任务进行查询是否完成、获取执行接口、超时控制等接口

**Example 1**： 
```
func Example01() {
	executor := NewExecutor()

	// 定义任务
	executable1 := func() (interface{}, error) {
		return "Executable-1", nil
	}
	executable2 := func() (interface{}, error) {
		return "Executable-2", nil
	}

	// 提交执行
	future1, err := executor.Go(executable1)
	if err != nil {
		fmt.Println(err)
		return
	}
	future2, err := executor.Go(executable2)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 获取结果
	ret, err := future1.Get()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("result is : ", ret)
	}

	ret, err = future2.Get()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("result is : ", ret)
	}
}
```

**Example 2**： 通过 `future#GetWithTimeout()` 进行超时控制
```
func Example02() {
	executor := NewExecutor()

	// 定义任务
	executable := func() (interface{}, error) {
		// execute some time
		time.Sleep(1 * time.Second)
		return "Executable", nil
	}

	// 提交执行
	f, err := executor.Go(executable)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 可超时获取结果
	ret, err := f.GetWithTimeout(500 * time.Millisecond)
	if err != nil {
		fmt.Println("timeout")
	} else {
		fmt.Println("result is : ", ret)
	}
}
```

**Example 3**: `future#Get()` 支持并发执行
```
func Example03() {
	executor := NewExecutor()

	executable := func() (interface{}, error) {
		return "Executable", nil
	}

	future, err := executor.Go(executable)
	if err != nil {
		fmt.Println(err)
		return
	}

	go func() {
		ret, err := future.Get()  // Get 
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("result is : ", ret)
		}
	}()

	ret, err := future.Get() // Get
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("result is : ", ret)
	}

	time.Sleep(1 * time.Second) // waiting for goroutine
}
```

**Example 4**: 
```
type Person struct {
	Name string
	Age int32
}


func Example04() {
	executor := NewExecutor()

	executable := func() (interface{}, error) {
		return Person{
			Name: "Bennett",
			Age: 22,
		}, nil
	}

	f, err := executor.Go(executable)
	if err != nil {
		fmt.Println(err)
		return
	}

	ret, err := f.Get()
	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println("result is : ", ret)
	}

	if p, ok :=ret.(Person); ok {
		fmt.Println(p.Name, p.Age)
	}
}
```




