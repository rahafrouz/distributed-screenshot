package common

//Data Structure of the message envelope for sending request. It contains the URL of the website to take the screenshot.
type SSRequest struct {
	URL string
}

//The response going from worker to dispatcher, containting the result of job.
type SSResponse struct {
	Result    bool
	ImagePath string
}
