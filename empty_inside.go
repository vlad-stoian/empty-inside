package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"

	. "github.com/vlad-stoian/empty-inside/bosh"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
	yaml "gopkg.in/yaml.v2"
)

//TODO: Write this utility for a random manifest

var (
	verbose      = kingpin.Flag("verbose", "Verbose mode").Short('v').Bool()
	manifestPath = kingpin.Arg("manifest-path", "Path of the manifest file").Required().String()
)

type Manifest struct {
	InstanceGroups []InstanceGroup `yaml:"instance_groups"`
}

type InstanceGroup struct {
	Name string `yaml:"name"`
	Jobs []Job  `yaml:"jobs"`
}

type Job struct {
	Name    string `yaml:"name"`
	Release string `yaml:"release"`
}

type Deployment struct {
	Releases []*Release
}

type Release struct {
	Name string
	Jobs []string
}

func (r *Release) AddJob(newJobName string) {
	for _, job := range r.Jobs {
		if job == newJobName {
			return
		}
	}

	r.Jobs = append(r.Jobs, newJobName)
}

func (d *Deployment) GetOrCreateRelease(newReleaseName string) *Release {
	for _, release := range d.Releases {
		if release.Name == newReleaseName {
			return release
		}

	}
	newRelease := &Release{Name: newReleaseName}
	d.Releases = append(d.Releases, newRelease)

	return newRelease
}

func CreateManifest(manifestBytes []byte) Manifest {
	var manifest Manifest

	err := yaml.Unmarshal(manifestBytes, &manifest)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return manifest
}

func CreateDeployment(manifest Manifest) Deployment {
	deployment := Deployment{}

	for _, instanceGroup := range manifest.InstanceGroups {
		for _, job := range instanceGroup.Jobs {
			release := deployment.GetOrCreateRelease(job.Release)
			release.AddJob(job.Name)
		}
	}

	return deployment
}

func main() {
	kingpin.Parse()

	fmt.Printf("%v, %s\n", *verbose, *manifestPath)

	manifestBytes, err := ioutil.ReadFile(*manifestPath)
	if err != nil {
		fmt.Print(err)
	}

	manifest := CreateManifest(manifestBytes)

	deployment := CreateDeployment(manifest)

	firstRelease := deployment.Releases[0]

	var releaseArchiveBuffer bytes.Buffer

	GenerateReleaseArchive(&releaseArchiveBuffer, firstRelease.Jobs)

	ioutil.WriteFile("/tmp/crazy-file.tgz", releaseArchiveBuffer.Bytes(), 0777)
}
