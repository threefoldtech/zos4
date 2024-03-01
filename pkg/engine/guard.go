package engine

import (
	"sync"
	"sync/atomic"
)

// Guard is a kind of shared lock on a single resource.
// Once a guard is retrieved by a call to AccessGuard.Enter()
// the guard must be "Exited" (defer guard.Exit()) during the time
// the guard is acquired, a guard can be locked and unlocked as many
// time as needed.
type Guard interface {
	sync.Locker
	RLock()
	RUnlock()
	// Exit drops the Guard completely and
	// it's illegal to try to lock it after it has been exited
	Exit()
}

// accessGuard is used internally to
// manage exclusive access to objects
type accessGuard struct {
	id     string
	parent *AccessGuard
	sync.RWMutex
	count atomic.Int32
}

func (g *accessGuard) Exit() {
	g.parent.exit(g.id)
}

type AccessGuard struct {
	guards map[string]*accessGuard
	m      sync.Mutex
}

func NewAccessGuard() *AccessGuard {
	return &AccessGuard{
		guards: map[string]*accessGuard{},
	}
}

func (g *AccessGuard) Enter(id string) Guard {
	g.m.Lock()
	defer g.m.Unlock()

	guard, ok := g.guards[id]
	if !ok {
		guard = &accessGuard{
			id:     id,
			parent: g,
		}
		g.guards[id] = guard
	}
	guard.count.Add(1)
	return guard
}

func (g *AccessGuard) exit(id string) {
	g.m.Lock()
	defer g.m.Unlock()

	guard := g.guards[id]
	n := guard.count.Add(-1)
	if n == 0 {
		delete(g.guards, id)
	} else if n < 0 {
		panic("invalid guard exit")
	}
}