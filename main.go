package main

import (
	"fatcache"
	"fmt"
	"log"
)

func main() {
	// 模拟一个 Getter 函数，从数据源加载数据
	getter := fatcache.GetterFunc(func(key string) ([]byte, error) {
		log.Printf("Loading data for key: %s", key)
		return []byte("Value for " + key), nil
	})

	// 创建一个缓存组
	group := fatcache.NewGroup("exampleGroup", 1024*1024, getter)

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
