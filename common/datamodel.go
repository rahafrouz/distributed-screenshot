package datamodel

import (
	"crypto/rand"
	"fmt"
)

//Data Structure of the message envelope for sending request
type SSRequest struct {
	URL string
}

type SSResponse struct {
	Result    bool
	ImagePath string
}

func TokenGenerator(size int) string {
	b := make([]byte, size)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
