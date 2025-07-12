# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Development
- `wails3 dev` - Run in development mode with hot-reload
- `task dev` - Alternative way to run development mode

### Build
- `wails3 build` - Build production executable
- `task build` - Build using task runner (cross-platform)
- `task package` - Package production build
- `task run` - Run the built application

### Frontend Development
- `cd frontend && npm run dev` - Run Vite dev server
- `cd frontend && npm run build` - Build frontend for production
- `cd frontend && npm run check` - Run Svelte type checking

## Architecture

This is a **Wails v3** desktop application combining:
- **Go backend** (`main.go`, `greetservice.go`) - Handles application logic and services
- **Svelte frontend** (`frontend/`) - TypeScript-based UI with Vite bundling
- **Cross-platform builds** managed through Task files in `build/` directory

### Key Components

**Backend Structure:**
- `main.go` - Application entry point, window configuration, and event emitter
- `greetservice.go` - Service layer exposed to frontend via Wails bindings
- Services are registered in `application.New()` and auto-generate TypeScript bindings

**Frontend Structure:**
- `frontend/src/App.svelte` - Main UI component
- `frontend/src/bindings/` - Auto-generated TypeScript bindings for Go services
- Uses `@wailsio/runtime` for Go service calls and event handling

**Build System:**
- `Taskfile.yml` - Main task runner with OS-specific builds
- `build/` directory contains platform-specific build configurations
- `frontend/dist/` gets embedded into Go binary via `embed.FS`

### Development Flow

1. Frontend builds are embedded into Go binary at compile time
2. Go services are exposed to frontend via Wails bindings
3. Real-time communication via events (e.g., time updates from Go to frontend)
4. Frontend calls Go services using generated TypeScript bindings

### Testing

- Frontend: `cd frontend && npm run check` for Svelte type checking
- No test runner currently configured - add tests as needed

### Logging and Debugging

- I will run dev tools or the application itself. If you want to see logs from backend or frontend, just ask.