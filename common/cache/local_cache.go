package cache

import (
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

const mutexLocked = 1 << iota

// 实现一个带有 tryLock 功能的锁
type Mutex struct {
	sync.Mutex
}

func (mu *Mutex) TryLock() bool {
	return atomic.CompareAndSwapInt32((*int32)(unsafe.Pointer(&mu.Mutex)), 0, mutexLocked)
}

const maximumCapacity int = 1 << 30

// 1. 空间换时间。 通过冗余的两个数据结构(read、dirty),减少加锁对性能的影响。
// 2. 使用只读数据(read)，避免读写冲突。
// 3. 动态调整，miss次数多了之后，将dirty数据提升为read。
// 4. double-checking。
// 5. 延迟删除。 删除一个键值只是打标记，只有在提升dirty的时候才清理删除的数据。
// 6. 优先从read读取、更新、删除，因为对read的读取不需要锁。
type LocalCache struct {
	mu Mutex

	read  atomic.Value // readOnly
	dirty map[interface{}]*referenceEntry

	misses int

	loader Loader

	expireAfterAccessDuration time.Duration
	expireAfterWriteDuration  time.Duration

	//ticker time.Ticker
	//statsCounter SimpleStatsCounter
}

// entry 被清除后的标记 （read 中的数据不会被直接删除，而是先被标记为删除）
//
// 当 entry 被从 dirty 中删除时，标记为删除
var expunged = unsafe.Pointer(new(interface{}))

type readOnly struct {
	m       map[interface{}]*referenceEntry
	amended bool // 当 LocalCache.dirty 中包含 m 中没有的 entry 时为 true
}

// 1. nil: entry 已被删除了，并且 cache.dirty == nil
// 2. expunged: entry已被删除了，并且m.dirty不为nil，而且这个entry不存在于m.dirty中
// 3. 其它情况下： entry是一个正常的值
type referenceEntry struct {
	p unsafe.Pointer // *interface{}

	accessTime time.Duration // 记录 entry 最近一次被访问到时间
	writeTime  time.Duration // 记录 entry 最近一次被写入的时间
}

func newEntry(val interface{}) *referenceEntry {
	return &referenceEntry{
		p:          unsafe.Pointer(&val),
		accessTime: 0,
		writeTime:  0,
	}
}

// ========================================================================================================
// =========== local cache
// ========================================================================================================

func (cache *LocalCache) Get(key interface{}) (interface{}, error) {
	read, _ := cache.read.Load().(readOnly)
	if entry, ok := read.m[key]; ok {
		now := time.Duration(time.Now().UnixNano())
		value, ok := cache.getLiveValue(entry, now)
		if ok {
			cache.recordRead(entry, now)
			return value, nil
		}
	}

	// locked GetOrLoad

	cache.mu.Lock()
	defer cache.mu.Unlock()
	read, _ = cache.read.Load().(readOnly)

	now := time.Duration(time.Now().Nanosecond())

	var actual interface{}
	var existing bool

	if entry, ok := read.m[key]; ok {
		value, ok := entry.load()
		if ok && !cache.isExpired(entry, now) {
			actual, existing = value, true
		} else {
			if entry.unexpungeLocked() {
				cache.dirty[key] = entry
			}

			// 通过 loader 获取新值
			newValue, loadErr := cache.loader.Load(key)
			if nil != loadErr {
				return nil, loadErr
			}

			entry.storeLocked(&newValue)
			actual, existing = newValue, false
		}

		if existing {
			cache.recordRead(entry, now)
		} else {
			cache.recordWrite(entry, now)
		}
	} else if entry, ok := cache.dirty[key]; ok {
		value, ok := entry.load()
		if ok && !cache.isExpired(entry, now) {
			actual, existing = value, true
		} else {
			if entry.unexpungeLocked() {
				cache.dirty[key] = entry
			}

			// 通过 loader 获取新值
			newValue, loadErr := cache.loader.Load(key)
			if nil != loadErr {
				return nil, loadErr
			}

			entry.storeLocked(&newValue)
			actual, existing = newValue, false
		}

		if existing {
			cache.recordRead(entry, now)
		} else {
			cache.recordWrite(entry, now)
		}
		cache.missLocked()
	} else {
		// 通过 loader 获取新值
		newValue, loadErr := cache.loader.Load(key)
		if nil != loadErr {
			return nil, loadErr
		}

		if !read.amended {
			cache.dirtyLocked()
			cache.read.Store(readOnly{m: read.m, amended: true})
		}
		newEntry := newEntry(newValue)
		cache.dirty[key] = newEntry

		cache.recordWrite(newEntry, now)
		actual, existing = newValue, false
	}
	//cache.mu.Unlock()

	return actual, nil

}

func (cache *LocalCache) GetIfPresent(key interface{}) interface{} { //(interface{}, bool) {
	read, _ := cache.read.Load().(readOnly)
	entry, ok := read.m[key]

	if !ok && read.amended {
		cache.mu.Lock()
		// 双检查，避免加锁的时候 dirty 提升为 read,这个时候 read 可能被替换了。
		read, _ = cache.read.Load().(readOnly)
		entry, ok = read.m[key]
		if !ok && read.amended {
			entry, ok = cache.dirty[key]
			// 不管 dirty 中存不存在，都将misses计数加一
			// missLocked() 中满足条件后就会提升m.dirty
			cache.missLocked()
		}
		cache.mu.Unlock()
	}

	if ok {
		now := time.Duration(time.Now().UnixNano())
		value, ok := cache.getLiveValue(entry, now)
		if ok {
			cache.recordRead(entry, now)
			return value//, true
		}
	}

	return nil//, false
}

func (cache *LocalCache) Put(key, value interface{}) {
	read, _ := cache.read.Load().(readOnly)
	now := time.Duration(time.Now().Nanosecond())
	if entry, ok := read.m[key]; ok && entry.tryStore(&value) {
		cache.recordWrite(entry, now)
		return
	}

	// 元素不存在或者已经被标记为删除
	cache.mu.Lock()
	read, _ = cache.read.Load().(readOnly)
	if entry, ok := read.m[key]; ok {
		if entry.unexpungeLocked() {
			// entry 已经被标记为删除，则表明 dirty != nil，并且 entry 不存在于 dirty 中
			cache.dirty[key] = entry
		}
		entry.storeLocked(&value)

		cache.recordWrite(entry, now)
	} else if entry, ok := cache.dirty[key]; ok {
		entry.storeLocked(&value)

		cache.recordWrite(entry, now)
	} else { // 新键值
		if !read.amended { // dirty 中没有新的数据，往 dirty 中增加第一个新键
			cache.dirtyLocked() // //从 read 中复制未删除的数据
			cache.read.Store(readOnly{m: read.m, amended: true})
		}
		newEntry := newEntry(value)
		cache.dirty[key] = newEntry

		cache.recordWrite(newEntry, now)
	}

	cache.mu.Unlock()
}

func (cache *LocalCache) Delete(key interface{}) {
	read, _ := cache.read.Load().(readOnly)
	entry, ok := read.m[key]
	if !ok && read.amended {
		cache.mu.Lock()
		read, _ = cache.read.Load().(readOnly)
		entry, ok = read.m[key]
		if !ok && read.amended {
			delete(cache.dirty, key)
		}
		cache.mu.Unlock()
	}
	if ok {
		entry.delete()
	}
}

func (cache *LocalCache) Range(f func(key, value interface{}) bool) {
	// We need to be able to iterate over all of the keys that were already
	// present at the start of the call to Range.
	// If read.amended is false, then read.cache satisfies that property without
	// requiring us to hold cache.mu for a long time.
	read, _ := cache.read.Load().(readOnly)
	if read.amended {
		// cache.dirty contains keys not in read.cache. Fortunately, Range is already O(N)
		// (assuming the caller does not break out early), so a call to Range
		// amortizes an entire copy of the map: we can promote the dirty copy
		// immediately!
		cache.mu.Lock()
		read, _ = cache.read.Load().(readOnly)
		if read.amended {
			read = readOnly{m: cache.dirty}
			cache.read.Store(read)
			cache.dirty = nil
			cache.misses = 0
		}
		cache.mu.Unlock()
	}

	for k, e := range read.m {
		v, ok := e.load()
		if !ok {
			continue
		}
		if !f(k, v) {
			break
		}
	}
}

// ------------------------------------------------------

func (cache *LocalCache) missLocked() {
	cache.misses++
	if cache.misses < len(cache.dirty) {
		return
	}
	cache.read.Store(readOnly{m: cache.dirty})
	cache.dirty = nil
	cache.misses = 0
}

func (cache *LocalCache) dirtyLocked() {
	if cache.dirty != nil {
		return
	}

	read, _ := cache.read.Load().(readOnly)
	cache.dirty = make(map[interface{}]*referenceEntry, len(read.m))
	for k, e := range read.m {
		if !e.tryExpungeLocked() { // 这里不用检查过期，过期标记在 getLiveEntry 时触发
			cache.dirty[k] = e
		}
	}
}

func (cache *LocalCache) getLiveValue(entry *referenceEntry, now time.Duration) (interface{}, bool) {
	value, ok := entry.load()
	if !ok {
		return nil, false
	}
	if cache.isExpired(entry, now) {
		cache.tryExpireEntries(now)
		return nil, false
	}
	return value, true
}

func (cache *LocalCache) recordRead(entry *referenceEntry, now time.Duration) {
	if cache.expiresAfterAccess() {
		entry.accessTime = now
	}
}

func (cache *LocalCache) recordWrite(entry *referenceEntry, now time.Duration) {
	if cache.expiresAfterAccess() {
		entry.accessTime = now
	}
	if cache.expiresAfterWrite() {
		entry.writeTime = now
	}
}

func (cache *LocalCache) isExpired(entry *referenceEntry, now time.Duration) bool {
	if cache.expiresAfterAccess() && (now-entry.accessTime >= cache.expireAfterAccessDuration) {
		return true
	}
	if cache.expiresAfterWrite() && (now-entry.writeTime >= cache.expireAfterWriteDuration) {
		return true
	}
	return false
}

func (cache *LocalCache) expiresAfterAccess() bool {
	return cache.expireAfterAccessDuration > 0
}

func (cache *LocalCache) expiresAfterWrite() bool {
	return cache.expireAfterWriteDuration > 0
}

func (cache *LocalCache) tryExpireEntries(now time.Duration) {
	if cache.mu.TryLock() { // NOTE: 这里只是尝试做过期标记，不做强制
		cache.expireEntries(now)
		cache.mu.Unlock()
	}
}

func (cache *LocalCache) expireEntries(now time.Duration) {
	read, _ := cache.read.Load().(readOnly)
	// 检查所有 entry，如果有过期的 entry，则进行过期标记
	// NOTE: 这里不进行删除，删除只在 dirty 上升为 read 时进行
	for _, entry := range read.m {
		if cache.isExpired(entry, now) {
			entry.delete()
		}
	}
}

// ========================================================================================================
// =========== reference entry
// ========================================================================================================

func (entry *referenceEntry) load() (value interface{}, ok bool) {
	p := atomic.LoadPointer(&entry.p)
	if p == nil || p == expunged {
		return nil, false
	}
	return *(*interface{})(p), true
}

func (entry *referenceEntry) tryStore(i *interface{}) bool {
	p := atomic.LoadPointer(&entry.p)
	if p == expunged {
		return false
	}
	for {
		if atomic.CompareAndSwapPointer(&entry.p, p, unsafe.Pointer(i)) {
			return true
		}
		p = atomic.LoadPointer(&entry.p)
		if p == expunged {
			return false
		}
	}
}

func (entry *referenceEntry) delete() (hadValue bool) {
	for {
		p := atomic.LoadPointer(&entry.p)
		if p == nil || p == expunged {
			return false
		}
		if atomic.CompareAndSwapPointer(&entry.p, p, nil) { // CAS
			return true
		}
	}
}

func (entry *referenceEntry) tryExpungeLocked() (isExpunged bool) {
	p := atomic.LoadPointer(&entry.p)
	for p == nil {
		if atomic.CompareAndSwapPointer(&entry.p, nil, expunged) {
			return true
		}
		p = atomic.LoadPointer(&entry.p)
	}
	return p == expunged
}

// 如果 entry 已经被标记为删除，那么一定要在 unlock 之前，将它加入到 dirty 中
func (entry *referenceEntry) unexpungeLocked() (wasExpunged bool) {
	return atomic.CompareAndSwapPointer(&entry.p, expunged, nil)
}

// 一定要保证 entry 没有被标记为删除
func (entry *referenceEntry) storeLocked(i *interface{}) {
	atomic.StorePointer(&entry.p, unsafe.Pointer(i))
}
