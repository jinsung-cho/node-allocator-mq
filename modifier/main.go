package main

import (
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"

	"log"
	"os"
)

type ContainerInfo struct {
	Name     string                 `json:"name"`
	Image    string                 `json:"image"`
	Limits   map[string]interface{} `json:"limits"`
	Requests map[string]interface{} `json:"requests"`
}

type Workflow struct {
	Filename   string          `json:"filename"`
	Containers []ContainerInfo `json:"containers"`
}

// Handle Error
func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func subscribe(byteCh chan<- []byte) {
	err := godotenv.Load("../env/.env")
	failOnError(err, ".env Load fail")

	ip := os.Getenv("MQ_IP")
	port := os.Getenv("MQ_PORT")
	id := os.Getenv("MQ_ID")
	passwd := os.Getenv("MQ_PASSWD")
	queue := os.Getenv("MQ_NODE_QUE")

	conn, err := amqp.Dial("amqp://" + id + ":" + passwd + "@" + ip + ":" + port)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		queue, // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)

	failOnError(err, "Failed to declare a queue")
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	for msg := range msgs {
		body := msg.Body
		byteCh <- body
	}

	var forever chan struct{}
	<-forever
	return
}

func UpdateWorkflow(byteCh <-chan []byte) {
	log.Println(byteCh)
}

func main() {
	byteCh := make(chan []byte)
	go subscribe(byteCh)
	go UpdateWorkflow(byteCh)
}
