package hx

type ValueWrapper[T any] struct {
	inner T
}

func Value[T any](t T) ValueWrapper[T] {
	return ValueWrapper[T]{
		inner: t,
	}
}

func (v ValueWrapper[T]) Get() T {
	return v.inner
}

func (v *ValueWrapper[T]) Set(t T) {
	v.inner = t
}

type singleValue[T any] struct {
	value  T
	status changeStatus
}

func (sv *singleValue[T]) assign(val T, status changeStatus) bool {
	sv.value = val
	sv.status = status
	return true
}

func (sv singleValue[T]) Value() T {
	return sv.value
}
func (sv singleValue[T]) Status() changeStatus {
	return sv.status
}

type diffValue[T comparable] struct {
	value     T
	nextValue T
	status    changeStatus
}

func (sv diffValue[T]) Value() T {
	return sv.nextValue
}
func (sv diffValue[T]) Status() changeStatus {
	return sv.status
}

func (sv *diffValue[T]) reset() {
	sv.nextValue = sv.value
	sv.status = unchanged
}

func (sv *diffValue[T]) assign(val T, status changeStatus) bool {
	if val == sv.value {
		sv.reset()
		return false
	}
	if val == sv.nextValue {
		return true
	}
	sv.nextValue = val
	sv.status = status
	return true
}

func (sv *diffValue[T]) equals(val T) bool {
	return sv.nextValue == val || sv.value == val
}

func (sv *diffValue[T]) tick() {
	sv.value = sv.nextValue
	sv.status = unchanged
}
