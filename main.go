package main

import (
	"fmt"
	"log"
	"os"
	"text/template"
)

type PublisherData struct {
	TopicName string
	Name      string
	Value     int
}

func createTemplate(name, templateFilename string) (*template.Template, error) {
	b, err := os.ReadFile(templateFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to read %q: %v", templateFilename, err)
	}

	templ, err := template.New(name).Parse(string(b))
	if err != nil {
		return nil, fmt.Errorf(
			"failed to parse template %q (file %q): %v",
			name,
			templateFilename,
			err,
		)
	}
	return templ, nil
}

func main() {
	publisherTemplate, err := createTemplate("publisher", "./templates/publisher.py")
	if err != nil {
		log.Fatal(err)
	}

	vcdlTemplate, err := createTemplate("vcdl", "./templates/canoe.vcdl")
	if err != nil {
		log.Fatal(err)
	}

	data := PublisherData{
		TopicName: "test1",
		Name:      "name1",
		Value:     1.0,
	}
	err = publisherTemplate.Execute(os.Stdout, data)
	if err != nil {
		log.Fatalf("failed to execute template %q: %v", publisherTemplate.Name(), err)
	}

	err = vcdlTemplate.Execute(os.Stdout, []PublisherData{data})
	if err != nil {
		log.Fatalf("failed to execute template %q: %v", vcdlTemplate.Name(), err)
	}
}
