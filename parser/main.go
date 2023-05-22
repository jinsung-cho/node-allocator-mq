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
	"path/filepath"
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

// For get work's specific resource information
func specResource(r map[string]interface{}, name string) string {
	_, exists := r[name]
	if exists {
		return r[name].(string)
	}
	return ""
}

// For get work's resource information
func resource(resource map[string]interface{}) (map[string]interface{}, map[string]interface{}) {
	cSpec, containerMarshalErr := json.Marshal(resource)
	failOnError(containerMarshalErr, "Failed Marshal json")

	requestSpec := gjson.Get(string(cSpec), "resources.requests").Value().(map[string]interface{})
	limitSpec := gjson.Get(string(cSpec), "resources.limits").Value().(map[string]interface{})

	return requestSpec, limitSpec
}

func yaml2json(yamlFile []byte) map[string]interface{} {
	// Unmarshal the YAML data
	var data map[string]interface{}
	yamlErr := yaml.Unmarshal(yamlFile, &data)
	failOnError(yamlErr, "Failed unmarshal yaml")

	// Convert the desired values to JSON
	jsonData := make(map[string]interface{})
	jsonData["wolkflow"] = data["spec"].(map[string]interface{})["templates"]

	// Marshal the JSON data to string
	jsonBytes, jsonMarshalErr := json.Marshal(jsonData)
	failOnError(jsonMarshalErr, "Failed Marshal json")

	// Access the value
	var templates map[string]interface{}
	jsonUnmarshalErr := json.Unmarshal(jsonBytes, &templates)
	failOnError(jsonUnmarshalErr, "Failed UnMarshal json")

	return templates
}

// For get workflow information
func workflowInfo(path string) []byte {
	fileFullName := filepath.Base(path)
	fileName := fileFullName[:len(fileFullName)-len(filepath.Ext(fileFullName))]
	// Read the YAML file
	yamlFile, readErr := ioutil.ReadFile(path)
	failOnError(readErr, "Failed to read file")

	templates := yaml2json(yamlFile)

	result := Workflow{}
	result.Filename = fileName
	result.Containers = []ContainerInfo{}
	// Check the workflow in turn and extract the contents for each work
	workflowList := templates["wolkflow"].([]interface{})
	for _, work := range workflowList {
		workInfo := work.(map[string]interface{})

		// Check whether the content of the index is related to the container
		_, existContainer := workInfo["container"]

		if existContainer {
			// Initializing ContainerInfo struct
			containerInfo := ContainerInfo{}
			containerInfo.Name = workInfo["name"].(string)
			containerSpec := workInfo["container"].(map[string]interface{})
			containerInfo.Image = containerSpec["image"].(string)

			// Check if the value for resource setting exists
			// Save the value if it exists
			_, existResource := containerSpec["resources"]
			if existResource {
				request, limit := resource(containerSpec)
				containerInfo.Requests = request
				containerInfo.Limits = limit
			}
			result.Containers = append(result.Containers, containerInfo)
		}
	}
	finalResult, _ := json.Marshal(result)
	return finalResult
}

func publish(b []byte) {
	envErr := godotenv.Load("../env/.env")
	failOnError(envErr, "Failed load .env")
	id := os.Getenv("MQ_ID")
	passwd := os.Getenv("MQ_PASSWD")
	ip := os.Getenv("MQ_IP")
	port := os.Getenv("MQ_PORT")
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

	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        b,
		})
	failOnError(err, "Failed to publish a message")
	log.Printf(" [x] Sent %s\n", b)
}

func main() {
	paths := os.Args[1:]

	for _, path := range paths {
		result := workflowInfo(path)
		publish(result)
	}
}
