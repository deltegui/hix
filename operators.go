package hx

func If(condition Gettable[bool], child INode) INode {
	var element INode = nil
	if condition.Get() {
		element = child
	}
	return element
}

func Each[T any](src Gettable[[]T], renderOne func(index int, value T) INode) INode {
	list := src.Get()
	result := make([]INode, len(list))
	for i, element := range list {
		result[i] = renderOne(i, element)
	}
	return Noop().Body(result...)
}

func EachMap[K comparable, T any](src Gettable[map[K]T], renderOne func(index K, value T) INode) INode {
	list := src.Get()
	result := make([]INode, len(list))
	i := 0
	for key, element := range list {
		result[i] = renderOne(key, element)
		i++
	}
	return Noop().Body(result...)
}
