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

func New(pubsub PubSub, rootarea, me string, weight int64) *Backends {
	t := new(Backends)
	t.pubsub = pubsub
	t.b = make(map[string]*backend)
	t.ab = make(map[string]map[string]*backend)
	t.ab[rootarea] = make(map[string]*backend)
	t.mu = &sync.RWMutex{}
	t.ow = make(map[string]string)
	t.subscribe()
	t.me = me
	t.area = rootarea
	t.start = time.Now().UnixNano()
	t.weight = weight
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
	ab     map[string]map[string]*backend
	bl     []*backend
	ow     map[string]string
	me     string
	start  int64
	weight int64
	area   string
}

func (b *Backends) get() *backend {
	var be *backend
	for _, bl := range b.bl {
		if bl.alive() && bl.Addr != b.me && bl.Area == b.area {
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

func (b *Backends) Get() (string, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	bbe := b.get()
	if bbe == nil {
		return "", ErrNoAliveBackends
	}

	addr, exists := b.ow[bbe.Addr]
	if exists {
		be, ok := b.ab[b.area][addr]
		if ok {
			if be.alive() && be.Addr != b.me {
				return addr, nil
			}
		}
	}
	return bbe.Addr, nil
}

func (b *Backends) Start(area, addr string, d int, weight int64) {
	go func() {
		for {
			b.pubsub.Publish(fmt.Sprintf("alive %s %s %d %d", area, addr, b.start, weight))
			time.Sleep(time.Duration(d) * time.Second)
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
		_, ok := b.ab[_area]
		if !ok {
			b.ab[_area] = make(map[string]*backend)
		}
		be, ok := b.ab[_area][_addr]
		if !ok {
			i, _ := strconv.Atoi(l[3])
			w, _ := strconv.Atoi(l[4])
			_backend := backend{_addr, _area, time.Now(), int64(i), int64(w)}
			b.ab[_area][_addr] = &_backend

			be = &_backend
			if b.area == _area {
				b.bl = append(b.bl, be)
			}
		}
		be.ping()
		b.mu.Unlock()
		return
	}

}
