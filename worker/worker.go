package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/rahafrouz/distributed-screenshot/common"
	"github.com/streadway/amqp"
	"log"
	"os"
	"strconv"
	"sync"
)

type Worker struct {
	GOWITNESS_PATH    string
	ch                *amqp.Channel
	q                 amqp.Queue
	svc               *s3manager.Uploader
	readDirMux        sync.Mutex
	screenshotHandler common.ScreenshotHanlder
	storageHandler    common.CloudStorageHandler
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		os.Exit(88)
	}
}

func main() {
	w := Worker{}
	w.InitAndListen()

}

func (w *Worker) InitAndListen() {
	//Setting up engines for screenshot and storage
	w.screenshotHandler = &common.ChromeDPScreenshotHandler{}
	w.storageHandler = &common.S3Storage{}
	log.Printf("Initialized the worker")

	//Not for current implementation
	//w.GOWITNESS_PATH = os.Getenv("GOWITNESS_PATH")

	//Connect to rabbitMQ Broker
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

	//This is the queue of tasks. We take jobs from here to process.
	w.q, err = w.ch.QueueDeclare(
		"task_queue", // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	failOnError(err, "Failed to declare a queue")

	//Number of active screenshot-taking jobs to handle in parallel. It should be tuned by experiment to utilize the hardware.
	workerThreadSize, _ := strconv.Atoi(os.Getenv("WORKER_THREADS"))
	err = w.ch.Qos(
		workerThreadSize, // prefetch count
		0,                // prefetch size
		false,            // global
	)
	failOnError(err, "Failed to set QoS")

	//Consume tasks from the task_queue.
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

	//main loop of worker. Listen for jobs.
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

//Publish the result to the callback_queue. Using the correlation_id, the sender would identify it, and can get it.
func (w Worker) PublishResponse(resultLocation string, d amqp.Delivery) error {
	log.Println("Finished taking screenshot. Sending back the response")

	//building the response envelope
	responseMsg := common.SSResponse{
		Result:    true,
		ImagePath: resultLocation,
	}
	responseMsgBody, err := json.Marshal(responseMsg)
	failOnError(err, "Failed to create message object")

	//Send back the response to the callback queue
	err = w.ch.Publish(
		"",        // exchange
		d.ReplyTo, // routing key: It should be callback_queue
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: d.CorrelationId,
			Body:          []byte(responseMsgBody),
		})

	failOnError(err, "Failed to publish the RESPONSE message")

	return err
}

//Takes the request, build the screenshot as a byte[], and then upload the array to s3.
//No disk Operation involved with this type of handling.
func (w *Worker) handleJobRequest(d amqp.Delivery) (string, error) {

	log.Printf("Received a message: %s", d.Body)

	msg := common.SSRequest{}
	err := json.Unmarshal(d.Body, &msg)
	failOnError(err, "unable to unmarshal message")

	//Here we can pass other options{resolution, viewpoint, etc} for taking screenshot
	//Take the screenshot!
	screenshotData, _ := w.screenshotHandler.TakeScreenshot(msg.URL, "", false)

	//Generate the random token for having unique file names.
	//Suggestion: have versioning here, or some other logic
	magicName := common.TokenGenerator(6)

	magicName = magicName + ".png"

	return w.storageHandler.UploadDataToCloud(magicName, screenshotData)
}
