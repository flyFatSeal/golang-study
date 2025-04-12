package main

import (
	"fatcache"
	"net/http"

	"log"
)

// 定义一个结构体
type MyGetter struct{}

func main() {
	// 模拟一个 Getter 函数，从数据源加载数据
	getter := fatcache.GetterFunc(func(key string) ([]byte, error) {
		log.Printf("Loading data for key: %s", key)
		return []byte("Value for " + key), nil
	})

	// 创建一个缓存组
	addr := "localhost:9000"
	fatcache.NewGroup("test", 30, getter)
	peers := fatcache.NewHTTPPool("localhost:9000")

	http.ListenAndServe(addr, peers)

}
