package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
	yaml "gopkg.in/yaml.v2"
)

//TODO: Write this utility for a random manifest

var (
	verbose       = kingpin.Flag("verbose", "Verbose mode").Short('v').Bool()
	handcraftPath = kingpin.Arg("handcraft-path", "Path of the handcraft file").Required().String()
)

type Template struct {
	Name    string `yaml:"name"`
	Release string `yaml:"release"`
}

type JobType struct {
	Name          string     `yaml:"name"`
	Label         string     `yaml:"label"`
	ResourceLabel string     `yaml:"resource_label"`
	Description   string     `yaml:"description"`
	Errand        bool       `yaml:"errand"`
	Templates     []Template `yaml:"templates"`
}

type Handcraft struct {
	JobTypes []JobType `yaml:"job_types"`
}

func createReleaseGraph(handcraftPath string) map[string][]string {
	handcraftBytes, err := ioutil.ReadFile(handcraftPath) // just pass the file name
	if err != nil {
		fmt.Print(err)
	}

	var handcraft Handcraft

	err = yaml.Unmarshal(handcraftBytes, &handcraft)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	handcraftBytesAgain, err := yaml.Marshal(&handcraft)
	fmt.Printf("%s\n", string(handcraftBytesAgain))

	graph := make(map[string][]string)

	for _, jobType := range handcraft.JobTypes {
		for _, template := range jobType.Templates {
			graph[template.Release] = append(graph[template.Release], template.Name)

		}
	}

	return graph
}

func main() {
	kingpin.Parse()

	fmt.Printf("%v, %s\n", *verbose, *handcraftPath)

	graph := createReleaseGraph(*handcraftPath)

	for release := range graph {
		fmt.Printf("%s -> [%s]\n", release, strings.Join(graph[release], ", "))
	}
}
