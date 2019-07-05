package cache

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
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

	localCache := NewBuilder().
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

	localCache := NewBuilder().
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

	cache := NewBuilder().
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

	cache := NewBuilder().
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

	cache := NewBuilder().
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

	cache := NewBuilder().
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

	cache := NewBuilder().
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

func TestLocalCache7(t *testing.T) {

	loader := &StringLoader{}

	cache := NewBuilder().
		ExpireAfterWrite(100 * time.Second).
		ExpireAfterAccess(200 * time.Second).
		Build(loader)

	const size = 1000


	go func() {
		for i := 0; i < size; i++ {
			cache.Get(i)
		}
	}()

	time.Sleep(time.Second*5)

	for i:=0; i< size; i++ {
		go func() {
			cache.Get(i)
		}()
	}

	time.Sleep(time.Second * 5)
}

func TestConcurrentRange(t *testing.T) {
	const mapSize = 1 << 10

	loader := &StringLoader{}

	cache := NewBuilder().Build(loader)
	for n := int64(1); n <= mapSize; n++ {
		cache.Put(n, int64(n))
	}

	done := make(chan struct{})
	var wg sync.WaitGroup
	defer func() {
		close(done)
		wg.Wait()
	}()
	for g := int64(runtime.GOMAXPROCS(0)); g > 0; g-- {
		r := rand.New(rand.NewSource(g))
		wg.Add(1)
		go func(g int64) {
			defer wg.Done()
			for i := int64(0); ; i++ {
				select {
				case <-done:
					return
				default:
				}
				for n := int64(1); n < mapSize; n++ {
					if r.Int63n(mapSize) == 0 {
						cache.Put(n, n*i*g)
					} else {
						cache.GetIfPresent(n)
					}
				}
			}
		}(g)
	}

	iters := 1 << 10
	if testing.Short() {
		iters = 16
	}
	for n := iters; n > 0; n-- {
		seen := make(map[int64]bool, mapSize)

		cache.Range(func(ki, vi interface{}) bool {
			k, v := ki.(int64), vi.(int64)
			if v%k != 0 {
				t.Fatalf("while Storing multiples of %v, Range saw value %v", k, v)
			}
			if seen[k] {
				t.Fatalf("Range visited key %v twice", k)
			}
			seen[k] = true
			return true
		})

		if len(seen) != mapSize {
			t.Fatalf("Range visited %v elements of %v-element Map", len(seen), mapSize)
		}
	}
}
