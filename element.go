package hx

import (
	"honnef.co/go/js/dom/v2"
)

type Renderer interface {
	ScheduleRender()
	Mark(element *VNode)
}

type EventContext struct {
	Target *VNode
	Event  dom.Event
}

type Event string

const (
	EventClick  Event = "click"
	EventInput  Event = "input"
	EventChange Event = "change"
)

type changeStatus int

const (
	unchanged changeStatus = iota
	changeNew
	changeModified
	changeDeleted
)

type INode interface {
	Id(id string) INode

	Text(t string) INode
	BindText(signal Gettable[string]) INode

	Body(childs ...INode) INode
	BodyList(childs []INode) INode

	Attribute(key, value string) INode
	RemoveAttribute(key string) INode
	Style(key, value string) INode
	RemoveStyle(key string) INode
	Class(classes ...string) INode
	RemoveClass(classes ...string) INode

	On(event Event, handler func(ctx EventContext)) INode
	OnClick(handler func(ctx EventContext)) INode
}

type dirtyFlag int

const flagNumber = 5

const (
	flagStyles dirtyFlag = iota
	flagClasses
	flagAttributes
	flagEventListeners
	flagChildren
)

type VNode struct {
	domElement     dom.Element
	haveDomElement bool

	father *VNode
	status changeStatus

	renderer     Renderer
	haveRenderer bool

	tag   string
	id    diffValue[string]
	text  diffValue[string]
	value diffValue[string]

	styles         map[string]diffValue[string]
	classes        map[string]changeStatus
	attributes     map[string]diffValue[string]
	eventListeners map[Event]singleValue[func(EventContext)]
	children       []*VNode

	dirtyFlags [flagNumber]bool
}

func NewWithRenderer(element dom.Element, renderer Renderer) *VNode {
	VNode := newVNode(element.NodeName())
	VNode.status = unchanged
	VNode.domElement = element
	VNode.haveDomElement = true
	VNode.renderer = renderer
	VNode.haveRenderer = true
	return VNode
}

func New(element dom.Element) *VNode {
	return NewWithRenderer(element, newRenderer(element))
}

func NewFromId(id string) *VNode {
	mountPoint := dom.GetWindow().Document().GetElementByID(id)
	return New(mountPoint)
}

func NewFromIdWithRenderer(id string, renderer Renderer) *VNode {
	mountPoint := dom.GetWindow().Document().GetElementByID(id)
	return NewWithRenderer(mountPoint, renderer)
}

func NewWithoutMount(tag string, renderer Renderer) *VNode {
	VNode := newVNode(tag)
	VNode.setRenderer(renderer, true)
	return VNode
}

func newVNode(tag string) *VNode {
	return &VNode{
		status: changeNew,

		father: nil,

		tag:            tag,
		id:             diffValue[string]{},
		text:           diffValue[string]{},
		value:          diffValue[string]{},
		styles:         map[string]diffValue[string]{},
		classes:        map[string]changeStatus{},
		attributes:     map[string]diffValue[string]{},
		eventListeners: map[Event]singleValue[func(EventContext)]{},
		children:       []*VNode{},

		dirtyFlags: [flagNumber]bool{false},
	}
}

func (element *VNode) mark() {
	if element.haveRenderer {
		element.renderer.Mark(element)
	}
}

func (element *VNode) clearDirty() {
	element.dirtyFlags = [flagNumber]bool{false}
}

func (element *VNode) setDirty(flag dirtyFlag) {
	element.dirtyFlags[flag] = true
	element.mark()
}

func (element *VNode) isDirty(flag dirtyFlag) bool {
	return element.dirtyFlags[flag]
}

func (element *VNode) setRenderer(renderer Renderer, haveRenderer bool) {
	element.renderer = renderer
	element.haveRenderer = haveRenderer
	for index := range element.children {
		element.children[index].setRenderer(renderer, haveRenderer)
		element.children[index].haveRenderer = haveRenderer
	}
	element.setDirty(flagChildren)
}

func (element *VNode) scheludeRender() {
	if element.haveRenderer {
		element.renderer.ScheduleRender()
	}
}

func (element *VNode) Body(childs ...INode) INode {
	return element.BodyList(childs)
}

func (element *VNode) BodyList(childs []INode) INode {
	for _, child := range element.children {
		child.status = changeDeleted
	}

	for _, child := range childs {
		realNode := asVNode(child)
		if child == nil {
			continue
		}
		realNode.status = changeNew
		realNode.setRenderer(element.renderer, element.haveRenderer)
		realNode.father = element
		element.children = append(element.children, realNode)
	}

	element.setDirty(flagChildren)
	element.scheludeRender()

	return element
}

