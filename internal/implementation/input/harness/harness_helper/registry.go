package harness_helper

import (
	"sync"

	"hexago/internal/helpers/custom_error"
)

type Registry[P any] struct {
	mu          sync.Mutex
	label       string
	max         int
	procs       map[string]P
	uninstalled bool
}

func NewRegistry[P any](label string, max int) *Registry[P] {
	return &Registry[P]{
		label: label,
		max:   max,
		procs: map[string]P{},
	}
}

func (r *Registry[P]) Get(id string) (P, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.lookup(id)
}

func (r *Registry[P]) Take(id string) (P, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	proc, err := r.lookup(id)
	if err == nil {
		delete(r.procs, id)
	}

	return proc, err
}

func (r *Registry[P]) lookup(id string) (P, error) {
	proc, found := r.procs[id]
	if !found {
		var zero P

		return zero, custom_error.Critical("no running agent with id %s", id)
	}

	return proc, nil
}

func (r *Registry[P]) Reserve() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.procs) >= r.max {
		return r.atLimit()
	}

	return nil
}

func (r *Registry[P]) Admit(id string, proc P) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.uninstalled {
		return custom_error.Critical("%s was uninstalled", r.label)
	}

	if len(r.procs) >= r.max {
		return r.atLimit()
	}

	r.procs[id] = proc

	return nil
}

func (r *Registry[P]) atLimit() error {
	return custom_error.Critical("%s is at its limit of %d instances", r.label, r.max)
}

func (r *Registry[P]) Forget(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.procs, id)
}

func (r *Registry[P]) Drain() []P {
	r.mu.Lock()
	defer r.mu.Unlock()

	procs := make([]P, 0, len(r.procs))

	for id, proc := range r.procs {
		procs = append(procs, proc)
		delete(r.procs, id)
	}

	return procs
}

func (r *Registry[P]) Count() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	return len(r.procs)
}

func (r *Registry[P]) MarkUninstalled() {
	r.setUninstalled(true)
}

func (r *Registry[P]) MarkInstalled() {
	r.setUninstalled(false)
}

func (r *Registry[P]) setUninstalled(uninstalled bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.uninstalled = uninstalled
}
