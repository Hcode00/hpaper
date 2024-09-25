package wallpapers

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"

	u "github.com/Hcode00/hpaper/utils"
)

func DownloadFile(dir, width, height string, max uint, webP bool) error {
	lenList := make([]int, 0, max)
	logger := u.LOG
	ext := ""
	if webP {
		ext = ".webp"
	} else {
		ext = ".jpg"
	}
	url := "https://picsum.photos/" + width + "/" + height + ext
	for i := uint(0); i < max; i++ {
		id := rand.Uint32()
		idStr := strconv.Itoa(int(id))
		path := dir + idStr + ext

		l := fmt.Sprintf("Downloading image %d saved in %s", i+1, path)
		logger.Log(l)
		out, err := os.Create(path)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to create file %s: %v", path, err))
			return err
		}
		defer out.Close()

		resp, err := http.Get(url)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to download from %s: %v", url, err))
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			logger.Warn(fmt.Sprintf("Unexpected status code: %s", resp.Status))
			continue
		}
		b, _ := io.ReadAll(resp.Body)
		for _, v := range lenList {
			if len(b) == v {
				logger.Debug("Repeated Image Skipping ...")
				i -= 1
				continue
			}
		}
		lenList = append(lenList, len(b))
		err = os.WriteFile(path, b, 0o644)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to write to file %s: %v", path, err))
			return err
		}

		logger.Debug(fmt.Sprintf("Successfully downloaded and saved %s", idStr+ext))
	}

	logger.Log(fmt.Sprintf("Completed downloading %d wallpapers", max))
	return nil
}
