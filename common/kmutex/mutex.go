package kmutex

import "sync"

type Kmutex[T comparable] struct {
	l sync.Locker
	s map[T]*klock
}

type klock struct {
	cond *sync.Cond
	ref  uint64
}

func New[T comparable]() *Kmutex[T] {
	l := sync.Mutex{}
	return &Kmutex[T]{
		l: &l,
		s: make(map[T]*klock),
	}
}

func (km *Kmutex[T]) Unlock(key T) {
	km.l.Lock()
	defer km.l.Unlock()
	kl, ok := km.s[key]
	if !ok || kl.ref == 0 {
		panic("unlock of unlocked kmutex")
	}
	kl.ref--
	if kl.ref == 0 {
		delete(km.s, key)
		return
	}
	kl.cond.Signal()
}

func (km *Kmutex[T]) Lock(key T) {
	km.l.Lock()
	defer km.l.Unlock()
	for {
		kl, ok := km.s[key]
		if !ok {
			km.s[key] = &klock{
				cond: sync.NewCond(km.l),
				ref:  1,
			}
			return
		}
		kl.ref++
		kl.cond.Wait()
		return
	}
}
