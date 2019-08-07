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

func (d *Dispatcher) init() {
	//var dispatch Dispatcher
	d.conn, _ = amqp.Dial("amqp://admin:mypass@localhost:5672/")
	//failOnError(err, "Failed to connect to RabbitMQ")
	//defer conn.Close()

	d.ch, _ = d.conn.Channel()
	//failOnError(err, "Failed to open a channel")
	//defer ch.Close()

	d.callbackQ, _ = d.ch.QueueDeclare(
		"callback_queue", //+common.TokenGenerator(5), // name
		true,             // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	//failOnError(err, "Failed to declare a queue")

	fmt.Printf("Initialized...")
}

func SendRequest(url string) {

	//var dispatch Dispatcher
	conn, err := amqp.Dial("amqp://admin:mypass@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	callbackQ, err := ch.QueueDeclare(
		"callback_queue", //+common.TokenGenerator(5), // name
		true,             // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	//failOnError(err, "Failed to declare a queue")

	fmt.Printf("Initialized...")

	msg := common.SSRequest{URL: url}
	body, err := json.Marshal(msg)
	failOnError(err, "Failed to create message object")

	corrId := common.TokenGenerator(5)

	err = ch.Publish(
		"",           // exchange
		"task_queue", // routing key
		false,        // mandatory
		false,
		amqp.Publishing{
			DeliveryMode:  amqp.Persistent,
			ContentType:   "text/plain",
			CorrelationId: corrId,
			ReplyTo:       callbackQ.Name, //callback_queue
			Body:          []byte(body),
		})

	//Now get the response back
	log.Printf("sent a message with correlation_id: %s", corrId)
	msgs, err := ch.Consume(
		callbackQ.Name, // queue
		"",             // consumer
		false,          // auto-ack
		false,          // exclusive
		false,          // no-local
		false,          // no-wait
		nil,            // args
	)
	//log.Printf("^^^^^^^^^^^^^^%s^^^^^^^^^^^^^^^^^^^^^^^",d.callbackQ.Name)
	failOnError(err, "Failed to register a consumer")

	for d := range msgs {

		if corrId == d.CorrelationId {
			response := common.SSResponse{}
			json.Unmarshal(d.Body, &response)
			log.Printf(" [RESULT] URL: %s\nSCREENSHOT: %s", url, response.ImagePath)
			//log.Printf("Received a message: %s", d.Body)
			//log.Printf("\n+++++++ %s The response is: %s ",corrId, response.ImagePath)
			failOnError(err, "Failed to convert body to integer")
			d.Ack(false)
			break
		}
		d.Nack(false, true)

	}
	failOnError(err, "Failed to publish a message")
	//log.Printf("\n ********* [x] Response of %s is received: %s",url, body)

}
func main() {

	//dispatch :=Dispatcher{}
	//dispatch.init()
	var fileName string
	if os.Getenv("INPUT_FILE") != "" {
		fileName = os.Getenv("INPUT_FILE")
	} else {
		fileName = "input.data"
	}

	parseInputFile(fileName)

	//for i:=0;i<5 ;i++  {
	//	go SendRequest("http://google.com")
	//	//go SendRequest("http://ltu.se")
	//}

	// wait for Control+C to quit
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	<-c
}

func parseInputFile(filename string) {
	file, err := os.Open(filename)
	failOnError(err, "Unabe to open the file")
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		rawURL := scanner.Text()
		_, err := url.ParseRequestURI(scanner.Text())
		if err == nil { //If the URL is valid
			//Process it!
			go SendRequest(rawURL)
		} //We do not process invalid URLs.
	}
	failOnError(scanner.Err(), "Scanner got crazy... Watch your file!")
}

//return correct result
//Validate URL
//Remove comments
//Add helpful comment
//Parse File
//Fix ACL
//Do other type
//Write explanation
//Draw diagram
