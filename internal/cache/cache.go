package cache

import (
	"container/heap"
	"sync"
)

type Item struct {
	key   string
	value interface{}
	freq  int
	index int
}

type LFUCache struct {
	mu       sync.RWMutex
	cap      int
	data     map[string]*Item
	freqList *freqHeap
	minFreq  int
}

func NewLFUCache(capacity int) *LFUCache {
	fh := &freqHeap{}
	heap.Init(fh)
	return &LFUCache{
		cap:      capacity,
		data:     make(map[string]*Item),
		freqList: fh,
		minFreq:  0,
	}
}

func (c *LFUCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if item, ok := c.data[key]; ok {
		item.freq++
		heap.Fix(c.freqList, item.index)
		return item.value, true
	}
	return nil, false
}

func (c *LFUCache) Set(key string, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cap <= 0 {
		return
	}

	if item, ok := c.data[key]; ok {
		item.value = value
		item.freq++
		heap.Fix(c.freqList, item.index)
		return
	}

	if len(c.data) >= c.cap {
		c.eject()
	}

	item := &Item{
		key:   key,
		value: value,
		freq:  1,
	}
	heap.Push(c.freqList, item)
	c.data[key] = item
}

func (c *LFUCache) Delete(key string) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if item, found := c.data[key]; found {
		c.remove(item)
	}
}

func (c *LFUCache) ClearCache() {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for key, _ := range c.data {
		item := c.data[key]
		c.remove(item)
	}
}

func (c *LFUCache) eject() {
	item := heap.Pop(c.freqList).(*Item)
	delete(c.data, item.key)
}

func (c *LFUCache) remove(item *Item) {
	heap.Remove(c.freqList, item.index)
	delete(c.data, item.key)
}

type freqHeap []*Item

func (h freqHeap) Len() int {
	return len(h)
}

func (h freqHeap) Less(i, j int) bool {
	return h[i].freq < h[j].freq
}

func (h freqHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *freqHeap) Push(x any) {
	item := x.(*Item)
	item.index = len(*h)
	*h = append(*h, item)
}

func (h *freqHeap) Pop() interface{} {
	old := *h
	item := old[len(old)-1]
	old[len(old)-1] = nil
	item.index = -1
	*h = old[:len(old)-1]
	return item
}
