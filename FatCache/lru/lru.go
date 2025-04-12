package lru

import (
	"container/list"
	"fmt"
)

type Cache struct {
	maxBytes int64
	nbytes   int64
	ll       *list.List
	cache    map[string]*list.Element
	onEvict  func(key string, value Value)
}

type Value interface {
	Len() int
}

type entry struct {
	key   string
	value Value
}

func (e *entry) Len() int {
	return e.value.Len()
}

func New(maxBytes int64, onEvict func(key string, value Value)) *Cache {
	if maxBytes <= 0 {
		return nil
	}
	return &Cache{
		maxBytes: maxBytes,
		ll:       list.New(),
		cache:    make(map[string]*list.Element),
		onEvict:  onEvict,
	}
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.onEvict != nil {
			c.onEvict(kv.key, kv.value)
		}
	}
}

func (c *Cache) Add(key string, value Value) {

	if int64(len(key))+int64(value.Len()) > c.maxBytes {
		fmt.Printf("Value for key %s is too large to cache\n", key)
		return // 超大值直接返回，不缓存
	}

	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.Len())
		kv.value = value
	} else {
		ele := &entry{key, value}
		listEle := c.ll.PushFront(ele)
		c.cache[key] = listEle
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

func (c *Cache) Len() int {
	return c.ll.Len()
}
