# Ditto (WIP)

**Declarative package sync for Arch-based systems.**

Define all your `pacman` packages in simple `.def` files and sync your system with one command — no more guessing or trying to remember what you installed last week.

## Why not NixOS?

Because you just want *one* thing: to declaratively manage packages without rewriting your OS.

---

## Installation

Install via `go install`:

```sh
go install github.com/ony-boom/ditto@latest
```

---

## Usage

Define your desired packages in plain text files inside:

```
~/.config/ditto/packages/
```

Each `.def` file is just a list of package names, one per line. Lines starting with `#` are comments.

### Per-host definitions

For machine-specific packages, create files under:

```
~/.config/ditto/packages/hosts/<hostname>.def
```

You can also organize them into subdirectories (e.g. `hosts/laptop/gaming.def`).

---

## Syncing

Once your definitions are ready, run:

```sh
ditto sync
```

By default, Ditto will install any missing packages.  
If you want to **remove** packages that are not listed (excluding ignored ones), use strict mode:

```sh
ditto sync --strict
```

You can preview changes with:

```sh
ditto sync --dry-run
```

---

## Notes

- The config file is automatically created at `~/.config/ditto/config.toml` on first run.
- You can specify an AUR helper and other options in the config.

---

## Example `.def` file

```text
# Core tools
git
htop
curl

# Optional
neovim
firefox
```
