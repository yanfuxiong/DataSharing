package utils

import (
	"log"
	"sync"
)

type CondGroup struct {
	mu      sync.Mutex
	counter int        // current goroutine count
	cond    *sync.Cond // notice  all goroutine done!
}

func NewCondGroup() *CondGroup {
	c := &CondGroup{}
	c.cond = sync.NewCond(&c.mu)
	c.counter = 0
	return c
}

func (g *CondGroup) Add(delta int) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.counter += delta
	if g.counter < 0 {
		log.Printf("CondGroup sync: negative CondWaitGroup counter")
		g.counter = 0
	}
	if g.counter == 0 {
		// awaken  all Wait  when counter is zero
		g.cond.Broadcast()
	}
}

func (g *CondGroup) Done() {
	g.Add(-1)
}

func (g *CondGroup) Wait() {
	g.mu.Lock()
	for g.counter != 0 {
		log.Printf("CondGroup sync: CondWaitGroup waiting")
		g.cond.Wait()
	}
	log.Printf("CondGroup sync: CondWaitGroup done")
	g.mu.Unlock()
}
