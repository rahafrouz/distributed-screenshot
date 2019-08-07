package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/rahafrouz/distributed-screenshot/common"
	"github.com/streadway/amqp"
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

type Dispatcher struct {
	taskQ     amqp.Queue
	callbackQ amqp.Queue
	ch        *amqp.Channel
	conn      *amqp.Connection
	err       error
}

//Initialize by crating connection, and channels. The main logic happens in the main function
func (d *Dispatcher) init() {

	//Connect to rabbitMQ Broker
	ConnectionString := fmt.Sprintf("amqp://%s:%s@%s:%s",
		os.Getenv("RMQ_USER"),
		os.Getenv("RMQ_PASS"),
		os.Getenv("RMQ_BROKER_ADDRESS"),
		os.Getenv("RMQ_BROKER_PORT"),
	)
	//ConnectionString="amqp://admin:mypass@localhost:5672"
	var err error
	d.conn, err = amqp.Dial(ConnectionString)
	failOnError(err, "Failed to connect to RabbitMQ")

	d.ch, err = d.conn.Channel()
	failOnError(err, "Failed to open a channel")

	//The jobs would be sent here
	d.callbackQ, err = d.ch.QueueDeclare(
		"callback_queue", //+common.TokenGenerator(5), // name
		true,             // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	failOnError(err, "Failed to declare a queue")

	fmt.Printf("Initialized...")
}

//Send a request to the broker, so that it would be handed over to a worker eventually.
func SendRequest(url string, d *Dispatcher) {

	msg := common.SSRequest{URL: url}
	body, err := json.Marshal(msg)
	failOnError(err, "Failed to create message object")

	corrId := common.TokenGenerator(5)

	err = d.ch.Publish(
		"",           // exchange
		"task_queue", // routing key
		false,        // mandatory
		false,
		amqp.Publishing{
			DeliveryMode:  amqp.Persistent,
			ContentType:   "text/plain",
			CorrelationId: corrId,
			ReplyTo:       d.callbackQ.Name, //callback_queue
			Body:          []byte(body),
		})

	//Now get the response back
	log.Printf("sent a message with correlation_id: %s", corrId)
	msgs, err := d.ch.Consume(
		d.callbackQ.Name, // queue
		"",               // consumer
		false,            // auto-ack
		false,            // exclusive
		false,            // no-local
		false,            // no-wait
		nil,              // args
	)
	failOnError(err, "Failed to register a consumer")

	//Right now, we have sent the message. We sit here and listen to the callback queue.
	//Whenever a response comes that belongs to us, we ACK it, and we eat it!
	for d := range msgs {

		if corrId == d.CorrelationId {
			response := common.SSResponse{}
			json.Unmarshal(d.Body, &response)
			log.Printf(" [RESULT] URL: %s\nSCREENSHOT: %s", url, response.ImagePath)
			failOnError(err, "Failed to convert body to integer")
			//We got the correct message, ACK it.
			d.Ack(false)
			break
		}
		//The message was not ours. Requeue it
		d.Nack(false, true)

	}
	failOnError(err, "Failed to publish a message")
	//log.Printf("\n ********* [x] Response of %s is received: %s",url, body)

}
func main() {

	dispatch := Dispatcher{}
	dispatch.init()
	var fileName string
	//Is there any input file specified?
	if os.Getenv("INPUT_FILE") != "" {
		fileName = os.Getenv("INPUT_FILE")
	} else {
		fileName = "input.data"
	}

	parseInputFile(fileName, &dispatch)

	// wait for Control+C to quit
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	<-c
}

func parseInputFile(filename string, d *Dispatcher) {
	file, err := os.Open(filename)
	failOnError(err, "Unabe to open the file")
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		rawURL := scanner.Text()
		_, err := url.ParseRequestURI(scanner.Text())
		if err == nil { //If the URL is valid
			//Process it!
			go SendRequest(rawURL, d)
		} //We do not process invalid URLs.
	}
	failOnError(scanner.Err(), "Scanner got crazy... Watch your file!")
}
