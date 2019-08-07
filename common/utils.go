package common

import (
	"crypto/rand"
	"fmt"
)

//Generate a unique random token.
func TokenGenerator(size int) string {
	b := make([]byte, size)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

//This is the interface for the engine of taking screenshot.
// In the current implementation they are `chromedp` and `gowitness`.
//Possible engines for future could be eyewitness, pupeteer, etc.
type ScreenshotHanlder interface {
	TakeScreenshot(url string, destination string, savetofile bool) ([]byte, bool)
}

//The cloud storage, should be swappable.
// In our example, we use S3, but to add new stoarge engine, you should create a type that implements such interfaces.
type CloudStorageHandler interface {
	init()
	UploadDataToCloud(fileNameInCloud string, data []byte) (string, error)
}
