# Sticky Display

Make all windows on selected display(s) always sticky. Should work with Openbox, Fluxbox, IceWM, Xfwm, KWin, Marco, Muffin, Mutter and other [EWMH](https://en.wikipedia.org/wiki/Extended_Window_Manager_Hints#List_of_window_managers_that_support_Extended_Window_Manager_Hints) compliant window managers using the [X11](https://en.wikipedia.org/wiki/X_Window_System) window system.
Therefore, this project provides dynamic tiling for XFCE, LXDE, LXQt, KDE and GNOME (Mate, Deepin, Cinnamon, Budgie) based desktop environments.

Simply keep your current window manager and install **sticky-display on top** of it.
Once enabled, the window manager will sticky all windows that are created or enter the configured display(s).

## Features

- [x] Socket communication commands.
- [x] Adjustment of layout proportions.

## Installation

### Arch

Run `makepkg -si`.

## Usage

Run `sticky-display`.

## Configuration

To find the index of a display, open a terminal emulator on the display to check and run 'sticky-display -print-display'

The configuration file is located at `~/.config/sticky-display/config.toml` (or `XDG_CONFIG_HOME`) and is created with default values during the first startup.

## Development

Requirements: [go >= 1.18](https://go.dev/dl/)

### Install sticky-display via remote source

Install directly from main branch:
```bash
go install github.com/seyys/sticky-display@main
```

Start cortile in verbose mode:
```bash
$GOPATH/bin/cortile -v
```

## Issues

Debugging:
- If you encounter problems start the process with `sticky-display -vv`, which provides additional debug outputs.
- A log file is created by default under `/tmp/sticky-display.log`.

## Credits

Based on [cortile](https://github.com/leukipp/cortile) ([leukipp](https://github.com/leukipp/cortile)), [zentile](https://github.com/blrsn/zentile) ([Berin Larson](https://github.com/blrsn)), and [pytyle3](https://github.com/BurntSushi/pytyle3) ([Andrew Gallant](https://github.com/BurntSushi)).  
The main libraries used in this project are [xgbutil](https://github.com/BurntSushi/xgbutil), [toml](https://github.com/BurntSushi/toml), [fsnotify](https://github.com/fsnotify/fsnotify), and [logrus](https://github.com/sirupsen/logrus).

## License

MIT