func (element *VNode) On(event Event, handler func(ctx EventContext)) INode {
	listener, ok := element.eventListeners[event]
	if !ok {
		element.eventListeners[event] = singleValue[func(EventContext)]{
			value: func(ctx EventContext) {
				handler(ctx)
				element.scheludeRender()
			},
			status: changeNew,
		}
	}
	listener.assign(handler, changeModified)

	element.setDirty(flagEventListeners)
	return element
}

func (element *VNode) OnClick(handler func(ctx EventContext)) INode {
	return element.On(EventClick, handler)
}

func (element *VNode) Text(t string) INode {
	if element.text.assign(t, changeModified) {
		element.mark()
	}
	return element
}

func (element *VNode) BindText(signal Gettable[string]) INode {
	element.Text(signal.Get())
	element.scheludeRender()
	return element
}

func (element *VNode) Class(classes ...string) INode {
	for _, create := range classes {
		_, ok := element.classes[create]
		if !ok {
			element.classes[create] = changeNew
			element.setDirty(flagClasses)
		}
	}
	return element
}

func (element *VNode) RemoveClass(classes ...string) INode {
	for _, remove := range classes {
		_, ok := element.classes[remove]
		if ok {
			element.classes[remove] = changeDeleted
			element.setDirty(flagClasses)
		}
	}
	return element
}

func (element *VNode) Attribute(key, value string) INode {
	oldValue, ok := element.attributes[key]
	if ok && oldValue.equals(value) {
		return element
	}

	var status changeStatus
	if ok {
		status = changeModified
	} else {
		status = changeNew
	}
	oldValue.assign(value, status)
	element.attributes[key] = oldValue

	element.setDirty(flagAttributes)
	return element
}

func (element *VNode) RemoveAttribute(key string) INode {
	current, ok := element.attributes[key]
	if !ok {
		return element
	}
	current.status = changeDeleted
	element.attributes[key] = current
	element.setDirty(flagAttributes)
	return element
}

func (element *VNode) Style(key, value string) INode {
	oldValue, ok := element.styles[key]
	if ok && oldValue.equals(value) {
		return element
	}

	var status changeStatus
	if ok {
		status = changeModified
	} else {
		status = changeNew
	}
	oldValue.assign(value, status)
	element.styles[key] = oldValue

	element.setDirty(flagStyles)
	return element
}

func (element *VNode) RemoveStyle(key string) INode {
	current, ok := element.styles[key]
	if !ok {
		return element
	}
	current.status = changeDeleted
	element.styles[key] = current
	element.setDirty(flagStyles)
	return element
}

func (element *VNode) Id(id string) INode {
	if element.id.value == id {
		return element
	}
	if element.id.assign(id, changeModified) {
		element.mark()
	}
	return element
}

func (element *InputVNode) BindOnChange(signal Settable[string]) *InputVNode {
	element.On(EventChange, func(ctx EventContext) {
		input := ctx.Event.Target().(*dom.HTMLInputElement)
		signal.Set(input.Value())
	})
	return element
}

func (e *VNode) Src(v string) INode { return e.Attribute("src", v) }

type AVNode struct {
	VNode
}

func newA(element *VNode) *AVNode {
	return &AVNode{
		*element,
	}
}

func (e *AVNode) Href(v string) *AVNode {
	e.Attribute("href", v)
	return e
}

type InputVNode struct {
	VNode
}

func asInput(node *VNode) *InputVNode {
	return &InputVNode{*node}
}

func (element *InputVNode) Value(t string) INode {
	if element.value.assign(t, changeModified) {
		element.mark()
	}
	return element
}

func (element *InputVNode) BindOnInput(signal Settable[string]) *InputVNode {
	element.On(EventInput, func(ctx EventContext) {
		input := ctx.Event.Target().(*dom.HTMLInputElement)
		signal.Set(input.Value())
		element.scheludeRender()
	})
	return element
}

func (element *InputVNode) BindValue(signal Gettable[string]) *InputVNode {
	element.Value(signal.Get())
	element.scheludeRender()
	return element
}

func (e *InputVNode) Placeholder(v string) *InputVNode {
	return e.Attribute("placeholder", v).(*InputVNode)
}

func (e *InputVNode) Type(v string) *InputVNode {
	return e.Attribute("type", v).(*InputVNode)
}

type TextAreaNode struct {
	VNode
}

func asTextArea(node *VNode) *TextAreaNode {
	return &TextAreaNode{
		*node,
	}
}

func (element *TextAreaNode) Value(t string) INode {
	if element.value.assign(t, changeModified) {
		element.mark()
	}
	return element
}

