# hpaper: Automated Wallpaper Management for Hyprland Using Hyprpaper

hpaper is a flexible Go application that automates wallpaper management for Hyprland using hyprpaper. It offers a seamless wallpaper rotation experience with  preloading and unloading mechanisms.

## Key Features

- **Smart Rotation**: Automatically cycles through your wallpaper collection at customizable intervals.
- **Efficient Resource Management**: Preloads upcoming wallpapers and unloads unnecessary ones to optimize memory usage.
- **Manual Control**: Allows users to manually switch to the next or previous wallpaper, with the rotation timer automatically resetting.
- **Customizable Buffer**: Adjust the number of preloaded wallpapers to balance between responsiveness and resource usage.
- **Wallpaper Downloading**: Download beautiful wallpapers directly from the [Picsum API](https://picsum.photos)  for easy customization.

## How It Works

hpaper maintains a buffer of preloaded wallpapers, managing which images to keep in memory. It supports both automatic timed rotation and manual control, making it suitable for various user preferences.

## Pre Installation
**Enable IPC for Hyprpaper**

Add the following line to your `~/.config/hyprpaper/hyprpaper.conf` file:
```
ipc = true
```
**For more information, refer to the Hyprpaper wiki:**

- [Hyprpaper Wiki](https://wiki.hyprland.org/Hypr-Ecosystem/hyprpaper/)


## Installation
if you are on arch linux the package is available in the AUR

```sh
paru -S hpaper
```

```sh
go install github.com/Hcode00/hpaper
```

**Alternatively, download the binary from the release page.**


## hpaper Usage


**Basic Usage:**

```sh
# Starts wallpaper rotation from a specified directory, sets rotation interval, and preloads a maximum number of wallpapers.
hpaper start [directory] [duration in seconds] [maximum number of pictures to preload] [flags]
# Sets a single image as the wallpaper.
hpaper start [image file]
# Downloads a specified number of wallpapers and saves them to the given directory.
hpaper download [directory] [number of pictures] [width] [height] [flags]
```
* **`start`** Begins the wallpaper rotation or sets a single image as wallpaper.
* **`directory`** (for `start` with directory) Specifies the directory containing the wallpaper images.
* **`duration`** (for `start` with directory) Sets the time interval between wallpaper changes (in seconds).
* **`maximum number of pictures to preload`** (for `start` with directory) Limits the number of wallpapers preloaded into memory.
* **`image file`** (for `start` with image file) Specifies a single image file as the wallpaper.
* **`directory`** (for `download`) Path to save the downloaded wallpapers.
* **`number of pictures`** (for `download`) Defines the number of wallpapers to download (1-20).
* **`width`** (for `download`) Sets the desired width of downloaded wallpapers .
* **`height`** (for `download`) Sets the desired height of downloaded wallpapers.
* **`flags`** Added flags
    * **`-r`** (for `start` with directory) Randomize wallpapers list at start (optional)
    * **`-w`** (for `download`) Download wallpapers in WebP format (optional)
  
**Commands:**
```sh
hpaper [next | prev | status | help | quit]
```
* **`next`** Sets the next wallpaper in the list.
* **`prev`** Sets the previous wallpaper in the list.
* **`status`** Displays the current wallpaper name and preloaded wallpapers.
* **`help`** Show useful help information.
* **`quit`** Stops the wallpaper rotation.

## Examples

**Example for downloading an image:**

```bash
hpaper download ~/Pictures 1 1920 1080
```

This command will download a single wallpaper with a resolution of 1920x1080 from the Picsum API and save it to the `~/Pictures` directory.

**For Hyprland Config:**

in your Hyprland config at **`~/.config/hypr/hyprland.config`:**

```hyprlang
# start hpaper on this directory and keep 3 images preloaded at all times and switch images every one hour
# use -r flag to randomize wallpapers list at the start
exec-once = hpaper start ~/.config/hypr/wallpapers/ 3600 3 -r

# as simple as that switch to next and previous wallpaper
bind = SUPER, W, exec, hpaper next
bind = SUPER SHIFT, W, exec, hpaper prev
```
