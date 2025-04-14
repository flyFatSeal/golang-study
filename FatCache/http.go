package fatcache

import (
	"fatcache/consisitenthash"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

const defaultBasePath = "/_fatcache/"

type httpGetter struct {
	base string
}

func (h httpGetter) Get(group string, key string) ([]byte, error) {
	// 构造请求 URL

	url := fmt.Sprintf("http://%s%s%s/%s", h.base, defaultBasePath, group, key)
	// 发起 GET 请求
	fmt.Println("url:", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 检查 HTTP 响应状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", resp.Status)
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	fmt.Println("httpGetter body:", url)
	return body, nil
}

type HTTPPool struct {
	self        string
	mu          sync.Mutex
	basePath    string
	peers       *consisitenthash.Map
	httpGetters map[string]*httpGetter
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:        self,
		basePath:    defaultBasePath,
		peers:       consisitenthash.New(100, nil),
		httpGetters: make(map[string]*httpGetter, 0),
	}
}

func (h *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 检查路径是否以 basePath 开头
	if !strings.HasPrefix(r.URL.Path, h.basePath) {
		http.Error(w, "unexpected path", http.StatusBadRequest)
		return
	}

	// 解析 key 和 group
	parts := strings.SplitN(r.URL.Path[len(h.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	// 获取缓存组
	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	// 获取缓存数据
	value, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回缓存数据
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(value.ByteSlice())
}

func (h *HTTPPool) Set(peers ...string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.peers.Add(peers...)

	for _, peer := range peers {
		h.httpGetters[peer] = &httpGetter{
			base: peer,
		}
	}
}

func (h *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	peer := h.peers.Get(key)

	// 如果选中的节点是自身，直接返回 nil
	if peer == h.self {
		return nil, false
	}

	if v, ok := h.httpGetters[peer]; ok {
		fmt.Println("Pick peer", peer)
		return v, true
	}
	return nil, false
}
