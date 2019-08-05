package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/rahafrouz/distributed-screenshot/common"
	"github.com/streadway/amqp"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

type Worker struct {
	GOWITNESS_PATH string
	ch             *amqp.Channel
	q              amqp.Queue
}

func failOnError(err error, msg string) {
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

func (w *Worker) InitAndListen() {
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

	err = w.ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	failOnError(err, "Failed to set QoS")
	msgs, err := w.ch.Consume(
		w.q.Name, // queue
		"",       // consumer
		false,    // auto-ack
		false,    // exclusive
		false,    // no-local
		false,    // no-wait
		nil,      // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			msg := datamodel.SSRequest{}
			err := json.Unmarshal(d.Body, &msg)
			failOnError(err, "unable to unmarshal message")
			w.HandleRequest(msg)
			//w.TakeScreenshot(msg.URL,".")
			//dot_count := bytes.Count(d.Body, []byte("."))
			//t := time.Duration(dot_count)
			//time.Sleep(t * time.Second)
			log.Printf("message is:" + msg.URL)

			//d.Ack(false)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever

}

func (w Worker) TakeScreenshot(url string, dest string) bool {
	fmt.Println("Taking Screenshot")
	//args := fmt.Sprintf("--headless --disable-gpu --screenshot %s --no-sandbox",url)
	//args := fmt.Sprintf(" single --url=%s --destination=%s",url,dest)
	cmd := exec.Command(os.Getenv(
		"GOWITNESS_PATH"),
		"single", "--url="+url, "--destination="+dest,
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
func tokenGenerator() string {
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

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

func findFileinDir(path string) string {

	files, err := ioutil.ReadDir(path)
	if err != nil {
		failOnError(err, "cannot open folder on: "+path)
	}

	if len(files) > 1 {
		failOnError(nil, "There are more than one files in the job folder. What happened?")
	}
	for _, f := range files {
		return f.Name()
		//fmt.Println(f.Name())
	}
	return ""
}

func (w Worker) HandleRequest(msg datamodel.SSRequest) {

	fmt.Println("Check if it a request for fresh one, or cache is allowed; Would take screenshot or send cache")
	//Generate the destination path
	MagicPath := tokenGenerator()

	//Create Directory
	failOnError(makeDir(MagicPath), "Unable to create directory")

	//Save Screenshot in the MagicPath
	BasePath := "/usr/src/app/"
	log.Println("Path to save screenshot: %s", BasePath+MagicPath)
	w.TakeScreenshot(msg.URL, BasePath+MagicPath)

	filename := MagicPath + "/" + findFileinDir(MagicPath)
	bucket := os.Getenv("BUCKET_NAME")

	file, err := os.Open(filename)

	failOnError(err, "Failed to open file "+filename)

	defer file.Close()

	//select Region to use.
	conf := aws.Config{
		Region:      aws.String(os.Getenv("AWS_REGION")),
		Credentials: credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), ""),
	}
	conf.WithCredentialsChainVerboseErrors(true)
	sess, err := session.NewSession(&conf)
	failOnError(err, "Problem in setting aws session")
	svc := s3manager.NewUploader(sess)

	fmt.Println("Uploading file to S3...")
	result, err := svc.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filepath.Base(filename)),
		Body:   file,
	})

	failOnError(err, "Error Uploading to S3"+err.Error())
	//Everything was fine so far

	fmt.Println("Result is: " + result.Location)
}

func (w Worker) ListItemsInBucket() {
	sess := session.Must(session.NewSession())
	creds := stscreds.NewCredentials(sess, "myRoleArn")
	svc := s3.New(sess, &aws.Config{Credentials: creds})
	fmt.Println(svc)
}
