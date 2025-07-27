# Ditto

A tiny tool to keep your `pacman` packages in sync using plain text files.

## Install

```sh
go install github.com/ony-boom/ditto@latest
```

## How it works

1. Throw some package names into `~/.config/ditto/packages/`.
2. One package per line. `#` starts a comment because of course it does.
3. Run:

   ```sh
   ditto sync
   ```

   Ditto will install whatever’s missing and make you look like you have your life together.

### Host-specific packages

Got multiple machines? Put their stuff in:

```
~/.config/ditto/packages/hosts/<hostname>.pkgs
```

Organize into subfolders if you’re *that* person (`hosts/laptop/gaming.pkgs`).

## Options

* `--strict` → yeets packages not in your list! (be careful with this one)
* `--dry-run` → shows what would happen without touching anything (like commitment-free package management).

## Example `.pkgs`

```text
# Tools I actually use
git
htop
curl

# Tools I *promise* I’ll learn someday
neovim
firefox
```
