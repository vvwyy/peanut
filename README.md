# Peanut
> peanut is for gopher

## concurrent package

- executor 执行任务管理器
- future 及其相关接口，是对任务的抽象，提供对任务进行查询是否完成、获取执行接口、超时控制等接口

**Example 0**： 
```
func Example() {
	executor := NewExecutor()

	ret, err := executor.Go(func() (interface{}, error) {
		return "Executable", nil
	}).Get()

	if err != nil {
		return
	}
	fmt.Println("future.Get(), result is : ", ret)
}
```

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
	future1 := executor.Go(executable1)
	future2 := executor.Go(executable2)

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
	f := executor.Go(executable)

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

	future := executor.Go(executable)

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

	f := executor.Go(executable)

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

**Example 5**: `executor#Shutdown`
```
func Example05() {
	executor := NewExecutor()

	executable := func() (interface{}, error) {
		time.Sleep(10*time.Second)
		return "Executable", nil
	}

	f := executor.Go(executable)

	go func() {
		time.Sleep(2*time.Second)
		executor.Shutdown()
	}()

	ret, err := f.Get()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("result is : ", ret)
		return
	}
}
```

**Example 6** 批量处理
```
func Example6() {
	executor := NewExecutor()

	futures := make([]Future, 0)

	for i := 0; i < 10; i++ {
		count := i
		future := executor.Go(func() (interface{}, error) {
			seconds := time.Duration(time.Now().Second() * 10)
			time.Sleep(seconds * time.Millisecond)
			return fmt.Sprintf("Executable-%d", count), nil
		})
		futures = append(futures, future)
	}

	wg := sync.WaitGroup{}
	for _, future := range futures {
		wg.Add(1)
		go func(f Future) {
			ret, err := f.Get()
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Printf("result is: %s \n", ret)
			}
			wg.Done()
		}(future)
	}
	wg.Wait()
}
```

