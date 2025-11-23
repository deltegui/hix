package hx

import (
	"fmt"
	"runtime/debug"
	"sync"
)

var (
	currentEffect *Effect
	untrack       bool
	mu            sync.Mutex
)

func accessEffect(action func(*Effect)) {
	mu.Lock()
	action(currentEffect)
	mu.Unlock()
}

type SignalT[T any] struct {
	value       T
	subscribers map[*Effect]struct{}
}

func Signal[T any](initial T) *SignalT[T] {
	return &SignalT[T]{
		value:       initial,
		subscribers: make(map[*Effect]struct{}),
	}
}

func (signal *SignalT[T]) Get() T {
	accessEffect(func(currentEffect *Effect) {
		if currentEffect != nil {
			signal.subscribers[currentEffect] = struct{}{}
			currentEffect.cleanUps = append(currentEffect.cleanUps, func() {
				delete(signal.subscribers, currentEffect)
			})
		} else if !untrack {
			fmt.Printf("Cannot call Signal.Get if there is no effect in %s\n", debug.Stack())
		}
	})
	return signal.value
}

func (signal *SignalT[T]) Set(v T) {
	signal.value = v
	signal.notify()
}

func (signal *SignalT[T]) Update(fn func(T) T) {
	signal.Set(fn(signal.value))
}

func (signal *SignalT[T]) notify() {
	for effect := range signal.subscribers {
		effect.schedule()
	}
}

type Effect struct {
	fn          func()
	isScheduled bool
	childs      []*Effect
	cleanUps    []func()
}

func EffectFunc(fn func()) *Effect {
	e := &Effect{
		fn:       fn,
		childs:   make([]*Effect, 0),
		cleanUps: make([]func(), 0),
	}
	accessEffect(func(currentEffect *Effect) {
		if currentEffect != nil {
			currentEffect.childs = append(currentEffect.childs, e)
		}
	})
	e.run()
	return e
}

func (e *Effect) run() {
	mu.Lock()
	prev := currentEffect
	currentEffect = e
	mu.Unlock()

	e.clean()
	e.fn()

	mu.Lock()
	currentEffect = prev
	mu.Unlock()
}

func (e *Effect) clean() {
	for _, child := range e.childs {
		child.clean()
	}
	e.childs = []*Effect{}

	for _, cleanfn := range e.cleanUps {
		cleanfn()
	}
	e.cleanUps = make([]func(), 0)

}

func (e *Effect) schedule() {
	if e.isScheduled {
		return
	}
	e.isScheduled = true
	e.run()
	e.isScheduled = false
}

type ComputedT[T comparable] struct {
	value           T
	dependentEffect *Effect
	subscribers     map[*Effect]struct{}
}

func Computed[T comparable](fn func() T) *ComputedT[T] {
	c := &ComputedT[T]{
		subscribers: make(map[*Effect]struct{}),
	}

	c.dependentEffect = EffectFunc(func() {
		newVal := fn()
		if newVal != c.value {
			c.value = newVal
			c.notify()
		}
	})

	return c
}

func (c *ComputedT[T]) Get() T {
	accessEffect(func(currentEffect *Effect) {
		if currentEffect != nil {
			c.subscribers[currentEffect] = struct{}{}
		}
	})
	return c.value
}

func (c *ComputedT[T]) notify() {
	for effect := range c.subscribers {
		effect.schedule()
	}
}

func Untrack(fn func()) {
	mu.Lock()
	prev := currentEffect
	untrack = true
	currentEffect = nil
	mu.Unlock()

	fn()

	mu.Lock()
	currentEffect = prev
	untrack = false
	mu.Unlock()
}

func UntrackGet[T any](gettable Gettable[T]) T {
	mu.Lock()
	prev := currentEffect
	untrack = true
	currentEffect = nil
	mu.Unlock()

	value := gettable.Get()

	mu.Lock()
	currentEffect = prev
	untrack = false
	mu.Unlock()

	return value
}

type Gettable[T any] interface {
	Get() T
}

type Settable[T any] interface {
	Set(T)
}
