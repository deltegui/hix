package hx

func If(condition Gettable[bool], child INode) INode {
	var element INode = nil
	if condition.Get() {
		element = child
	}
	return element
}

func Each(src Gettable[[]string], renderOne func(index int, value string) INode) INode {
	list := src.Get()
	result := make([]INode, len(list))
	for i, element := range list {
		result[i] = renderOne(i, element)
	}
	return Span().Body(result...)
}
