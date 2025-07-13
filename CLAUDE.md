# CLAUDE.md

This is a Wails v3 desktop application built with Go and Svelte. Wails automates the process of building and packaging the Svelte frontend and Go backend into a single executable.

## Commands

All project tasks are made available through the taskfile. You can run, but are not limited to, the following commands:

- `task build`: Build the application for the current platform.
- `task dev`: Run the application in development mode.
- `task test`: Run the tests.
- `task lint`: Run the linter.
- `task fmt`: Format the code.

## Architecture

Backkend loads configuration, manages integrations, and sends data to the frontend.

- `main.go` loads configuration, starts services, and runs the Wails application.
- `pkg/` contains Go packages for services, utilities, and types.
  - `pkg/config/` handles configuration loading and validation.
  - `pkg/dashboard/service.go` defines the main service for the dashboard.
  - `pkg/integrations/` defines integration services for various platforms.
  - `pkg/widgets/` contain widget definitions and logic.
- `frontend/` contains the Svelte frontend code.
  - `frontend/bindings/` contains auto-generated TypeScript bindings for Go services.
  - `frontend/src/Dashboard.svelte` is the main dashboard component.
  - `frontend/src/widgets` defines frontend components for backend widgets.

## Wails

This project uses Wails v3alpha. Here are some key points:

- Wails provides a bridge between the Go backend and the Svelte frontend using *Services* and *Bindings*
- Wails auto-generates bindings from Go services. `wails3 generate bindings -ts`
- Services are Go structs with methods that can be called from the frontend.
- Services have optional lifecycle methods: ServiceStartup(), ServiceShutdown(), ServiceName(), and ServeHTTP() for HTTP handling.
- Structs are converted to TypeScript classes with propery typing.
- Bindings support `context.Context`. Context cancellation yields promise rejection.
- Errors automatically become promise rejections.

## Svelte

This project uses Svelte 5. Importantly, this version adds support for *runes*.

- Runes are compiler instructions starting with `$` that control reactivity explicitly. They are not imported.
- They work in `.svelte` and `.svelte.ts` files.
- `$state` declares reactive state. It is deeply reactive unless you use `$state.raw()`.
  - `let count = $state(0)`
- `$derived` declares auto-updating computed values.
  - `const doubled = $derived(count * 2)`
  - `const filtered = $derived.by(() => items.filter(...))` for more complex cases.
- `$effect` runs a function when dependencies change.
  - `$effect(() => { console.log(count); })`
  - `$effect(() => { setup(); return () => cleanup(); })` to support cleanup.
- `$props` is used to declare reactive props in components. Read only by default.
  - `let { name, age = 18 } = $props()`
- `$bindable` to make props two-way bindable.
  - `let { value = $bindable('') } = $props()`
  - Parent uses `<Component bind:value={myValue} />`
- `$inspect` is a dev-only state tracker.
  - `$inspect(count)` to log changes in `count` during development.
