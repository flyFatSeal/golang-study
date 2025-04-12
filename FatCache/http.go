package fatcache

import (
	"fatcache/consisitenthash"
	"net/http"
	"strings"
	"sync"
)

const defaultBasePath = "/_fatcache/"

type httpGetter struct {
}

type PeerGetter struct {
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

func (h *HTTPPool) set(peers ...string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.peers.Add(peers...)
}

// func (h *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
// 	peer := h.peers.Get(key)

// 	if v, ok := *h.httpGetters[peer]; ok {
// 		return v, true
// 	}

// 	return nil, false
// }
