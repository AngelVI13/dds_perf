package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
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

func createPublisherProcesses(
	data []PublisherData,
	tmpl *template.Template,
	pythonExec string,
) ([]*exec.Cmd, string, error) {
	tempDir, err := os.MkdirTemp("./", "temp")
	if err != nil {
		return nil, "", fmt.Errorf("failed to create temp dir: %v", err)
	}

	var processes []*exec.Cmd
	var step int
	if len(data) < 10 {
		step = 1
	} else {
		step = len(data) / 10
	}

	for i, d := range data {
		f, err := os.CreateTemp(tempDir, "dds_pub")
		if err != nil {
			return nil, "", fmt.Errorf("failed to create temp file %d: %v", i, err)
		}

		err = tmpl.Execute(f, d)
		if err != nil {
			return nil, "", fmt.Errorf(
				"failed to execute template %q: %v",
				tmpl.Name(),
				err,
			)
		}

		cmd := exec.Command(pythonExec, f.Name())
		err = cmd.Start()
		if err != nil {
			log.Printf("failed to start process: %d (%q): %v", i, f.Name(), err)
			continue
		}

		processes = append(processes, cmd)
		if i%step == 0 {
			log.Printf("started n=%d", i+1)
		}
	}

	return processes, tempDir, nil
}

func createVCDL(data []PublisherData, tmpl *template.Template, filename string) error {
	os.Remove(filename)
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create vcdl file %q: %v", filename, err)
	}
	defer f.Close()

	err = tmpl.Execute(f, data)
	if err != nil {
		log.Fatalf("failed to execute template %q: %v", tmpl.Name(), err)
	}

	return nil
}

func main() {
	var args struct {
		PubNum     int    `arg:"-n,--num"     default:"10"          help:"Number of publishers to start"`
		PythonExec string `arg:"-p,--py"      default:"python"      help:"Python executable path"`
		VCDL       string `arg:"-v,--vcdl"    default:"perf.vcdl"   help:"Name of vCDL file to generate."`
		MultiVCDL  string `arg:"-u,--multi-vcdl"    default:"multi_perf.vcdl"   help:"Name of multi vCDL file to generate."`
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

	multiVcdlTemplate, err := createTemplate("vcdl", "./templates/multi_canoe.vcdl")
	if err != nil {
		log.Fatal(err)
	}

	data := createPublisherData(args.PubNum)

	pubProcesses, tempDir, err := createPublisherProcesses(
		data,
		publisherTemplate,
		args.PythonExec,
	)
	if err != nil {
		log.Fatal(err)
	}

	err = createVCDL(data, vcdlTemplate, args.VCDL)
	if err != nil {
		log.Fatal(err)
	}

	err = createVCDL(data, multiVcdlTemplate, args.MultiVCDL)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("all processes are running. Press enter to stop")
	fmt.Scanf("%s\n")
	log.Println("cleaning up")

	for _, p := range pubProcesses {
		err := p.Process.Kill()
		if err != nil {
			log.Printf("failed to kill process: %s: %v", strings.Join(p.Args, " "), err)
		}
	}

	os.RemoveAll(tempDir)
}
