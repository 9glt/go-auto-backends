package autobackends

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

type publisher interface {
	Publish(msg string)
}

type subscriber interface {
	Subscribe(func(string))
}

type PubSub interface {
	publisher
	subscriber
}

func New(pubsub PubSub, me string, weight int64) *Backends {
	t := new(Backends)
	t.pubsub = pubsub
	t.b = make(map[string]*backend)
	t.mu = &sync.RWMutex{}
	t.ow = make(map[string]string)
	t.subscribe()
	t.me = me
	t.start = time.Now().UnixNano()
	t.weight = weight
	return t
}

type backend struct {
	Addr   string
	Seen   time.Time
	Start  int64
	Weight int64
}

func (b *backend) ping() {
	b.Seen = time.Now()
}

func (b *backend) alive() bool {
	if time.Now().Sub(b.Seen).Seconds() > 3 {
		return false
	}
	return true
}

func (b *backend) reset() {
	b.Start = time.Now().UnixNano()
}

type Backends struct {
	pubsub PubSub

	mu     *sync.RWMutex
	b      map[string]*backend
	bl     []*backend
	ow     map[string]string
	me     string
	start  int64
	weight int64
}

func (b *Backends) Cmd(cmd, node, key, value string) {
	b.pubsub.Publish(cmd + " " + node + " " + key + " " + value)
}

func (b *Backends) AddRoute(src, dst string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.ow[src] = dst
}

func (b *Backends) RemoveRoute(src string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	_, exists := b.ow[src]
	if exists {
		delete(b.ow, src)
	}
}

func (b *Backends) get() *backend {
	var be *backend
	for _, bl := range b.bl {
		if bl.alive() && bl.Addr != b.me {
			if be == nil {
				be = bl
				continue
			}
			if be.Weight > bl.Weight {
				be = bl
			}
		}
	}
	return be
}

func (b *Backends) Get() string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	bbe := b.get()
	if bbe == nil {
		return "no alive backends"
	}

	addr, exists := b.ow[bbe.Addr]
	if exists {
		be, ok := b.b[addr]
		if ok {
			if be.alive() && be.Addr != b.me {
				return addr
			}
		}
	}
	return bbe.Addr
}

func (b *Backends) Start(addr string, d int) {
	go func() {
		for {
			b.pubsub.Publish(fmt.Sprintf("alive %s %d %d", addr, b.start, b.weight))
			time.Sleep(time.Duration(d) * time.Second)
		}
	}()
}

func (b *Backends) worker() {
	b.mu.Lock()
	defer b.mu.Unlock()
	ll := []*backend{}
	for _, be := range b.bl {
		if !be.alive() {
			delete(b.b, be.Addr)
		}
		ll = append(ll, be)
	}
	b.bl = ll
}

func (b *Backends) subscribe() {
	b.pubsub.Subscribe(func(s string) {
		b.ping(s)
	})
}

func (b *Backends) ping(s string) {
	l := strings.SplitN(s, " ", 4)
	switch l[0] {
	case "alive":
		if len(l) < 4 {
			log.Printf("alive command invalid: %v", s)
			return
		}
		b.mu.Lock()
		be, ok := b.b[l[1]]
		if !ok {
			i, _ := strconv.Atoi(l[2])
			w, _ := strconv.Atoi(l[3])
			_backend := backend{l[1], time.Now(), int64(i), int64(w)}
			b.b[l[1]] = &_backend
			b.bl = append(b.bl, &_backend)
			be = &_backend
		}
		be.ping()
		b.mu.Unlock()
		return
	case "route-add":
		if len(l) < 4 {
			log.Printf("route-add command invalid: %v", s)
			return
		}
		node, src, dst := l[1], l[2], l[3]
		if node != b.me {
			return
		}
		b.AddRoute(src, dst)
		return
	case "route-rm":
		if len(l) < 3 {
			log.Printf("route-rm command invalid: %v", s)
			return
		}
		node, src := l[1], l[2]
		if node != b.me {
			return
		}
		b.RemoveRoute(src)
		return
	}

}
