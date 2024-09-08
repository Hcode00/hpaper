
# hpaper: Automated Wallpaper Management for Hyprland Using Hyprpaper

hpaper is a powerful and flexible Go application that automates wallpaper management for Hyprland using hyprpaper. It offers a seamless wallpaper rotation experience with intelligent preloading and unloading mechanisms.

## Key Features

- **Smart Rotation**: Automatically cycles through your wallpaper collection at customizable intervals.
- **Efficient Resource Management**: Preloads upcoming wallpapers and unloads unnecessary ones to optimize memory usage.
- **Manual Control**: Allows users to manually switch to the next or previous wallpaper, with the rotation timer automatically resetting.
- **Customizable Buffer**: Adjust the number of preloaded wallpapers to balance between responsiveness and resource usage.
- **Graceful Error Handling**: Continues operation even if a wallpaper fails to load, ensuring uninterrupted service.

## How It Works

hpaper maintains a buffer of preloaded wallpapers, intelligently managing which images to keep in memory. It supports both automatic timed rotation and manual control, making it suitable for various user preferences.

## Installation

```fish
  go build github.com/Hcode00/hpaper
```

## hpaper Usage

**Basic Usage:**
```bash
hpaper start [directory] [duration in seconds] [maximum number of pictures to preload]
hpaper start [image file]
```
* **`start`:** Begins the wallpaper rotation.
* **`directory`:** Specifies the directory containing the wallpaper images.
* **`duration`:** Sets the time interval between wallpaper changes (in seconds).
* **`maximum number of pictures to preload`:** Limits the number of wallpapers preloaded into memory.
* **`image file`:** Directly specifies a single image file as the wallpaper.

**Commands:**
```bash
hpaper [next|prev|status|quit]
```
* **`next`:** Sets the next wallpaper in the list.
* **`prev`:** Sets the previous wallpaper in the list.
* **`status`:** Displays the current wallpaper name and preloaded wallpapers.
* **`quit`:** Stops the wallpaper rotation.

