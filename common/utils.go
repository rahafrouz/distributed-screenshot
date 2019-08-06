package common

import (
	"crypto/rand"
	"fmt"
)

func TokenGenerator(size int) string {
	b := make([]byte, size)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

type ScreenshotHanlder interface {
	TakeScreenshot(url string, destination string, savetofile bool) ([]byte, bool)
}
