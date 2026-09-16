# Configuration

The configuration reference lives on the docs site: https://tuios.gaurav.zip/docs/configuration

It covers the whole `config.toml`: the `[appearance]` table and its `sidebar`, `scrollbar`, dock, and window-button options, `[notifications.agent]`, all 19 `[keybindings]` sections, `[daemon]`, `[startup]`, `[tape]`, `[screenshot]`, `[wallpaper]`, `[hooks]`, and `[debug]`, along with what hot-reloads and what needs a restart.

`tuios list-options` describes every settable path with its type, default, and accepted values, straight from the registry the validator uses. The in-app settings page (`Ctrl+B ,`) edits and persists the same options, and its rows are derived from that same registry: an option an agent can set is an option a person can reach, and a test fails the build if one is not.

## The wallpaper

The `[wallpaper]` table puts a picture on the empty desktop, behind every
pane and under the dock, the sidebar and any popup. It is drawn by the client
on its own terminal, so it is client-local like the theme.

```toml
[wallpaper]
path = "~/Pictures/wall.png"  # png, jpeg or gif; empty for none
mode = "fill"                 # fill crops to cover the desktop, fit shows all of it
dim = 30                      # percent darkening, 0 to 100
renderer = "auto"             # auto, kitty or cells
```

On a terminal with kitty graphics (kitty, ghostty, wezterm) the picture is
drawn as pixels, placed only where no pane is. Anywhere else it is drawn as
half-block cells, two pixels per cell, which needs truecolor. `renderer`
forces one path, which is how the cell rendering can be checked on a
terminal that would otherwise never show it. The table hot-reloads.

## The dock's components

The `[dock]` table is the one part of the configuration that is not a set of
scalar options, so `list-options` does not carry it. The settings page edits it
through an editor of its own rather than a row: **Dock → Components**, where the
three regions and what is in them are one list. Shifted arrows move a component
and carry it into the next region off the end of its own, Enter takes one off
the bar or puts it back, `u` undoes the session's edits and `r` restores the
defaults. It is three ordered lists of component names, plus a table per custom
component:

```toml
[dock]
left   = ["mode", "workspaces", "trail", "tape"]
center = ["windows"]
right  = ["notifications", "copy-help", "cpu", "ram", "clock", "session-controls"]

[dock.clock]
format = "15:04"

[dock.custom.branch]
command  = "~/.config/tuios/dock/git-branch.sh"
refresh  = "event:after-focus-change"
on-click = "tuios new-window log git log --oneline -20"
```

The lists above are the default: omit the whole table and the bar is unchanged.
A custom component's first line of stdout becomes its cell, it is hidden when
the command fails, and `tuios list-dock-components` says which and why.

`examples/dock/README.md` is the full contract and five working recipes.
