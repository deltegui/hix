package main

import (
	"fmt"

	"github.com/deltegui/hx"
)

func BSRow() hx.INode {
	return hx.Div().Class("row")
}

func BSCol(size int) hx.INode {
	colClass := fmt.Sprintf("col-%d", size)
	return hx.Div().Class(colClass)
}

func DemoCounterPage() hx.INode {
	c := hx.Signal(0)
	cstr := hx.Computed(func() string {
		return fmt.Sprintf("Current count %d", c.Get())
	})
	div := hx.Div()
	hx.EffectFunc(func() {
		div.Body(
			BSRow().Body(
				BSCol(12).Body(
					hx.P().Text("Demo of computed elements"),
				),
				BSCol(12).Body(
					hx.Button().
						Text("-").
						Class("btn", "btn-primary").
						OnClick(func(ctx hx.EventContext) {
							c.Set(c.Get() - 1)
						}),
					hx.Input().
						BindValue(cstr).
						Class("form-control").
						Attribute("readonly", "readonly"),
					hx.Button().
						Text("+").
						Class("btn", "btn-primary").
						OnClick(func(ctx hx.EventContext) {
							c.Set(c.Get() + 1)
						}),
				),
			),
		)
	})
	return div
}

func DemoInputPage() hx.INode {
	txt := hx.Signal("")
	return hx.Div().Body(
		BSRow().Body(
			BSCol(12).Body(
				hx.H2().Text("Demo echo input"),
				hx.P().Text("Write in this input. All you write will be echoed in the text area below. Demo of Hix reactivity"),
			),
			BSCol(12).Body(
				hx.Input().
					BindOnInput(txt).
					Class("form-control"),
			),
			BSCol(12).Class("mt-3").Body(
				hx.TextArea().
					BindText(txt).
					Class("form-control"),
			),
		),
	)
}

func main() {
	point := hx.NewFromId("wasm_mount_point")
	tabs := bsTabsComponent{
		Tabs: []BSTab{
			{
				Text:   "Demo input",
				ID:     "demo_input",
				Active: true,
				Body:   DemoInputPage(),
			},
			{
				Text: "Counter",
				ID:   "counter",
				Body: DemoCounterPage(),
			},
		},
	}
	point.Body(
		BSRow().Class("mt-4").Body(
			BSCol(3).Body(
				hx.Img().
					Src("https://raw.githubusercontent.com/deltegui/hix/refs/heads/main/logo.png").
					Class("img-thumbnail"),
			),
			BSCol(9).Body(
				hx.H1().Text("Hix framework"),
				hx.H3().Text("A simple golang wasm web framework!"),
			),
		),
		BSRow().Class("mt-4").Body(
			BSCol(12).Body(tabs.render()),
		),
	)
	select {}
}

type BSTab struct {
	Text     string
	ID       string
	Active   bool
	Disabled bool
	Body     hx.INode
}

type bsTabsComponent struct {
	Tabs []BSTab
}

func (bs *bsTabsComponent) render() hx.INode {
	signal := hx.Signal(bs.Tabs)

	tabBodyList := hx.Div().Class("tab-content")
	tabList := hx.Ul().Class("nav", "nav-tabs")
	hx.EffectFunc(func() {
		tabBodyList.Body(
			hx.Each(signal, func(i int, value BSTab) hx.INode {
				//fmt.Println("Render body!")
				//fmt.Println(bs.Tabs)
				div := hx.Div().
					Class("tab-pane", "fade").
					Id(value.ID).
					Body(value.Body)
				if value.Active {
					div.Class("active", "show")
				} else {
					div.RemoveClass("active", "show")
				}
				return div
			}),
		)
		tabList.Body(
			hx.Each(signal, func(i int, value BSTab) hx.INode {
				a := hx.A().Class("nav-link")
				if value.Active {
					a.Class("active")
				}
				if value.Disabled {
					a.Class("disabled")
				}
				a.Text(value.Text)
				a.OnClick(func(ctx hx.EventContext) {
					//fmt.Println("Click!", i)
					for i, tab := range bs.Tabs {
						if tab.ID == value.ID {
							bs.Tabs[i].Active = true
						} else {
							bs.Tabs[i].Active = false
						}
					}
					//fmt.Println(bs.Tabs)
					signal.Set(bs.Tabs)
				})
				return hx.Li().Class("nav-item").Body(a)
			}),
		)
	})

	return hx.Div().Body(tabList, tabBodyList)
}
