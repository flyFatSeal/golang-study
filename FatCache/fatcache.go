package fatcache

import (
	"fmt"
	"sync"
)

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

// 实现 Getter 接口
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

type Group struct {
	name      string
	mainCache Cache
	getter    Getter
}

func GetGroup(name string) *Group {
	return groups[name]
}

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	mu.Lock()
	defer mu.Unlock()

	groups[name] = &Group{
		name:      name,
		mainCache: Cache{cacheBytes: cacheBytes},
		getter:    getter,
	}
	return groups[name]

}

func (g *Group) Get(key string) (ByteView, error) {
	if value, ok := g.mainCache.get(key); ok {
		fmt.Println("Cache hit")
		return value, nil
	}
	fmt.Println("Cache miss - loading data")
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	// 检查值大小
	if int64(len(key))+int64(len(bytes)) > g.mainCache.cacheBytes {
		return ByteView{}, fmt.Errorf("value for key %s is too large to cache", key)
	}
	// 将数据添加到缓存
	g.populateCache(key, bytes)

	return ByteView{b: bytes}, nil
}

// 将数据添加到缓存
func (g *Group) populateCache(key string, bytes []byte) {
	g.mainCache.add(key, ByteView{b: bytes})
}
