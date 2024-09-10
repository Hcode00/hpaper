# hpaper: Automated Wallpaper Management for Hyprland Using Hyprpaper

hpaper is a flexible Go application that automates wallpaper management for Hyprland using hyprpaper. It offers a seamless wallpaper rotation experience with  preloading and unloading mechanisms.

## Key Features

- **Smart Rotation**: Automatically cycles through your wallpaper collection at customizable intervals.
- **Efficient Resource Management**: Preloads upcoming wallpapers and unloads unnecessary ones to optimize memory usage.
- **Manual Control**: Allows users to manually switch to the next or previous wallpaper, with the rotation timer automatically resetting.
- **Customizable Buffer**: Adjust the number of preloaded wallpapers to balance between responsiveness and resource usage.

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

```fish
paru -S hpaper
```

```fish
go build github.com/Hcode00/hpaper
```


or you can download the binary from the release page


## hpaper Usage


**Basic Usage:**

```bash
hpaper start [directory] [duration in seconds] [maximum number of pictures to preload]
hpaper start [image file]
```
* **`start`** Begins the wallpaper rotation.
* **`directory`** Specifies the directory containing the wallpaper images.
* **`duration`** Sets the time interval between wallpaper changes (in seconds).
* **`maximum number of pictures to preload`** Limits the number of wallpapers preloaded into memory.
* **`image file`** Directly specifies a single image file as the wallpaper.

**Commands:**
```bash
hpaper [next | prev | status | quit]
```
* **`next`** Sets the next wallpaper in the list.
* **`prev`** Sets the previous wallpaper in the list.
* **`status`** Displays the current wallpaper name and preloaded wallpapers.
* **`quit`** Stops the wallpaper rotation.

## Example
in your Hyprland config at **`~/.config/hypr/hyprland.config`:**

```hyprlang
# start hpaper on this directory and keep 3 images preloaded at all times and switch images every one hour
exec-once = hpaper start ~/.config/hypr/wallpapers/ 3600 3

# as simple as that switch to next and previous wallpaper
bind = SUPER, W, exec, hpaper next
bind = SUPER SHIFT, W, exec, hpaper prev
```
