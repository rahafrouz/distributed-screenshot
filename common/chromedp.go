package common

import (
	"context"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"io/ioutil"
	"log"
	"math"
	"sync"
)

//The ChromeDP handler of screenshot. This engine seems more mature than gowitness. Chrome has to be installed on the machine to use properly.
type ChromeDPScreenshotHandler struct {
	once sync.Once
	ctx  context.Context
}

//Take the screenshot using Chromedp, and save the result in the byte[] format.
func (h *ChromeDPScreenshotHandler) TakeScreenshot(url string, destination string, savetofile bool,
) ([]byte, bool) {

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

//Initialize should be called only once
func (h *ChromeDPScreenshotHandler) init() {
	log.Println("Init...")
	h.ctx, _ = chromedp.NewContext(context.Background())

}

//Uses the library of chromedp to take a fullscreenshot of the page.
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
