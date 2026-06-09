# gopacgo

A minimal, colorful CLI tool for checking outdated AUR packages on Arch Linux.

## What it does

`gopacgo` queries `pacman` for your locally installed foreign (AUR) packages, then checks the [AUR RPC API](https://aur.archlinux.org/rpc/) for their latest versions — all in a single batch request. Results are displayed in a live terminal UI that reveals each package's status as it resolves.

```
  🔍 gopacgo  AUR package checker
  ──────────────────────────────────────────────

  yay               12.3.1    ✓  up to date
  paru              2.0.1     ↑  2.1.0
  zsh-completions   0.35.0    ⣾  checking
  some-aur-pkg      1.0.0     ✗  not found

  ──────────────────────────────────────────────
  ✓  1 up to date  ·  ↑  1 to update  ·  ✗  1 not found
```

## Status indicators

| Symbol | Meaning |
|--------|---------|
| `✓` | Up to date |
| `↑` | Update available — shows the new version |
| `✗` | Package not found in AUR |
| `⣾` | Checking… |

## Requirements

- Arch Linux (or any system with `pacman`)
- Go 1.21+

## Build

```bash
git clone https://github.com/andrequeiroz/gopacgo
cd gopacgo
go build -o gopacgo
```

## Usage

```bash
./gopacgo
```

Press `q` or `Ctrl+C` to abort at any time. When all packages are resolved, the program exits automatically and the results remain visible in the terminal.

## About

The idea for `gopacgo` came from a real workflow problem: manually browsing the AUR website to check for package updates on a minimal Arch Linux installation that doesn't use an AUR helper. The core logic — running `pacman -Qm` to list foreign packages and querying the AUR RPC API to fetch latest versions — was designed and implemented by me.

The terminal UI was built with assistance from [Claude](https://claude.ai) (Anthropic), using [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lip Gloss](https://github.com/charmbracelet/lipgloss) from the [Charm](https://charm.sh) ecosystem. The per-package reveal animation uses a random delay of up to five seconds — not because the lookup takes that long (the AUR API responds instantly in a single batch call), but because watching results appear one by one feels more satisfying than a sudden complete render.

## License

MIT — see [LICENSE](LICENSE).
