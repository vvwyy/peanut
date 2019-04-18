package concurrent

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestExecutor_Go(t *testing.T) {
	executor := NewExecutor()

	executable := func() (interface{}, error) {

		return "Executable", nil
	}

	f, err := executor.Go(executable)
	if err != nil {
		t.Logf("executor go failed. Err: %s", err)
		t.FailNow()
	}

	ret, err := f.Get()
	if err != nil {
		t.Logf("future get result failed. Err: %s", err)
		t.FailNow()
	}
	fmt.Println("future.Get(), result is : ", ret)
}

func TestExecutor_Go_1(t *testing.T) {
	executor := NewExecutor()

	executable := func() (interface{}, error) {
		// execute some time
		time.Sleep(1 * time.Second)
		return "Executable", nil
	}

	f, err := executor.Go(executable)
	if err != nil {
		t.Logf("executor go failed. Err: %s", err)
		t.FailNow()
	}

	ret, err := f.Get()
	if err != nil {
		t.Logf("future get result failed. Err: %s", err)
		t.FailNow()
	}
	fmt.Println("future.Get(), result is : ", ret)
}

func TestExecutor_Go_2(t *testing.T) {
	executor := NewExecutor()

	executable1 := func() (interface{}, error) {
		return "Executable-1", nil
	}

	executable2 := func() (interface{}, error) {
		return "Executable-2", nil
	}

	f1, err := executor.Go(executable1)
	if err != nil {
		t.Logf("executor go-1 failed. Err: %s", err)
		t.FailNow()
	}

	f2, err := executor.Go(executable2)
	if err != nil {
		t.Logf("executor go-2 failed. Err: %s", err)
		t.FailNow()
	}

	ret1, err := f1.Get()
	if err != nil {
		t.Logf("future get result failed. Err: %s", err)
		t.FailNow()
	}
	fmt.Println("future.Get(), result is : ", ret1)

	ret2, err := f2.Get()
	if err != nil {
		t.Logf("future get result failed. Err: %s", err)
		t.FailNow()
	}
	fmt.Println("future.Get(), result is : ", ret2)
}

func TestExecutor_Go_3(t *testing.T) {
	executor := NewExecutor()

	executable := func() (interface{}, error) {
		return "Executable", nil
	}

	f, err := executor.Go(executable)
	if err != nil {
		t.Logf("executor go failed. Err: %s", err)
		t.FailNow()
	}

	go func() {
		ret, err := f.Get()
		if err != nil {
			t.Logf("[GO] future get result failed. Err: %s", err)
			t.FailNow()
		}
		fmt.Println("[GO] future.Get(), result is : ", ret)
	}()

	ret, err := f.Get()
	if err != nil {
		t.Logf("future get result failed. Err: %s", err)
		t.FailNow()
	}
	fmt.Println("future.Get(), result is : ", ret)

	time.Sleep(1 * time.Second) // waiting for goroutine
}

func TestExecutor_Go_4(t *testing.T) {
	executor := NewExecutor()

	executable := func() (interface{}, error) {
		return "Executable", nil
	}

	f, err := executor.Go(executable)
	if err != nil {
		t.Logf("executor go failed. Err: %s", err)
		t.FailNow()
	}

	go func() {
		ret, err := f.Get()
		if err != nil {
			t.Logf("[GO] future get result failed. Err: %s", err)
			t.FailNow()
		}
		fmt.Println("[GO] future.Get(), result is : ", ret)
	}()

	go func() {
		ret, err := f.Get()
		if err != nil {
			t.Logf("[GO] future get result failed. Err: %s", err)
			t.FailNow()
		}
		fmt.Println("[GO] future.Get(), result is : ", ret)
	}()

	ret, err := f.Get()
	if err != nil {
		t.Logf("future get result failed. Err: %s", err)
		t.FailNow()
	}
	fmt.Println("future.Get(), result is : ", ret)

	time.Sleep(1 * time.Second) // waiting for goroutine
}

