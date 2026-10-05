package gui

import "sync"

// registry tracks the repos open in the window, in tab order. Each repo owns a
// stop channel that its watcher goroutine selects on; closing the repo closes it.
type registry struct {
	mu    sync.Mutex
	order []string
	stops map[string]chan struct{}
}

func newRegistry() *registry {
	return &registry{stops: map[string]chan struct{}{}}
}

// add registers path if it isn't open yet. It returns the repo's stop channel
// and whether this call opened it (false means it was already open).
func (r *registry) add(path string) (stop <-chan struct{}, added bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if ch, ok := r.stops[path]; ok {
		return ch, false
	}
	ch := make(chan struct{})
	r.stops[path] = ch
	r.order = append(r.order, path)
	return ch, true
}

// remove closes path's stop channel and drops it from the tab order.
func (r *registry) remove(path string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	ch, ok := r.stops[path]
	if !ok {
		return false
	}
	close(ch)
	delete(r.stops, path)
	for i, p := range r.order {
		if p == path {
			r.order = append(r.order[:i], r.order[i+1:]...)
			break
		}
	}
	return true
}

func (r *registry) list() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.order...)
}
