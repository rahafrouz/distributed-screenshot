package datamodel

//Data Structure of the message envelope for sending request
type SSRequest struct {
	URL string
}

type SSResponse struct {
	Result    bool
	ImagePath string
}
