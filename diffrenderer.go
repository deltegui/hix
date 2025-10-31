package hx

import (
	"log"
	"strings"
	"syscall/js"

	"honnef.co/go/js/dom/v2"
)

type DiffRenderer struct {
	mountpoint  dom.Element
	markNodes   map[*VNode]struct{}
	scheduled   bool
	rafCallback js.Func
}

func newRenderer(element dom.Element) *DiffRenderer {
	r := &DiffRenderer{
		mountpoint: element,
		markNodes:  map[*VNode]struct{}{},
		scheduled:  false,
	}
	r.createRaf()
	return r
}

func (r *DiffRenderer) Mark(element *VNode) {
	r.markNodes[element] = struct{}{}
}

func (r *DiffRenderer) createRaf() {
	r.rafCallback = js.FuncOf(func(this js.Value, args []js.Value) any {
		r.render()
		r.scheduled = false
		return nil
	})
}

func (r *DiffRenderer) ScheduleRender() {
	if !r.scheduled {
		r.scheduled = true
		js.Global().Call("requestAnimationFrame", r.rafCallback)
	}
}

func (renderer DiffRenderer) render() {
	rootLCA := renderer.GetMarkedCommonAncestor()
	if rootLCA != nil {
		renderer.syncNodes(rootLCA)
	}
	for node := range renderer.markNodes {
		node.render()
		delete(renderer.markNodes, node)
	}
}

func (renderer DiffRenderer) GetMarkedCommonAncestor() *VNode {
	if len(renderer.markNodes) == 0 {
		return nil
	}

	var path []*VNode
	isFirstMark := true

	for markedNode := range renderer.markNodes {
		if isFirstMark {
			path = renderer.generatePathToRoot(markedNode)
			isFirstMark = false
			continue
		}

		path = renderer.generateCommonPathToRoot(path, markedNode)
	}
	return path[0]
}

func (renderer DiffRenderer) generatePathToRoot(source *VNode) []*VNode {
	path := []*VNode{source}
	parent := source.father
	for parent != nil {
		path = append(path, parent)
		parent = parent.father
	}
	return path
}

func (renderer DiffRenderer) generateCommonPathToRoot(currentPath []*VNode, markedNode *VNode) []*VNode {
	parent := markedNode.father
	for parent != nil {
		for pathNodeIndex, pathNode := range currentPath {
			if pathNode == parent {
				return currentPath[pathNodeIndex:]
			}
		}
		parent = parent.father
	}
	return currentPath
}

func (renderer *DiffRenderer) syncNoopNode(element *VNode) bool {
	if element.status == changeDeleted {
		parent := element.father
		if parent != nil {
			for _, child := range element.children {
				if child.haveDomElement {
					parent.domElement.RemoveChild(child.domElement)
				}
			}
			delete(renderer.markNodes, element)
			element.father = nil
		}
		return false
	}

	for index, child := range element.children {
		if child == nil {
			continue
		}
		keepChild := renderer.syncNodes(child)
		if !keepChild {
			element.father.children = removeListItem(element.father.children, index)
		}
	}

	if element.status == changeNew {
		element.status = unchanged
		parent := element.father
		if parent != nil && parent.haveDomElement {
			for _, child := range element.children {
				parent.domElement.AppendChild(child.domElement)
			}
		}
	}

	return true
}

func (renderer *DiffRenderer) syncNodes(element *VNode) bool {
	if element.tag == noopIdNode {
		return renderer.syncNoopNode(element)
	}

	if element.status == changeDeleted {
		parent := element.father
		if parent != nil {
			if element.haveDomElement {
				parent.domElement.RemoveChild(element.domElement)
			}
			delete(renderer.markNodes, element)
			element.father = nil
		}
		return false
	}
	if element.status == changeNew {
		if !element.haveDomElement && len(element.tag) != 0 {
			domNode := dom.GetWindow().Document().CreateElement(element.tag)
			element.domElement = domNode
			element.haveDomElement = true
		}
		element.status = unchanged

		parent := element.father
		if parent != nil && parent.haveDomElement {
			parent.domElement.AppendChild(element.domElement)
		}
	}

	for index, child := range element.children {
		if child == nil {
			continue
		}
		keepChild := renderer.syncNodes(child)
		if !keepChild {
			element.children = removeListItem(element.children, index)
		}
	}

	return true
}

func (element *VNode) render() {
	if element.text.status != unchanged {
		element.updateText()
	}
	if element.value.status != unchanged {
		element.updateValue()
	}
	if element.isDirty(flagEventListeners) {
		element.updateEventListeners()
	}
	if element.isDirty(flagClasses) {
		element.updateClasses()
	}
	if element.isDirty(flagAttributes) {
		element.updateAttributes()
	}
	if element.isDirty(flagStyles) {
		element.updateStyles()
	}
	if element.id.status != unchanged {
		element.updateId()
	}

	if element.isDirty(flagChildren) {
		element.updateChildren()
	}

	element.clearDirty()
}

func (element *VNode) updateId() {
	element.domElement.SetID(element.id.Value())
	element.id.status = unchanged
}

func (element *VNode) updateAttributes() {
	for attribute, value := range element.attributes {
		switch value.status {
		case changeModified, changeNew:
			element.domElement.SetAttribute(attribute, value.Value())
		case changeDeleted:
			element.domElement.RemoveAttribute(attribute)
		}
		value.tick()
	}
}

func (element *VNode) updateStyles() {
	stylesString := strings.Builder{}
	for style, value := range element.styles {
		if value.status != changeDeleted {
			stylesString.WriteString(style)
			stylesString.WriteString(":")
			stylesString.WriteString(value.Value())
			stylesString.WriteString("; ")
		}
		value.tick()
	}
	if stylesString.Len() > 0 {
		element.domElement.SetAttribute("style", stylesString.String())
	} else {
		element.domElement.RemoveAttribute("style")
	}
}

func (element *VNode) updateClasses() {
	for class, status := range element.classes {
		if strings.Contains(class, " ") {
			log.Printf("Warning: ignoring class with spaces: %s", class)
			continue
		}
		switch status {
		case changeDeleted:
			element.domElement.Class().Remove(class)
		case changeNew:
			element.domElement.Class().Add(class)
		}
		element.classes[class] = unchanged
	}
}

func (element *VNode) updateEventListeners() {
	for event, listener := range element.eventListeners {
		switch listener.status {
		case changeNew:
			currentListener := listener
			element.domElement.AddEventListener(string(event), false, func(e dom.Event) {
				currentListener.value(EventContext{
					Target: element,
					Event:  e,
				})
			})
		}
	}
}

func (element *VNode) updateChildren() {
	for index, child := range element.children {
		if child == nil {
			continue
		}
		element.children[index].render()
	}
}

func (element *VNode) updateText() {
	element.domElement.SetTextContent(element.text.Value())
	element.text.tick()
}

func (element *VNode) updateValue() {
	jsVal := element.domElement.Underlying()
	val := element.value.Value()
	jsVal.Set("value", val)
	element.value.tick()
}
