package hx

import (
	"fmt"
	"sync"
)

var (
	canDeclareEffects       bool = true
	canDeclareEffectsMutext sync.Mutex
)

func setCanDelcareEffects(b bool) {
	canDeclareEffectsMutext.Lock()
	canDeclareEffects = b
	canDeclareEffectsMutext.Unlock()
}

func getCanDelcareEffects() bool {
	canDeclareEffectsMutext.Lock()
	b := canDeclareEffects
	canDeclareEffectsMutext.Unlock()
	return b
}

func invalidateEffects(handle func()) {
	setCanDelcareEffects(false)
	handle()
	setCanDelcareEffects(true)
}

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
		currentEffect.cleanUps = append(currentEffect.cleanUps, func() {
			delete(signal.subscribers, currentEffect)
		})
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
	childs      []*Effect
	cleanUps    []func()
}

func EffectFunc(fn func()) *Effect {
	if !getCanDelcareEffects() {
		fmt.Println("Warning: You cannot declare Effects inside a event handler. This event will be omitted")
		fn()
		return nil
	}

	e := &Effect{
		fn:       fn,
		childs:   make([]*Effect, 0),
		cleanUps: make([]func(), 0),
	}
	if currentEffect != nil {
		currentEffect.childs = append(currentEffect.childs, e)
	}
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
