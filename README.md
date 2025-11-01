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

Example with code using TinyGo (just 1.5 MB of WASM): [Hix example](https://deltegui.github.io/hx/).
