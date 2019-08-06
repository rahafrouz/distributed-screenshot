package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/streadway/amqp"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

type Worker struct {
	GOWITNESS_PATH    string
	ch                *amqp.Channel
	q                 amqp.Queue
	svc               *s3manager.Uploader
	readDirMux        sync.Mutex
	screenshotHandler datamodel.ScreenshotHanlder
}

func failOnError(err error, msg string) {
	//cmd.Execute()
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		os.Exit(88)
	}
}

func main() {
	w := Worker{}
	//w.TakeScreenshot("http://www.google.com",".")
	w.InitAndListen()

}

func (w *Worker) InitS3() {
	//select Region to use.
	conf := aws.Config{
		Region:      aws.String(os.Getenv("AWS_REGION")),
		Credentials: credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), ""),
	}
	conf.WithCredentialsChainVerboseErrors(true)
	sess, err := session.NewSession(&conf)
	failOnError(err, "Problem in setting aws session")
	w.svc = s3manager.NewUploader(sess)
}
func (w *Worker) InitAndListen() {
	//w.screenshotHandler = datamodel.GowitnessScreenshotHandler{}
	w.screenshotHandler = &datamodel.ChromeDPScreenshotHandler{}
	w.InitS3()
	fmt.Printf("Initialized the worker")

	w.GOWITNESS_PATH = os.Getenv("GOWITNESS_PATH")

	ConnectionString := fmt.Sprintf("amqp://%s:%s@%s:%s",
		os.Getenv("RMQ_USER"),
		os.Getenv("RMQ_PASS"),
		os.Getenv("RMQ_BROKER_ADDRESS"),
		os.Getenv("RMQ_BROKER_PORT"),
	)

	conn, err := amqp.Dial(ConnectionString)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	w.ch, err = conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer w.ch.Close()

	w.q, err = w.ch.QueueDeclare(
		"task_queue", // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	failOnError(err, "Failed to declare a queue")
	workerThreadSize, _ := strconv.Atoi(os.Getenv("WORKER_THREADS"))
	err = w.ch.Qos(
		workerThreadSize, // prefetch count
		0,                // prefetch size
		false,            // global
	)
	failOnError(err, "Failed to set QoS")
	msgs, err := w.ch.Consume(
		"task_queue", // queue
		"",           // consumer
		false,        // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			go func() {
				// Main Loop of the Program. First handle the job
				resultLocation, _ := w.handleJobRequest(d)

				//And the publish the response to the requester!
				w.PublishResponse(resultLocation, d)

				// Ack means that this job won't be delivered to any other worker
				d.Ack(false)
			}()

		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever

}
func (w Worker) PublishResponse(resultLocation string, d amqp.Delivery) bool {
	log.Println("Finished taking screenshot. Sending back the response")

	responseMsg := datamodel.SSResponse{
		Result:    true,
		ImagePath: resultLocation,
	}
	responseMsgBody, err := json.Marshal(responseMsg)
	failOnError(err, "Failed to create message object")

	//Send back the response to the callback queue
	err = w.ch.Publish(
		"",               // exchange
		"callback_queue", // routing key: It should be callback_queue
		false,            // mandatory
		false,            // immediate
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: d.CorrelationId,
			Body:          []byte(responseMsgBody),
		})

	failOnError(err, "Failed to publish the RESPONSE message")
	//Not necessary! Refactor later.
	if err != nil {
		return false
	}
	return true
}

//func (w *Worker) TakeScreenshot(u string, dest string) bool {
//
//
//	//utils.ProcessURL(url.URL{})
//	fmt.Println("Taking Screenshot")
//	//args := fmt.Sprintf("--headless --disable-gpu --screenshot %s --no-sandbox",url)
//	//args := fmt.Sprintf(" single --url=%s --destination=%s",url,dest)
//	cmd := exec.Command(os.Getenv(
//		"GOWITNESS_PATH"),
//		"single", "--url="+url, "--destination="+dest,
//		"--chrome-path=/usr/bin/chromium-browser",
//	)
//	cmd.Stdout = os.Stdout
//	cmd.Stderr = os.Stderr
//	log.Printf("Running command and waiting for it to finish...")
//	err := cmd.Run()
//	if err != nil {
//		return false
//	}
//	log.Printf("Command finished with error: %v", err)
//	return true
//
//	//chromium-browser --headless --disable-gpu --screenshot https://www.chromestatus.com/ --no-sandbox
//}

func makeDir(path string) error {
	_, err := os.Stat(path)

	if os.IsNotExist(err) {
		errDir := os.MkdirAll(path, 0777)
		if errDir != nil {
			return errDir
			log.Fatal(err)
		}
	}
	log.Println("Created Directory", path)
	return nil
}

func (w *Worker) findFileinDir(path string) string {
	// A very dangerous way of doing it! TODO: Refactor
	w.readDirMux.Lock()
	//for {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		failOnError(err, "cannot open folder on: "+path)
	}

	w.readDirMux.Unlock()

	fmt.Println("Number of files in this dir:", path, len(files))
	if len(files) > 1 {
		failOnError(nil, "There are more than one files in the job folder. What happened?")
	}
	for _, f := range files {
		return f.Name()
		//fmt.Println(f.Name())
	}
	return ""
}

func (w *Worker) handleJobRequest(d amqp.Delivery) (string, error) {

	log.Printf("Received a message: %s", d.Body)
	msg := datamodel.SSRequest{}
	err := json.Unmarshal(d.Body, &msg)
	failOnError(err, "unable to unmarshal message")

	fmt.Println("Check if it a request for fresh one, or cache is allowed; Would take screenshot or send cache")
	//Generate the destination path
	b := make([]byte, 25)
	rand.Read(b)
	MagicPath := fmt.Sprintf("%x", b)

	//MagicPath := datamodel.TokenGenerator(25)

	//Create Directory
	failOnError(makeDir(MagicPath), "Unable to create directory")

	//Save Screenshot in the MagicPath
	BasePath := "/usr/src/app/"
	log.Println("Path to save screenshot: %s", BasePath+MagicPath)

	w.screenshotHandler.TakeScreenshot(msg.URL, BasePath+MagicPath)

	//w.TakeScreenshot(msg.URL, BasePath+MagicPath)
	//w.TakeScreenshot2(msg.URL,BasePath+MagicPath)

	filename := BasePath + MagicPath + "/" + w.findFileinDir(BasePath+MagicPath+"/")
	bucket := os.Getenv("BUCKET_NAME")

	file, err := os.Open(filename)

	failOnError(err, "Failed to open file "+filename)

	defer file.Close()

	fmt.Println("Uploading file to S3...", filename)
	result, err := w.svc.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filepath.Base(filename)),
		Body:   file,
	})

	failOnError(err, "Error Uploading to S3")
	//Everything was fine so far

	fmt.Println("Result is: " + result.Location)
	return result.Location, nil
}
