package common

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"io/ioutil"
	"log"
	"math"
	"os"
	"os/exec"
	"sync"
)

func TokenGenerator(size int) string {
	b := make([]byte, size)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

type ScreenshotHanlder interface {
	TakeScreenshot(url string, destination string) bool
}

type ChromeDPScreenshotHandler struct {
	once sync.Once
	ctx  context.Context
}

type GowitnessScreenshotHandler struct{}

//It has problem with concurrency. When run in parallel, it does not behave normally (Sometimes it doesn't get screenshot)

func (h *GowitnessScreenshotHandler) TakeScreenshot(url string,
	destination string,
	savetofile bool) ([]byte, bool) {
	log.Println("Taking Screenshot Using GoWitness")

	//args := fmt.Sprintf("--headless --disable-gpu --screenshot %s --no-sandbox",url)
	//args := fmt.Sprintf(" single --url=%s --destination=%s",url,dest)
	cmd := exec.Command(os.Getenv(
		"GOWITNESS_PATH"),
		"single", "--url="+url, "--destination="+destination,
		"--chrome-path=/usr/bin/chromium-browser",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Printf("Running command and waiting for it to finish...")
	err := cmd.Run()
	if err != nil {
		return nil, false
	}
	log.Printf("Command finished with error: %v", err)
	return nil, true

	//chromium-browser --headless --disable-gpu --screenshot https://www.chromestatus.com/ --no-sandbox

}

func (h *ChromeDPScreenshotHandler) TakeScreenshot(url string,
	destination string,
	savetofile bool) ([]byte, bool) {
	log.Println("Taking Screenshot Using Chrome")

	//Do the initialization only once (Just the first time)
	(*h).once.Do(func() { // <-- atomic, does not allow repeating
		(*h).init()
	})

	var buf []byte

	// capture entire browser viewport, returning png with quality=90
	if err := chromedp.Run(h.ctx, h.fullScreenshot(url, 90, &buf)); err != nil {
		log.Fatal(err)
	}
	if savetofile {
		if err := ioutil.WriteFile(destination+"/screenshot.png", buf, 0644); err != nil {
			log.Fatal(err)
		}
		return nil, true
	}
	log.Println("Finished Taking screenshot!!!!!!!!!!!!")

	return buf, true
}

func (h *ChromeDPScreenshotHandler) init() {
	fmt.Println("Init...")
	h.ctx, _ = chromedp.NewContext(context.Background())

}

func (h ChromeDPScreenshotHandler) fullScreenshot(urlstr string, quality int64, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// get layout metrics
			_, _, contentSize, err := page.GetLayoutMetrics().Do(ctx)
			if err != nil {
				return err
			}

			width, height := int64(math.Ceil(contentSize.Width)), int64(math.Ceil(contentSize.Height))

			// force viewport emulation
			err = emulation.SetDeviceMetricsOverride(width, height, 1, false).
				WithScreenOrientation(&emulation.ScreenOrientation{
					Type:  emulation.OrientationTypePortraitPrimary,
					Angle: 0,
				}).
				Do(ctx)
			if err != nil {
				return err
			}

			// capture screenshot
			*res, err = page.CaptureScreenshot().
				WithQuality(quality).
				WithClip(&page.Viewport{
					X:      contentSize.X,
					Y:      contentSize.Y,
					Width:  contentSize.Width,
					Height: contentSize.Height,
					Scale:  1,
				}).Do(ctx)
			if err != nil {
				return err
			}
			return nil
		}),
	}
}
