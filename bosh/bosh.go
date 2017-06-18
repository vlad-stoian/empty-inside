package bosh

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha1"
	"fmt"
	"io"
	"time"

	yaml "gopkg.in/yaml.v2"
)

type ReleaseManifestJob struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Fingerprint string `yaml:"fingerprint"`
	SHA1        string `yaml:"sha1"`
}

type ReleaseManifestPackage struct {
	Name         string   `yaml:"name"`
	Version      string   `yaml:"version"`
	Fingerprint  string   `yaml:"fingerprint"`
	SHA1         string   `yaml:"sha1"`
	Dependencies []string `yaml:"dependencies"`
}

type ReleaseManifest struct {
	Name               string `yaml:"name"`
	Version            string `yaml:"version"`
	CommitHash         string `yaml:"commit_hash"`
	UncommittedChanges bool   `yaml:"uncommitted_changes"`
	Jobs               []ReleaseManifestJob
	Packages           []ReleaseManifestPackage
}

type JobManifest struct {
	Name       string                      `yaml:"name"`
	Packages   []string                    `yaml:"packages,omitempty"`
	Templates  map[string]string           `yaml:"templates,omitempty"`
	Properties map[interface{}]interface{} `yaml:"properties,omitempty"`
}

func GenerateTarHeader(name string, size int, isDir bool) *tar.Header {
	header := new(tar.Header)

	header.Name = name
	header.Size = int64(size)
	header.Gname = "root"
	header.Uname = "root"
	header.ModTime = time.Now()

	if isDir {
		header.Mode = 0755
		header.Typeflag = tar.TypeDir
	} else {
		header.Mode = 0644
		header.Typeflag = tar.TypeReg
	}

	return header
}

func GenerateJobManifest(writer io.Writer, name string) (int, string, error) {
	jobManifest := JobManifest{
		Name: name,
	}

	jobManifestBytes, err := yaml.Marshal(jobManifest)
	if err != nil {
		return -1, "", err
	}

	jobManifestBytes = append([]byte("---\n"), jobManifestBytes...)

	bytesWritten, err := writer.Write(jobManifestBytes)
	if err != nil {
		return -1, "", err
	}

	if bytesWritten != len(jobManifestBytes) {
		return -1, "", fmt.Errorf("")
	}

	sha1sum := fmt.Sprintf("%x", sha1.Sum(jobManifestBytes))
	return bytesWritten, sha1sum, nil
}

func GenerateMonitFile(writer io.Writer) (int, string, error) {
	sha1sum := fmt.Sprintf("%x", sha1.Sum(nil))
	return 0, sha1sum, nil
}

func GenerateJobArchive(writer io.Writer, name string) (string, error) {
	gw := gzip.NewWriter(writer)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	var buffer, fingerprint bytes.Buffer

	fingerprint.WriteString("v2")

	buffer.Reset()
	monitSize, monitFingerprint, _ := GenerateMonitFile(&buffer)
	monitHeader := GenerateTarHeader("./monit", monitSize, false)

	tw.WriteHeader(monitHeader)
	tw.Write(buffer.Bytes())
	fingerprint.WriteString(fmt.Sprintf("monit%s100644", monitFingerprint))

	buffer.Reset()
	jobManifestSize, jobManifestFingerprint, _ := GenerateJobManifest(&buffer, name)
	jobManifestHeader := GenerateTarHeader("./job.MF", jobManifestSize, false)

	tw.WriteHeader(jobManifestHeader)
	tw.Write(buffer.Bytes())
	fingerprint.WriteString(fmt.Sprintf("spec%s100644", jobManifestFingerprint)) // job.MF is still named spec when the fingerprint is computed

	sha1sum := fmt.Sprintf("%x", sha1.Sum(fingerprint.Bytes()))
	return sha1sum, nil
}

func GenerateReleaseManifest(writer io.Writer, releaseManifest ReleaseManifest) (int, error) {
	releaseManifestBytes, err := yaml.Marshal(releaseManifest)
	if err != nil {
		return -1, err
	}

	releaseManifestBytes = append([]byte("---\n"), releaseManifestBytes...)

	bytesWritten, err := writer.Write(releaseManifestBytes)
	if err != nil {
		return -1, err
	}

	if bytesWritten != len(releaseManifestBytes) {
		return -1, fmt.Errorf("")
	}

	return bytesWritten, nil
}

func GenerateReleaseArchive(writer io.Writer, name string, jobs []string) error {
	gw := gzip.NewWriter(writer)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	releaseManifestJobs := []ReleaseManifestJob{}

	jobsDirHeader := GenerateTarHeader("./jobs/", 0, true)

	tw.WriteHeader(jobsDirHeader)

	for _, job := range jobs {
		jobBuffer := new(bytes.Buffer)
		fingerprint, err := GenerateJobArchive(jobBuffer, job)
		if err != nil {
			return err
		}

		jobHeader := GenerateTarHeader(fmt.Sprintf("./jobs/%s.tgz", job), jobBuffer.Len(), false)

		tw.WriteHeader(jobHeader)
		tw.Write(jobBuffer.Bytes())

		releaseManifestJob := ReleaseManifestJob{
			Name:        job,
			Fingerprint: fingerprint,
			Version:     fingerprint,
			SHA1:        fmt.Sprintf("%x", sha1.Sum(jobBuffer.Bytes())),
		}

		releaseManifestJobs = append(releaseManifestJobs, releaseManifestJob)
	}

	releaseManifestBuffer := new(bytes.Buffer)
	releaseManifest := ReleaseManifest{
		Name:       name,
		Version:    "stub-version",
		CommitHash: "deadbeef",
		Jobs:       releaseManifestJobs,
	}
	releaseManifestSize, _ := GenerateReleaseManifest(releaseManifestBuffer, releaseManifest)

	releaseManifestHeader := GenerateTarHeader("./release.MF", releaseManifestSize, false)
	tw.WriteHeader(releaseManifestHeader)
	tw.Write(releaseManifestBuffer.Bytes())

	return nil
}
