package main

import (
	"fatcache"
	"fmt"
	"log"
)

// 定义一个结构体
type MyGetter struct{}

// 实现 Getter 接口的 Get 方法
func (g *MyGetter) Get(key string) ([]byte, error) {
	log.Printf("MyGetter: Loading data for key: %s", key)
	return []byte("MyGetter Value for " + key), nil
}

func main() {
	// 模拟一个 Getter 函数，从数据源加载数据
	// getter := fatcache.GetterFunc(func(key string) ([]byte, error) {
	// 	log.Printf("Loading data for key: %s", key)
	// 	return []byte("Value for " + key), nil
	// })

	getter2 := &MyGetter{}

	// 创建一个缓存组
	group := fatcache.NewGroup("exampleGroup", 1024*1024, getter2)

	// 测试缓存加载和命中
	testKey := "myKey"

	// 第一次获取，应该触发加载
	value, err := group.Get(testKey)
	if err != nil {
		log.Fatalf("Error getting value: %v", err)
	}
	fmt.Printf("First fetch: key=%s, value=%s\n", testKey, value.String())

	// 第二次获取，应该命中缓存
	value, err = group.Get(testKey)
	if err != nil {
		log.Fatalf("Error getting value: %v", err)
	}
	fmt.Printf("Second fetch (cache hit): key=%s, value=%s\n", testKey, value.String())
}