func (element *TextAreaNode) BindValue(signal Gettable[string]) *TextAreaNode {
	v := signal.Get()
	element.Value(v)
	element.Text(v)
	element.scheludeRender()
	return element
}

const noopIdNode string = "noop"

type NoopNode struct {
	VNode
}

func asNoop(node *VNode) *NoopNode {
	return &NoopNode{
		*node,
	}
}

func (nop *NoopNode) Id(id string) INode {
	return nop
}
func (nop *NoopNode) Text(t string) INode {
	return nop
}
func (nop *NoopNode) BindText(signal Gettable[string]) INode {
	return nop
}
func (nop *NoopNode) Body(childs ...INode) INode {
	nop.VNode.BodyList(childs)
	return nop
}
func (nop *NoopNode) BodyList(childs []INode) INode {
	nop.VNode.BodyList(childs)
	return nop
}
func (nop *NoopNode) Attribute(key, value string) INode {
	return nop
}
func (nop *NoopNode) RemoveAttribute(key string) INode {
	return nop
}
func (nop *NoopNode) Style(key, value string) INode {
	return nop
}
func (nop *NoopNode) RemoveStyle(key string) INode {
	return nop
}
func (nop *NoopNode) Class(classes ...string) INode {
	return nop
}
func (nop *NoopNode) RemoveClass(classes ...string) INode {
	return nop
}
func (nop *NoopNode) On(event Event, handler func(ctx EventContext)) INode {
	return nop
}
func (nop *NoopNode) OnClick(handler func(ctx EventContext)) INode {
	return nop
}

func asVNode(i INode) *VNode {
	switch n := i.(type) {
	case *VNode:
		return n
	case *InputVNode:
		return &n.VNode
	case *AVNode:
		return &n.VNode
	case *NoopNode:
		return &n.VNode
	case *TextAreaNode:
		return &n.VNode
	default:
		return nil
	}
}

func H1() *VNode { return newVNode("H1") }
func H2() *VNode { return newVNode("H2") }
func H3() *VNode { return newVNode("H3") }
func H4() *VNode { return newVNode("H4") }
func H5() *VNode { return newVNode("H5") }
func H6() *VNode { return newVNode("H6") }

func Div() *VNode { return newVNode("DIV") }

func P() *VNode { return newVNode("P") }

func Button() *VNode { return newVNode("BUTTON") }

func A() *AVNode              { return newA(newVNode("A")) }
func Span() *VNode            { return newVNode("SPAN") }
func Strong() *VNode          { return newVNode("STRONG") }
func Em() *VNode              { return newVNode("EM") }
func Small() *VNode           { return newVNode("SMALL") }
func Img() *VNode             { return newVNode("IMG") }
func Input() *InputVNode      { return asInput(newVNode("INPUT")) }
func Label() *VNode           { return newVNode("LABEL") }
func Form() *VNode            { return newVNode("FORM") }
func Select() *VNode          { return newVNode("SELECT") }
func Option() *VNode          { return newVNode("OPTION") }
func TextArea() *TextAreaNode { return asTextArea(newVNode("TEXTAREA")) }

func Ul() *VNode { return newVNode("UL") }
func Ol() *VNode { return newVNode("OL") }
func Li() *VNode { return newVNode("LI") }

func Table() *VNode { return newVNode("TABLE") }
func THead() *VNode { return newVNode("THEAD") }
func TBody() *VNode { return newVNode("TBODY") }
func TFoot() *VNode { return newVNode("TFOOT") }
func Tr() *VNode    { return newVNode("TR") }
func Th() *VNode    { return newVNode("TH") }
func Td() *VNode    { return newVNode("TD") }

func Nav() *VNode     { return newVNode("NAV") }
func Header() *VNode  { return newVNode("HEADER") }
func Footer() *VNode  { return newVNode("FOOTER") }
func Main() *VNode    { return newVNode("MAIN") }
func Section() *VNode { return newVNode("SECTION") }
func Article() *VNode { return newVNode("ARTICLE") }
func Aside() *VNode   { return newVNode("ASIDE") }

func Video() *VNode  { return newVNode("VIDEO") }
func Audio() *VNode  { return newVNode("AUDIO") }
func Source() *VNode { return newVNode("SOURCE") }

func Canvas() *VNode { return newVNode("CANVAS") }
func Svg() *VNode    { return newVNode("SVG") }
func Path() *VNode   { return newVNode("PATH") }

func Br() *VNode { return newVNode("BR") }
func Hr() *VNode { return newVNode("HR") }

func Code() *VNode { return newVNode("CODE") }
func Pre() *VNode  { return newVNode("PRE") }

func Noop() *NoopNode { return asNoop(newVNode(noopIdNode)) }
