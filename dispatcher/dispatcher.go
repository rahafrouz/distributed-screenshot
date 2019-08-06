package main

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"os"
	"strings"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

//type Dispatcher struct{}

func main() {
	conn, err := amqp.Dial("amqp://admin:mypass@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"callback_queue", // name
		true,             // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	failOnError(err, "Failed to declare a queue")
	msg := datamodel.SSRequest{URL: "http://www.ltu.se"}
	body, err := json.Marshal(msg)
	failOnError(err, "Failed to create message object")

	corrId := datamodel.TokenGenerator(32)

	err = ch.Publish(
		"",           // exchange
		"task_queue", // routing key
		false,        // mandatory
		false,
		amqp.Publishing{
			DeliveryMode:  amqp.Persistent,
			ContentType:   "text/plain",
			CorrelationId: corrId,
			ReplyTo:       q.Name, //callback_queue
			Body:          []byte(body),
		})

	msgs, err := ch.Consume(
		"callback_queue", // queue
		"",               // consumer
		true,             // auto-ack
		false,            // exclusive
		false,            // no-local
		false,            // no-wait
		nil,              // args
	)

	failOnError(err, "Failed to register a consumer")

	for d := range msgs {
		fmt.Println("received a message")
		if corrId == d.CorrelationId {
			response := datamodel.SSResponse{}
			json.Unmarshal(d.Body, response)
			log.Println("The response is: ", response.ImagePath)
			failOnError(err, "Failed to convert body to integer")
			break
		}
	}
	failOnError(err, "Failed to publish a message")
	log.Printf(" [x] Sent %s", body)
}

func bodyFrom(args []string) string {
	var s string
	if (len(args) < 2) || os.Args[1] == "" {
		s = "hello"
	} else {
		s = strings.Join(args[1:], " ")
	}
	return s
}

func Init() {
	fmt.Printf("Initialized...")
}

func SendJob(url string) {
	fmt.Println("Sending a job")
}
