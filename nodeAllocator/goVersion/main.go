package main

import (
	"encoding/json"

	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"
)

type ContainerInfo struct {
	Name     string                 `json:"name"`
	Image    string                 `json:"image"`
	Limits   map[string]interface{} `json:"limits"`
	Requests map[string]interface{} `json:"requests"`
	Cluster  string                 `json:"cluster"`
	Node     string                 `json:"node"`
}

type Workflow struct {
	Filename   string          `json:"filename"`
	Containers []ContainerInfo `json:"containers"`
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func subscribe(byteCh chan<- []byte) {
	err := godotenv.Load("../../env/.env")
	failOnError(err, ".env Load fail")

	ip := os.Getenv("MQ_IP")
	port := os.Getenv("MQ_PORT")
	id := os.Getenv("MQ_ID")
	passwd := os.Getenv("MQ_PASSWD")
	queue := os.Getenv("MQ_RESOURCE_QUE")

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

func allocateNode(container ContainerInfo) ContainerInfo {
	newContainer := ContainerInfo{}
	newContainer.Name = container.Name
	newContainer.Image = container.Image
	newContainer.Limits = container.Limits
	newContainer.Requests = container.Requests

	rand.Seed(time.Now().UnixNano())
	arr := []string{"private", "azure", "aws"}
	randomCloudIdx := rand.Intn(len(arr))
	randomCloudName := arr[randomCloudIdx]
	randomNodeNum := strconv.Itoa(rand.Intn(10) + 1)

	newContainer.Cluster = randomCloudName
	newContainer.Node = randomNodeNum

	return newContainer
}

func updateWorkflow(byteCh <-chan []byte, byteChV2 chan<- []byte) {
	for body := range byteCh {
		var workflow Workflow
		unmarshalErr := json.Unmarshal(body, &workflow)
		failOnError(unmarshalErr, ".unmarshal fail")
		workflowV2 := Workflow{}

		workflowV2.Filename = workflow.Filename
		containers := workflow.Containers

		for _, container := range containers {
			containerV2 := allocateNode(container)
			workflowV2.Containers = append(workflowV2.Containers, containerV2)
		}
		finalResult, _ := json.Marshal(workflowV2)
		byteChV2 <- finalResult
	}
}

func publish(byteChV2 <-chan []byte) {
	for body := range byteChV2 {
		envErr := godotenv.Load("../../env/.env")
		failOnError(envErr, "Failed load .env")

		id := os.Getenv("MQ_ID")
		passwd := os.Getenv("MQ_PASSWD")
		ip := os.Getenv("MQ_IP")
		port := os.Getenv("MQ_PORT")
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

		err = ch.Publish(
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "application/json",
				Body:        body,
			})
		failOnError(err, "Failed to publish a message")
	}
}

func main() {
	byteCh := make(chan []byte)
	byteChV2 := make(chan []byte)
	go subscribe(byteCh)
	go updateWorkflow(byteCh, byteChV2)
	go publish(byteChV2)

	var forever chan struct{}
	<-forever
}
