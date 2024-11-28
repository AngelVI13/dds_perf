package main

import (
	"fmt"
	"log"
	"os"
	"text/template"

	"github.com/alexflint/go-arg"
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

func createPublisherData(n int) []PublisherData {
	if n <= 0 {
		log.Fatalf("number of publishers must be > 0: but got %d", n)
	}
	var data []PublisherData

	for i := 1; i <= n; i++ {
		data = append(data, PublisherData{
			TopicName: fmt.Sprintf("test%d", i),
			Name:      fmt.Sprintf("car%d", i),
			Value:     i,
		})
	}

	return data
}

func createPublisherProcesses(data []PublisherData, tmpl *template.Template) error {
	tempDir, err := os.MkdirTemp("./", "temp")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	for i, d := range data {
		f, err := os.CreateTemp(tempDir, "pub")
		if err != nil {
			return fmt.Errorf("failed to create temp file %d: %v", i, err)
		}

		err = tmpl.Execute(f, d)
		if err != nil {
			return fmt.Errorf("failed to execute template %q: %v", tmpl.Name(), err)
		}
	}

	return nil
}

func main() {
	var args struct {
		PubNum int `arg:"-n,--num"     default:"100"   help:"Number of publishers to start"`
	}
	arg.MustParse(&args)

	publisherTemplate, err := createTemplate("publisher", "./templates/publisher.py")
	if err != nil {
		log.Fatal(err)
	}

	vcdlTemplate, err := createTemplate("vcdl", "./templates/canoe.vcdl")
	if err != nil {
		log.Fatal(err)
	}

	data := createPublisherData(args.PubNum)

	err = createPublisherProcesses(data, publisherTemplate)
	if err != nil {
		log.Fatal(err)
	}

	err = vcdlTemplate.Execute(os.Stdout, data)
	if err != nil {
		log.Fatalf("failed to execute template %q: %v", vcdlTemplate.Name(), err)
	}
}
