package fatcache

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

type String string

func (d String) Len() int {
	return len(d)
}

func TestNewGroup(t *testing.T) {
	getter := GetterFunc(func(key string) ([]byte, error) {
		return []byte("Value for " + key), nil
	})

	group := NewGroup("testGroup", 1024, getter)
	if group == nil {
		t.Fatalf("NewGroup returned nil")
	}
	if group.name != "testGroup" {
		t.Fatalf("Group name mismatch, got %s", group.name)
	}
}

func TestGet(t *testing.T) {
	getter := GetterFunc(func(key string) ([]byte, error) {
		return []byte("Value for " + key), nil
	})

	group := NewGroup("testGroup", 1024, getter)

	// 第一次获取，应该触发加载
	value, err := group.Get("myKey")
	if err != nil {
		t.Fatalf("Error getting value: %v", err)
	}
	if value.String() != "Value for myKey" {
		t.Fatalf("Unexpected value, got %s", value.String())
	}

	// 第二次获取，应该命中缓存
	value, err = group.Get("myKey")
	if err != nil {
		t.Fatalf("Error getting value: %v", err)
	}
	if value.String() != "Value for myKey" {
		t.Fatalf("Unexpected value, got %s", value.String())
	}
}
func TestCacheEviction(t *testing.T) {
	getter := GetterFunc(func(key string) ([]byte, error) {
		return []byte("Value for " + key), nil
	})

	// 缓存容量限制为 20 字节（包括键和值）
	group := NewGroup("testGroup", 20, getter)

	group.Get("key1") // 加载 "key1" + "Value for key1" (总字节数 = len("key1") + len("Value for key1"))
	group.Get("key2") // 加载 "key2" + "Value for key2"

	// "key1" 应该被淘汰
	if _, ok := group.mainCache.get("key1"); ok {
		t.Fatalf("Expected key1 to be evicted")
	}

	// "key2" 应该仍在缓存中
	if _, ok := group.mainCache.get("key2"); !ok {
		t.Fatalf("Expected key2 to be in cache")
	}
}

func TestConcurrentAccess(t *testing.T) {
	getter := GetterFunc(func(key string) ([]byte, error) {
		return []byte("Value for " + key), nil
	})

	group := NewGroup("testGroup", 1024, getter)

	var wg sync.WaitGroup
	for i := 0; i < 200000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key%d", i)
			value, err := group.Get(key)
			if err != nil {
				t.Errorf("Error getting value for key %s: %v", key, err)
			}
			expected := fmt.Sprintf("Value for %s", key)
			if value.String() != expected {
				t.Errorf("Unexpected value for key %s, got %s", key, value.String())
			}
		}(i)
	}
	wg.Wait()
}

func TestExtremeCases(t *testing.T) {
	getter := GetterFunc(func(key string) ([]byte, error) {
		if key == "" {
			return nil, fmt.Errorf("key cannot be empty")
		}
		if key == "nonexistentKey" {
			return nil, fmt.Errorf("key not existing")
		}
		return []byte("Value for " + key), nil
	})

	group := NewGroup("testGroup", 1024, getter)

	// 测试空 key
	_, err := group.Get("")
	if err == nil {
		t.Fatalf("Expected error for empty key")
	}

	// 测试超大值
	largeValue := make([]byte, 2048)
	getter = GetterFunc(func(key string) ([]byte, error) {
		return largeValue, nil
	})
	group = NewGroup("testGroup", 1024, getter)
	_, err = group.Get("largeKey")
	if err == nil {
		t.Fatalf("Expected error for large value exceeding cache capacity")
	}
}

func TestHTTPServer(t *testing.T) {
	// 创建一个 GetterFunc
	getter := GetterFunc(func(key string) ([]byte, error) {
		if key == "" {
			return nil, fmt.Errorf("key cannot be empty")
		}
		if key == "nonexistentKey" {
			return nil, fmt.Errorf("key not existing")
		}
		return []byte("Value for " + key), nil
	})

	// 创建一个缓存组
	group := NewGroup("testGroup", 1024, getter)

	// 创建 HTTP 服务器并挂载到缓存组
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := strings.TrimPrefix(r.URL.Path, "/")
		if key == "" {
			http.Error(w, "key is required", http.StatusBadRequest)
			return
		}
		value, err := group.Get(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(value.ByteSlice())
	}))
	defer server.Close()

	// 测试有效的 key
	resp, err := http.Get(server.URL + "/myKey")
	if err != nil {
		t.Fatalf("Failed to send GET request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	expected := "Value for myKey"
	if string(body) != expected {
		t.Fatalf("Unexpected response body, got %s, expected %s", string(body), expected)
	}

	// 测试空 key
	resp, err = http.Get(server.URL + "/")
	if err != nil {
		t.Fatalf("Failed to send GET request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected status 400 for empty key, got %d", resp.StatusCode)
	}

	// 测试不存在的 key
	resp, err = http.Get(server.URL + "/nonexistentKey")
	if err != nil {
		t.Fatalf("Failed to send GET request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("Expected status 400 for nonexistent key, got %d", resp.StatusCode)
	}
}
