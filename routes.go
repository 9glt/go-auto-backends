package autobackends

import (
	"sort"
	"sync"
)

type Node struct {
	Addr   string
	Weight int
}

// NewTable crates new routing table
func NewTable() R {
	return R{
		make(map[string][]Node),
		&sync.RWMutex{},
	}
}

type R struct {
	m  map[string][]Node
	mu *sync.RWMutex
}

func (r *R) Add(m, p string, weight int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, exists := r.m[m]
	if !exists {
		r.m[m] = []Node{}
	}
	// if inList(r.m[m], p) {
	for k, v := range r.m[m] {
		if v.Addr == p {
			r.m[m][k].Weight = weight
			return
		}
	}
	// }
	r.m[m] = append(r.m[m], Node{p, weight})
	s := SortBy(r.m[m])
	sort.Sort(s)
	r.m[m] = s
}

func (r *R) Remove(m, p string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, exists := r.m[m]
	if !exists {
		return
	}
	if !inList(r.m[m], p) {
		return
	}
	l := []Node{}
	for _, v := range r.m[m] {
		if v.Addr == p {
			continue
		}
		l = append(l, v)
	}
	r.m[m] = l
}

func (r *R) List(me string, root *string) []Node {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := SortBy(r.list(me, root))
	sort.Sort(list)
	return list
}
func (r *R) list(me string, root *string) []Node {
	list, exists := r.m[me]
	if !exists {
		return []Node{}
	}
	l := []Node{}
	for _, k := range list {
		l = append(l, k)
		if root != nil && k.Addr == *root {
			k.Weight = 100
		}
		for _, v := range r.list(k.Addr, root) {
			v.Weight += k.Weight
			l = append(l, v)
		}
	}
	return l
}

func inList(l []Node, in string) bool {
	for _, v := range l {
		if v.Addr == in {
			return true
		}
	}
	return false
}

type SortBy []Node

func (a SortBy) Len() int           { return len(a) }
func (a SortBy) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SortBy) Less(i, j int) bool { return a[i].Weight < a[j].Weight }
