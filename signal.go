package hx

import (
	"sync"
)

var (
	currentEffect *Effect
	mu            sync.Mutex
)

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
	mu.Lock()
	if currentEffect != nil {
		signal.subscribers[currentEffect] = struct{}{}
	}
	mu.Unlock()
	return signal.value
}

func (signal *SignalT[T]) Set(v T) {
	/*if reflect.DeepEqual(signal.value, v) {
		return
	}*/
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
}

func EffectFunc(fn func()) *Effect {
	e := &Effect{
		fn: fn,
	}
	e.run()
	return e
}

func (e *Effect) run() {
	mu.Lock()
	prev := currentEffect
	currentEffect = e
	mu.Unlock()

	e.fn()

	mu.Lock()
	currentEffect = prev
	mu.Unlock()
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
	mu.Lock()
	if currentEffect != nil {
		c.subscribers[currentEffect] = struct{}{}
	}
	mu.Unlock()
	return c.value
}

func (c *ComputedT[T]) notify() {
	for effect := range c.subscribers {
		effect.schedule()
	}
}

type Gettable[T any] interface {
	Get() T
}

type Settable[T any] interface {
	Set(T)
}