func TestExecutor_Go_5(t *testing.T) {
	executor := NewExecutor()

	executable := func() (interface{}, error) {
		// execute some time
		time.Sleep(1 * time.Second)
		return "Executable", nil
	}

	f, err := executor.Go(executable)
	if err != nil {
		t.Logf("executor go failed. Err: %s", err)
		t.FailNow()
	}

	_, err = f.GetWithTimeout(500 * time.Millisecond)
	if err != nil {
		t.Logf("future get result is timeout. Err: %s", err)
	} else {
		t.FailNow()
	}
}

func TestExecutor_Go_6(t *testing.T) {
	executor := NewExecutor()

	executable := func() (interface{}, error) {
		// execute some time
		time.Sleep(1 * time.Second)
		return "Executable", nil
	}

	f, err := executor.Go(executable)
	if err != nil {
		t.Logf("executor go failed. Err: %s", err)
		t.FailNow()
	}

	ret, err := f.GetWithTimeout(2 * time.Second)
	if err != nil {
		t.Logf("future get result failed. Err: %s", err)
		t.FailNow()
	} else {
		fmt.Println("future.Get(), result is : ", ret)
	}
}

func TestExecutor_Go_7(t *testing.T) {
	executor := NewExecutor()

	executable := func() (interface{}, error) {
		// execute some time
		time.Sleep(1 * time.Second)
		return "Executable", nil
	}

	f, err := executor.Go(executable)
	if err != nil {
		t.Logf("executor go failed. Err: %s", err)
		t.FailNow()
	}

	go func() {
		_, err = f.GetWithTimeout(500 * time.Millisecond)
		if err != nil {
			t.Logf("[Go] future get result is timeout. Err: %s", err)
		} else {
			t.FailNow()
		}
	}()

	ret, err := f.GetWithTimeout(2 * time.Second)
	if err != nil {
		t.Logf("future get result failed. Err: %s", err)
		t.FailNow()
	} else {
		fmt.Println("future.Get(), result is : ", ret)
	}

}

func TestExecutor_Go_8(t *testing.T) {
	executor := NewExecutor()

	executable := func() (interface{}, error) {

		return nil, errors.New("some error")
	}

	f, err := executor.Go(executable)
	if err != nil {
		t.Logf("executor go failed. Err: %s", err)
		t.FailNow()
	}

	_, err = f.Get()
	if err != nil {
		t.Logf("future get result failed. Err: %s", err)
	} else {
		t.FailNow()
	}
}

func TestExecutor_Go_9(t *testing.T) {
	executor := NewExecutor()

	executable := func() (interface{}, error) {
		time.Sleep(100 * time.Millisecond)
		return nil, errors.New("some error")
	}

	f, err := executor.Go(executable)
	if err != nil {
		t.Logf("executor go failed. Err: %s", err)
		t.FailNow()
	}

	_, err = f.GetWithTimeout(200 * time.Millisecond)
	if err != nil {
		t.Logf("future get result failed. Err: %s", err)
	} else {
		t.FailNow()
	}
}

type Person struct {
	Name string
	Age int32
}

func TestExecutor_Go_10(t *testing.T) {
	executor := NewExecutor()

	executable := func() (interface{}, error) {
		return Person{
			Name: "Bennett",
			Age: 22,
		}, nil
	}

	f, err := executor.Go(executable)
	if err != nil {
		t.Logf("executor go failed. Err: %s", err)
		t.FailNow()
	}

	ret, err := f.Get()
	if err != nil {
		t.Logf("future get result failed. Err: %s", err)
		t.FailNow()
	} else {
		fmt.Println("future.Get(), result is : ", ret)
	}

	if p, ok :=ret.(Person); ok {
		fmt.Println(p.Name, p.Age)
	}
}


func TestExecutor_Go_11(t *testing.T) {
	executor := NewExecutor()

	executable := func() (interface{}, error) {
		time.Sleep(10*time.Second)
		return "Executable", nil
	}

	f, err := executor.Go(executable)
	if err != nil {
		t.Logf("executor go failed. Err: %s", err)
		t.FailNow()
	}

	go func() {
		time.Sleep(2*time.Second)
		executor.Shutdown()
	}()

	ret, err := f.Get()
	if err != nil {
		t.Logf("future get result failed. Err: %s", err)
	} else {
		t.FailNow()
	}
	fmt.Println("future.Get(), result is : ", ret)
}
