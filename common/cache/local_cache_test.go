package cache

import (
	"fmt"
	"testing"
	"time"
)

type StringLoader struct {
}

func (loader *StringLoader) Load(key interface{}) (interface{}, error) {
	fmt.Printf("load from loader: %v \n", key)
	return fmt.Sprintf("A-%v", key), nil
}

func TestLocalCache(t *testing.T) {

	loader := &StringLoader{}

	localCache := newBuilder().
		ExpireAfterWrite(3 * time.Second).
		Build(loader)

	result := localCache.GetIfPresent("aaa")
	if nil != result {
		t.FailNow()
	}

	result, err := localCache.Get("aaa")
	if err != nil {
		t.FailNow()
	}
	if result == nil {
		t.FailNow()
	}

	value, _ := result.(string)
	if value != "A-aaa" {
		t.FailNow()
	}

	time.Sleep(4 * time.Second)

	result = localCache.GetIfPresent("aaa")
	if result != nil {
		t.FailNow()
	}
}

func TestLocalCache1(t *testing.T) {

	loader := &StringLoader{}

	localCache := newBuilder().
		ExpireAfterWrite(3 * time.Second).
		Build(loader)

	result, err := localCache.Get("aaa") // 调用 loader
	if err != nil {
		t.FailNow()
	}
	if result == nil {
		t.FailNow()
	}
	value, _ := result.(string)
	if value != "A-aaa" {
		t.FailNow()
	}

	result, err = localCache.Get("aaa")
	if err != nil {
		t.FailNow()
	}
	if result == nil {
		t.FailNow()
	}
	value, _ = result.(string)
	if value != "A-aaa" {
		t.FailNow()
	}

	time.Sleep(4 * time.Second)

	result = localCache.GetIfPresent("aaa")
	if result != nil {
		t.FailNow()
	}
}

func TestLocalCache2(t *testing.T) {

	loader := &StringLoader{}

	cache := newBuilder().
		ExpireAfterWrite(3 * time.Second).
		Build(loader)

	cache.Put("B", "bbbbb")

	result, err := cache.Get("B")
	if err != nil {
		t.FailNow()
	}
	if result == nil {
		t.FailNow()
	}

	value, _ := result.(string)
	if value != "bbbbb" {
		t.FailNow()
	}

	time.Sleep(3 * time.Second)

	result = cache.GetIfPresent("B")
	if result != nil {
		t.FailNow()
	}

	result, err = cache.Get("B")
	if err != nil {
		t.FailNow()
	}
	if result == nil {
		t.FailNow()
	}
	value, _ = result.(string)
	if value != "A-B" {
		t.FailNow()
	}

}


func TestLocalCache3(t *testing.T) {

	loader := &StringLoader{}

	cache := newBuilder().
		ExpireAfterWrite(30 * time.Second).
		ExpireAfterAccess(2*time.Second).
		Build(loader)

	cache.Put("B", "bbbbb")

	result, err := cache.Get("B")
	if err != nil {
		t.FailNow()
	}
	if result == nil {
		t.FailNow()
	}

	value, _ := result.(string)
	if value != "bbbbb" {
		t.FailNow()
	}

	time.Sleep(2 * time.Second)

	result = cache.GetIfPresent("B")
	if result != nil {
		t.FailNow()
	}

	result, err = cache.Get("B")
	if err != nil {
		t.FailNow()
	}
	if result == nil {
		t.FailNow()
	}
	value, _ = result.(string)
	if value != "A-B" {
		t.FailNow()
	}

}

func TestLocalCache4(t *testing.T) {

	loader := &StringLoader{}

	cache := newBuilder().
		ExpireAfterWrite(3 * time.Second).
		ExpireAfterAccess(10*time.Second).
		Build(loader)

	cache.Put("B", "bbbbb")

	result, err := cache.Get("B")
	if err != nil {
		t.FailNow()
	}
	if result == nil {
		t.FailNow()
	}

	value, _ := result.(string)
	if value != "bbbbb" {
		t.FailNow()
	}

	time.Sleep(3 * time.Second)

	result = cache.GetIfPresent("B")
	if result != nil {
		t.FailNow()
	}

	result, err = cache.Get("B")
	if err != nil {
		t.FailNow()
	}
	if result == nil {
		t.FailNow()
	}
	value, _ = result.(string)
	if value != "A-B" {
		t.FailNow()
	}

}

func TestLocalCache5(t *testing.T) {

	loader := &StringLoader{}

	cache := newBuilder().
		ExpireAfterWrite(100 * time.Second).
		ExpireAfterAccess(200 * time.Second).
		Build(loader)

	const size = 100000
	go func() {
		for i := 0; i < size; i++ {
			cache.Put(i, i)
		}
	}()

	go func() {
		for i := 0; i < size; i++ {
			cache.GetIfPresent(i)
		}
	}()

	time.Sleep(time.Second*5)

	go func() {
		for i := 0; i < size; i++ {
			cache.Get(i)
		}
	}()

	time.Sleep(time.Second * 5)
}

func TestLocalCache6(t *testing.T) {

	loader := &StringLoader{}

	cache := newBuilder().
		ExpireAfterWrite(100 * time.Second).
		ExpireAfterAccess(200 * time.Second).
		Build(loader)

	const size = 1000


	go func() {
		for i := 0; i < size; i++ {
			cache.Put(i, i)
		}
	}()

	for i:=0; i< size; i++ {
		go func() {
			cache.GetIfPresent(i)
		}()
	}

	time.Sleep(time.Second*5)

	for i:=0; i< size; i++ {
		go func() {
			cache.Get(i)
		}()
	}

	time.Sleep(time.Second * 5)
}
