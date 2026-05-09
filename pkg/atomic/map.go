package atomic

import "sync"

type AtomicMap[K comparable, V any] struct {
	mu sync.RWMutex
	ma map[K]V
}

func NewAtomicMap[K comparable, V any]() *AtomicMap[K, V] {
	return &AtomicMap[K, V]{ma: map[K]V{}}
}

func (m *AtomicMap[K, V]) Write(key K, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ma[key] = value
}

func (m *AtomicMap[K, V]) Read(key K) V {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.ma[key]
}

func (m *AtomicMap[K, V]) ReadSafe(key K) (V, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, ok := m.ma[key]
	return value, ok
}

func (m *AtomicMap[K, V]) WithLock(do func(m map[K]V)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	do(m.ma)
}
