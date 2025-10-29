![logo](https://raw.githubusercontent.com/deltegui/hix/refs/heads/main/logo.png)

# Hix

Hix is a reactive, component-based Golang WASM frontend framework. Example:

Here is the [demo](https://deltegui.github.io/hx/)

```go
func counter() hx.INode {
	c := hx.Signal(0)
	hx.EffectFunc(func() {
		fmt.Println("Change!", c.Get())
	})
	str := hx.Computed(func() string {
		return strconv.FormatInt(int64(c.Get()), 10)
	})

	div := hx.Div().
		BindText(str).
		Style("float", "left").
		Style("padding-left", "20px").
		Style("padding-right", "20px")
	return hx.Div().Body(
		hx.Button().
			Text("-").
			Style("float", "left").
			OnClick(func(ctx hx.EventContext) {
				c.Set(c.Get() - 1)
			}),
		div,
		hx.Button().
			Text("+").
			Style("float", "left").
			OnClick(func(ctx hx.EventContext) {
				c.Set(c.Get() + 1)
			}),
	)
}

func showMessage(msg string) hx.INode {
	show := hx.Signal(false)
	div := hx.Div().Id("43")

	hx.EffectFunc(func() {
		div.Body(
			hx.If(show, hx.P().Text(msg)),
			hx.Button().Text("Toggle").OnClick(func(ctx hx.EventContext) {
				show.Set(!show.Get())
			}),
			hx.A().Href("http://google.es").Text("Go to google"),
		)
	})

	return div
}

func demoEach() hx.INode {
	elements := hx.Signal([]string{})
	textInput := hx.Signal("")

	main := hx.Div()

	hx.EffectFunc(func() {
		input := hx.Input().BindOnInput(textInput)
		addBtn := hx.Button().Text("Add").OnClick(func(ctx hx.EventContext) {
			ee := elements.Get()
			ee = append(ee, textInput.Get())
			textInput.Set("")
			elements.Set(ee)
		})

		main.Body(
			input,
			addBtn,
			hx.Ul().Body(
				hx.Each(elements, func(index int, value string) hx.INode {
					return hx.Li().Body(
						hx.P().Text(fmt.Sprintf("[%d] %s", index, value)),
						hx.Button().Text("Delete").OnClick(func(ctx hx.EventContext) {
							ee := elements.Get()
							ee = append(ee[:index], ee[index+1:]...)
							elements.Set(ee)
						}),
					)
				}),
			),
		)
	})
	return hx.Div().Body(
		hx.H1().Text("List of items"),
		main,
	)
}

func reactTextInput() hx.INode {
	text := hx.Signal("")
	return hx.Div().Body(
		hx.P().BindText(text),
		hx.Input().BindOnInput(text),
	)
}

func main() {
	point := hx.NewFromId("wasm_mount_point")
	point.Body(
		hx.Div().Body(
			hx.P().
				Style("background-color", "blue").
				Style("color", "white").
				Text("Hola").
				On(hx.EventClick, func(ctx hx.EventContext) {
					ctx.Target.
						Style("background-color", "black").
						Style("color", "red")
				}),
			hx.Button().
				Text("Show something in console").
				On(hx.EventClick, func(ctx hx.EventContext) {
					fmt.Println("Hi!")
				}),
			counter(),
			hx.Br(),
			showMessage("Toggle text"),
			reactTextInput(),
			demoEach(),
		),
	)
	select {}
}

```
