package autobackends

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	ErrNoAliveBackends = errors.New("no alive backends")
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

func New(pubsub PubSub, routes []Node, me string) *Backends {
	t := new(Backends)
	t.pubsub = pubsub
	t.b = make(map[string]*backend)
	t.ab = make(map[string]map[string]*backend)
	t.mu = &sync.RWMutex{}
	t.ow = make(map[string]string)
	t.subscribe()
	t.me = me
	t.routes = routes
	t.start = time.Now().UnixNano()

	t.Start(me, 300)
	return t
}

type backend struct {
	Addr   string
	Area   string
	Seen   time.Time
	Start  int64
	Weight int64
}

func (b *backend) ping() {
	b.Seen = time.Now()
}

func (b *backend) alive() bool {
	if time.Now().Sub(b.Seen).Seconds() > 1 {
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
	ab     map[string]map[string]*backend
	ow     map[string]string
	me     string
	start  int64
	weight int64
	area   string
	routes []Node
}

func (b *Backends) get2() *backend {
	for _, node := range b.routes {
		if b.alive(node.Addr) && node.Addr != b.me {
			return b.b[node.Addr]
		}
	}
	return nil
}

func (b *Backends) alive(host string) bool {
	be, exists := b.b[host]
	if !exists {
		return false
	}
	return be.alive()
}

type Route struct {
	Addr   string
	Weight int64
	Alive  bool
}

func (b *Backends) Routes() []Node {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.routes
}

func (b *Backends) SetRoutes(r []Node) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.routes = r
}

func (b *Backends) Get() (string, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	bbe := b.get2()
	if bbe == nil {
		return "", ErrNoAliveBackends
	}

	return bbe.Addr, nil
}

func (b *Backends) Start(addr string, d int) {
	go func() {
		for {
			b.pubsub.Publish(fmt.Sprintf("alive %s %s %d %d", addr, addr, b.start, 10))
			time.Sleep(time.Duration(d) * time.Millisecond)
		}
	}()
}

func (b *Backends) subscribe() {
	b.pubsub.Subscribe(func(s string) {
		b.ping(s)
	})
}

func (b *Backends) ping(s string) {
	l := strings.Split(s, " ")
	switch l[0] {
	case "alive":
		if len(l) < 5 {
			log.Printf("alive command invalid: %v", s)
			return
		}
		_area, _addr := l[1], l[2]
		b.mu.Lock()
		defer b.mu.Unlock()
		tbe, exists := b.b[_addr]
		if !exists {
			i, _ := strconv.Atoi(l[3])
			w, _ := strconv.Atoi(l[4])
			b.b[_addr] = &backend{_addr, _area, time.Now(), int64(i), int64(w)}
			tbe = b.b[_addr]
		}
		tbe.ping()
		return
	}

}
