package main

import (
	"encoding/json"
	"gopkg.in/yaml.v3"

	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"fmt"
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
func workflowInfo(path string) {
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
	finalResult, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(finalResult))
	//_ = ioutil.WriteFile(fileName+".json", finalResult, 0644)
}

func main() {
	paths := os.Args[1:]

	for _, path := range paths {
		workflowInfo(path)
	}
}
