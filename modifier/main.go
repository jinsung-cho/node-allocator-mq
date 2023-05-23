package main

import (
	"encoding/json"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
)

type ContainerInfo struct {
	Name         string                 `json:"name"`
	Image        string                 `json:"image"`
	Limits       map[string]interface{} `json:"limits"`
	Requests     map[string]interface{} `json:"requests"`
	NodeSelector map[string]interface{} `json:"nodeSelector"`
}

type Workflow struct {
	Filename   string          `json:"filename"`
	OriginPath string          `json:"originPath"`
	Containers []ContainerInfo `json:"containers"`
}

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

func yaml2json(yamlFile []byte) []byte {
	var data map[string]interface{}
	yamlErr := yaml.Unmarshal(yamlFile, &data)
	failOnError(yamlErr, "Failed unmarshal yaml")

	jsonBytes, jsonMarshalErr := json.Marshal(data)
	failOnError(jsonMarshalErr, "Failed Marshal json")

	return jsonBytes
}

func json2yaml(jsonFile []byte) []byte {
	var data map[string]interface{}
	_ = json.Unmarshal(jsonFile, &data)

	yamlData, _ := yaml.Marshal(data)
	return yamlData
}

func modifyWorkflow(byteCh <-chan []byte) {
	for body := range byteCh {
		var modifiedWorkflow Workflow
		unmarshalErr := json.Unmarshal(body, &modifiedWorkflow)
		failOnError(unmarshalErr, "Unmarshal")

		// Origin JSON file
		yamlFile, readErr := ioutil.ReadFile(modifiedWorkflow.OriginPath)
		failOnError(readErr, "Failed to read file")
		jsonData := yaml2json(yamlFile)
		templates := gjson.Get(string(jsonData), "spec.templates").Value().([]interface{})

		var modifiedContainerInfo []ContainerInfo
		modifiedContainerInfo = modifiedWorkflow.Containers
		for _, container := range templates {
			containerMap, _ := container.(map[string]interface{})
			for _, mdContainer := range modifiedContainerInfo {
				if containerMap["name"] == mdContainer.Name {
					containerMap["nodeSelector"] = mdContainer.NodeSelector
				}
			}
		}
		var tmp map[string]interface{}
		_ = yaml.Unmarshal(yamlFile, &tmp)
		tmp["spec"].(map[string]interface{})["templates"] = templates

		yamlData, _ := yaml.Marshal(tmp)
		_ = ioutil.WriteFile("./modify_result/"+modifiedWorkflow.Filename+".yaml", yamlData, 0644)
	}
}

func main() {
	byteCh := make(chan []byte)
	go subscribe(byteCh)
	go modifyWorkflow(byteCh)

	var forever chan struct{}
	<-forever
}
