package hx

import (
	"html"
	"strings"
)

var voidElements = map[string]bool{
	"AREA": true, "BASE": true, "BR": true, "COL": true,
	"EMBED": true, "HR": true, "IMG": true, "INPUT": true,
	"LINK": true, "META": true, "PARAM": true, "SOURCE": true,
	"TRACK": true, "WBR": true,
}

type StringRenderer struct {
	buff strings.Builder
}

func (ssr *StringRenderer) ScheduleRender()     {}
func (ssr *StringRenderer) Mark(element *VNode) {}

func (ssr *StringRenderer) Render(current *VNode) {
	ssr.writeOpenTag(current)
	if !voidElements[current.tag] {
		ssr.writeBody(current)
	}
	ssr.writeCloseTag(current)
}

func (ssr *StringRenderer) writeBody(current *VNode) {
	if current.isDirty(flagChildren) {
		for _, child := range current.children {
			ssr.Render(child)
		}
	}
	if current.text.status == changeModified {
		ssr.buff.WriteString(html.EscapeString(current.text.nextValue))
	}
}

func (ssr *StringRenderer) writeCloseTag(current *VNode) {
	if voidElements[current.tag] {
		return
	}
	ssr.buff.WriteString("</")
	ssr.buff.WriteString(current.tag)
	ssr.buff.WriteString(">")
}

func (ssr *StringRenderer) writeOpenTag(current *VNode) {
	ssr.buff.WriteRune('<')
	ssr.buff.WriteString(current.tag)
	ssr.buff.WriteRune(' ')
	ssr.writeAttributes(current)

	if voidElements[current.tag] {
		ssr.buff.WriteString("/>")
	} else {
		ssr.buff.WriteRune('>')
	}
}

func (ssr *StringRenderer) writeAttributes(current *VNode) {
	if current.isDirty(flagClasses) {
		ssr.buff.WriteString(` class="`)
		for cls := range current.classes {
			ssr.buff.WriteString(html.EscapeString(cls))
			ssr.buff.WriteRune(' ')
		}
		ssr.buff.WriteString(`"`)
	}
	if current.isDirty(flagStyles) {
		ssr.buff.WriteString(` style="`)
		for styleName, styleValue := range current.styles {
			ssr.buff.WriteString(html.EscapeString(styleName))
			ssr.buff.WriteRune(':')
			ssr.buff.WriteString(html.EscapeString(styleValue.value))
			ssr.buff.WriteRune(';')
		}
		ssr.buff.WriteString(`"`)
	}
	if current.isDirty(flagAttributes) {
		for attrName, attrValue := range current.attributes {
			ssr.buff.WriteString(" ")
			ssr.buff.WriteString(attrName)
			ssr.buff.WriteString(`="`)
			ssr.buff.WriteString(html.EscapeString(attrValue.value))
			ssr.buff.WriteString(`"`)
		}
	}
}

func (ssr *StringRenderer) String() string {
	return ssr.buff.String()
}

func (ssr *StringRenderer) Reset() {
	ssr.buff.Reset()
}
