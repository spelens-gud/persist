package persist

import (
	"sync"
	"sync/atomic"
)

type GenericConcurrentMap[K comparable, V any] struct {
	mu sync.Mutex

	read   atomic.Pointer[readOnlyGeneric[K, V]]
	dirty  map[K]*entryGeneric[V]
	misses int

	// exp 是哨兵指针（标记 expunged 状态），类型 **V，与存储指针层级一致
	exp *(*V)
}
type readOnlyGeneric[K comparable, V any] struct {
	m       map[K]*entryGeneric[V]
	amended bool
}

type entryGeneric[V any] struct {
	p atomic.Pointer[*V] // 指向 *V；三态：nil / exp / *V
}
