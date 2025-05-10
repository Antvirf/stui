# Technology Context: stui

## Core Technologies

- **Programming Language**: Go (1.20+)
- **TUI Framework**: tview (GitHub.com/rivo/tview)
- **Build System**: Make + Goreleaser
- **Testing**: Standard Go testing framework

## Dependencies

### Required System Tools

- Slurm command-line tools:
  - `scontrol`
  - `sacct` (optional, for accounting features)
  - `sacctmgr` (optional, for accounting management)
  - `sdiag`

### Go Dependencies

- tview (TUI framework)
- Other Go packages as seen in go.mod

## Development Tooling

- **Build**: Makefile with standard targets
- **Release**: Goreleaser for cross-platform builds
- **Linting**: Pre-commit hooks
- **Testing**:
  - Unit tests in _test.go files
  - Test data in internal/model/testdata/

## Project Structure

```
stui/
├── internal/
│   ├── model/      # Data providers and parsers
│   ├── view/       # TUI components
│   └── config/     # Configuration handlers
├── testing/        # Integration test scripts
├── docs/           # Documentation
└── assets/         # Static assets
```

## Build & Run

```sh
make build  # Build binary
make run    # Run development version
```

## Configuration

- All configuration of `stui` happens via command line arguments given to the executable
