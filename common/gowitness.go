package common

import (
	"log"
	"os"
	"os/exec"
)

type GowitnessScreenshotHandler struct{}

//DEPRECATED. It has problem with concurrency. When run in parallel, it does not behave normally (Sometimes it doesn't get screenshot)
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
