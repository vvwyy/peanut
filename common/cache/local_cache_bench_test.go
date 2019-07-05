package cache

import (
	"fmt"
	"reflect"
	"sync/atomic"
	"testing"
)


var loader = &StringLoader{}

type bench struct {
	setup func(*testing.B, *localCache)
	perG  func(b *testing.B, pb *testing.PB, i int, m *localCache)
}

func benchMap(b *testing.B, bench bench) {

	cache := NewBuilder().
		Build(loader)
	b.Run(fmt.Sprintf("%T", cache), func(b *testing.B) {
		cache = reflect.New(reflect.TypeOf(cache).Elem()).Interface().(*localCache)
		if bench.setup != nil {
			bench.setup(b, cache)
		}

		b.ResetTimer()

		var i int64
		b.RunParallel(func(pb *testing.PB) {
			id := int(atomic.AddInt64(&i, 1) - 1)
			bench.perG(b, pb, id*b.N, cache)
		})
	})
	//for _, cache := range [...]*localCache{&localCache{loader: loader}} {
	//
	//}
}

func BenchmarkLoadMostlyHits(b *testing.B) {
	const hits, misses = 1023, 1

	benchMap(b, bench{
		setup: func(_ *testing.B, cache *localCache) {
			//for i := 0; i < hits; i++ {
			//	cache.Get(i)
			//}
			// Prime the map to get it into a steady state.
			for i := 0; i < hits*2; i++ {
				cache.GetIfPresent(i % hits)
			}
		},

		perG: func(b *testing.B, pb *testing.PB, i int, m *localCache) {
			for ; pb.Next(); i++ {
				m.GetIfPresent(i % (hits + misses))
			}
		},
	})
}

func BenchmarkAdversarialDelete(b *testing.B) {
	const mapSize = 1 << 10

	benchMap(b, bench{
		setup: func(_ *testing.B, cache *localCache) {
			for i := 0; i < mapSize; i++ {
				cache.Put(i, i)
			}
		},

		perG: func(b *testing.B, pb *testing.PB, i int, cache *localCache) {
			for ; pb.Next(); i++ {
				cache.GetIfPresent(i)

				if i%mapSize == 0 {
					cache.Range(func(k, _ interface{}) bool {
						cache.Delete(k)
						return false
					})
					cache.Put(i, i)
				}
			}
		},
	})
}
