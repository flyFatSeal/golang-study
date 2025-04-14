package fatcache

import (
	"fatcache/singleflight"
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
	peers     PeerPicker
	loader    *singleflight.Group
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
		loader:    &singleflight.Group{},
	}
	return groups[name]

}

func (g *Group) Get(key string) (ByteView, error) {
	if value, ok := g.mainCache.get(key); ok {
		fmt.Println("Cache hit")
		return value, nil
	}
	fmt.Println("Cache miss - loading data")

	bytes, err := g.load(key)

	if err != nil {
		return ByteView{}, err
	}
	// 检查值大小
	if int64(len(key))+int64(bytes.Len()) > g.mainCache.cacheBytes {
		return ByteView{}, fmt.Errorf("value for key %s is too large to cache", key)
	}
	// 将数据添加到缓存
	g.populateCache(key, bytes.ByteSlice())

	return ByteView{b: bytes.ByteSlice()}, nil
}

func (g *Group) RegisterPeerPicker(p PeerPicker) {
	g.peers = p
}

func (g *Group) load(key string) (ByteView, error) {
	value, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				bytes, err := g.getFromPeer(peer, key)
				if err == nil {
					return bytes, nil
				}
			}
		}
		return g.getLocally(key)
	})

	if err == nil {
		return value.(ByteView), nil
	}
	return ByteView{}, fmt.Errorf("failed to load data for key %s: %v", key, err)
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err == nil {
		return ByteView{b: bytes}, nil
	}
	fmt.Println("Failed to get from peer:", err)
	return ByteView{}, fmt.Errorf("no peer found for key %s", key)
}

func (g *Group) getLocally(key string) (ByteView, error) {

	bytes, err := g.getter.Get(key)
	if err == nil {
		return ByteView{b: bytes}, nil
	}

	return ByteView{}, fmt.Errorf("no locally found for key %s", key)
}

// 将数据添加到缓存
func (g *Group) populateCache(key string, bytes []byte) {
	g.mainCache.add(key, ByteView{b: bytes})
}
