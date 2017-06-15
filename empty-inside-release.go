package main

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
	yaml "gopkg.in/yaml.v2"
)

//TODO: Write this utility for a random manifest

var (
	verbose      = kingpin.Flag("verbose", "Verbose mode").Short('v').Bool()
	manifestPath = kingpin.Arg("manifest-path", "Path of the manifest file").Required().String()
)

type Job struct {
	Name    string `yaml:"name"`
	Release string `yaml:"release"`
}

type Release struct {
	Name string
	Jobs []Job
}

func (r *Release) AddJob(newJobName string) {
	for _, job := range r.Jobs {
		if job.Name == newJobName {
			return
		}
	}

	r.Jobs = append(r.Jobs, Job{Name: newJobName})
}

func (j *Job) createArchive() ([]byte, error) {
	var bytes []byte
	bytes = append(bytes, 1)
	return bytes, nil
}

func (r *Release) createArchive() ([]byte, error) {
	buffer := bytes.NewBuffer(nil)

	bw := bufio.NewWriter(buffer)

	gw := gzip.NewWriter(bw)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	for _, job := range r.Jobs {
		jobBytes, _ := job.createArchive()

		header := new(tar.Header)
		header.Name = fmt.Sprintf("./jobs/%s.tgz", job.Name)
		header.Size = int64(len(jobBytes))
		header.Typeflag = tar.TypeReg
		header.ModTime = time.Now()
		header.Mode = 0777
		fmt.Printf("%#v\n", header)

		if err := tw.WriteHeader(header); err != nil {
			return nil, err
		}

		n, err := tw.Write(jobBytes)
		if err != nil {
			return nil, err
		}
		fmt.Println(n)
	}

	fmt.Printf("%#v\n", tw)

	fmt.Printf("%#v\n", bw)

	fmt.Printf("%#v\n", buffer)

	tw.Flush()
	gw.Flush()
	bw.Flush()

	var buf []byte
	buffer.Write(buf)

	return buf, nil
}

type Deployment struct {
	Releases []*Release
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

type InstanceGroup struct {
	Name string `yaml:"name"`
	Jobs []Job  `yaml:"jobs"`
}

type Manifest struct {
	InstanceGroups []InstanceGroup `yaml:"instance_groups"`
}

func createReleases(manifestPath string) Deployment {
	manifestBytes, err := ioutil.ReadFile(manifestPath) // just pass the file name
	if err != nil {
		fmt.Print(err)
	}

	var manifest Manifest

	err = yaml.Unmarshal(manifestBytes, &manifest)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	manifestBytesAgain, err := yaml.Marshal(&manifest)
	fmt.Printf("%s\n", string(manifestBytesAgain))

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

	deployment := createReleases(*manifestPath)

	for _, release := range deployment.Releases {
		fmt.Printf("%s -> [", release.Name)
		for _, job := range release.Jobs {
			fmt.Printf("%s ", job.Name)
		}
		fmt.Printf("]\n")
	}

	archiveBytes, err := deployment.Releases[0].createArchive()
	if err != nil {
		fmt.Println(err)
	}

	ioutil.WriteFile("/tmp/crazy-file.tgz", archiveBytes, 0777)
}
