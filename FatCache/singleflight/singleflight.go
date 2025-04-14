package singleflight

import "sync"

type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

type Group struct {
	mu    sync.Mutex
	stash map[string]*call
}

func (g *Group) Do(key string, fn func() (value interface{}, err error)) (value interface{}, err error) {
	g.mu.Lock()

	if g.stash == nil {
		g.stash = make(map[string]*call)
	}

	if c, ok := g.stash[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}

	c := new(call)
	c.wg.Add(1)
	g.stash[key] = c
	c.val, c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	delete(g.stash, key)
	g.mu.Unlock()

	return c.val, c.err

}
