package datamodel

import (
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
)

type ScreenshotHanlder interface {
	TakeScreenshot(url string, destination string) bool
}

func TokenGenerator(size int) string {
	b := make([]byte, size)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

type ChromeDPScreenshotHandler struct {
	once sync.Once
}

func (h *ChromeDPScreenshotHandler) TakeScreenshot(url string, destination string) bool {
	//Do the initialization only once
	h.once.Do(func() { // <-- atomic, does not allow repeating
		h.init()
	})
	fmt.Println("Taking Screenshot Using Chrome")

	return true
}

func (h *ChromeDPScreenshotHandler) init() {
	fmt.Println("Init...")
}

type GowitnessScreenshotHandler struct{}

func (h *GowitnessScreenshotHandler) TakeScreenshot(url string, destination string) bool {
	fmt.Println("Taking Screenshot Using GoWitness")

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
		return false
	}
	log.Printf("Command finished with error: %v", err)
	return true

	//chromium-browser --headless --disable-gpu --screenshot https://www.chromestatus.com/ --no-sandbox

}
