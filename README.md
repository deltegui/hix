![logo](https://raw.githubusercontent.com/deltegui/hix/refs/heads/main/logo.png)

# Hix

Hix is a reactive frontend library for Go (WASM).
It aims to be **small, simple, and predictable**.

In around 1K lines of code, it can:

* Write HTML in plain old Go. No code generation needed. No strange mixed files without editor support.

* Handle JavaScript events directly.

* Provide reactivity using **Signals**, **Computed**, and **Effects** (the *Solid.js* model).

* Update the DOM efficiently — only when changes occur, minimizing Go ↔ JS calls.

* Work seamlessly with **TinyGo**, reducing bundle sizes dramatically.

Hix only depends on *honnef.co/go/dom/v2*.

Hix doesn’t try to be a full-fledged framework — it’s just a library for writing reactive web interfaces in Go.

If you need a full framework to build PWAs or complex applications, try go-app.

Example with code using TinyGo (just 1.5 MB of WASM, 191K gzip): [Hix example](https://deltegui.github.io/hx/).

Thanks to TinyGo and Hix simplicity, the example shown above can achieve these sizes:

| Compiled                  | Command                                                                   | Size |
|---------------------------|---------------------------------------------------------------------------|------|
| Go full                   | `GOOS=js GOARCH=wasm go build -o main.wasm main.go`                       | 5.1M |
| TinyGo                    | `GOOS=js GOARCH=wasm tinygo build -o main.wasm main.go `                  | 1.5M |
| Optimized TinyGo          | `GOOS=js GOARCH=wasm tinygo build --no-debug -opt=z -o main.wasm main.go` | 631K |
| Optimized TinyGo + GZip   | `gzip -k -9 main.wasm`                                                    | 191K |
| Optimized TinyGo + brotli | `brotli -f -q 11 main.wasm`                                               | 138K |

## Why Hix?

The Go WASM ecosystem feels somewhat abandoned. Vecty hasn’t received updates in four years. Vugu also hasn’t had any major updates for quite some time (and .vugu files have little to no support in editors). go-app seems to be the way to go, but it’s mainly oriented towards building PWAs.

I needed interactivity for a project that already uses backend rendering, but I didn’t want to rely on the Node.js ecosystem — they don’t seem to care much about having a ton of dependencies, dependency hell or breaking changes, which compromises long-term maintainability. I started Hix as a small demo, but I liked it so much that I began using it in my side project.
