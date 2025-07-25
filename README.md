# hpaper: Advanced Wallpaper Management for Wayland

hpaper is a blazingly fast Go-based wallpaper management daemon for Wayland compositors that provides automated and manual rotation, multi-backend support, and seamless integration with your desktop environment.

## Video Showcase

[![hpaper Demo](https://github.com/Hcode00/hpaper/blob/main/showcase.gif)]()

## Key Features

- **Multi-Backend Support**: Works with both [Swaybg](https://github.com/swaywm/swaybg) and [Hyprpaper](https://github.com/hyprwm/hyprpaper) wallpaper engines.
- **Smart Daemon Architecture**: Runs as a background daemon with instant client commands for seamless wallpaper switching.
- **Configuration-Driven**: Comprehensive configuration file support with runtime overrides and automatic defaults.
- **Pywal Integration**: Optional automatic color scheme generation using pywal for system-wide theming.
- **Flexible Rotation**: Customizable auto-rotation intervals with manual control and timer reset functionality.
- **Monitor-Specific Control**: Supports per-monitor wallpaper management for multi-display setups.
- **Low Memory Usage**: uses just about 1 to 8 megabytes of memory.

## How It Works

hpaper operates as a daemon process that manages wallpaper rotation and responds to client commands through Unix socket IPC. The daemon loads wallpapers from your specified directory, handles backend-specific wallpaper setting, and provides instant response to manual controls without interrupting the rotation cycle.

## Installation

**Arch Linux (AUR):**
```sh
paru -S hpaper
```

**Go Install:**
```sh
go install github.com/Hcode00/hpaper
```

**Binary Download:**
Download the latest binary from the [releases page](https://github.com/Hcode00/hpaper/releases).

## Usage

**Starting the Daemon:**

```sh
# Start with wallpaper directory (creates default config if none exists)
hpaper start <wallpaper_directory>

# Start with custom configuration file
hpaper start <wallpaper_directory> -config /path/to/config.conf
```

**Runtime Commands:**

```sh
hpaper next      # Switch to next wallpaper
hpaper prev      # Switch to previous wallpaper  
hpaper current   # Get path of current wallpaper
hpaper quit      # Stop the daemon
```

## Configuration

hpaper automatically creates a configuration file at `~/.config/hpaper/hpaper.conf` with the following options:

**General Settings:**
- **`wallpaper_dir`** - Directory containing wallpaper images
- **`rotation_interval`** - Auto-rotation interval in seconds (0 disables)
- **`randomize`** - Shuffle wallpaper order on startup
- **`backend`** - Backend engine: `swaybg` or `hyprpaper`

**Swaybg Settings:**
- **`swaybg_mode`** - Display mode: `fill`, `fit`, `stretch`, `center`, `tile`
- **`swaybg_output`** - Target output (monitor) or `*` for all

**Hyprpaper Settings:**
- **`monitor_name`** - Monitor identifier (e.g., `DP-1`) or `all`

**Pywal Integration:**
- **`pywal_enabled`** - Enable automatic color scheme generation
- **`pywal_command`** - Custom pywal command with arguments

## Examples

**Basic Setup:**
```bash
# Start daemon with 30-minute rotation
hpaper start ~/Pictures/Wallpapers/
```

**Hyprland Integration:**
```hyprlang
# Auto-start hpaper daemon
exec-once = hpaper start ~/.config/hypr/wallpapers/

# Wallpaper controls
bind = SUPER, W, exec, hpaper next
bind = SUPER SHIFT, W, exec, hpaper prev
```

**Configuration Example:**
```conf
# ~/.config/hpaper/hpaper.conf
wallpaper_dir = /home/user/wallpapers
rotation_interval = 1800
randomize = true
backend = hyprpaper
monitor_name = DP-1
pywal_enabled = true
pywal_command = wal --cols16 -n -q
```
