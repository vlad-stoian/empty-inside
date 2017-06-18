package bosh

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha1"
	"fmt"
	"io"
	"strconv"

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

	monitHeader := new(tar.Header)
	monitHeader.Name = "./monit"
	monitHeader.Mode = 100644
	monitHeader.Size = int64(monitSize)

	tw.WriteHeader(monitHeader)
	tw.Write(buffer.Bytes())
	fingerprint.WriteString("monit")
	fingerprint.WriteString(monitFingerprint)
	fingerprint.WriteString(strconv.FormatInt(monitHeader.Mode, 10))

	buffer.Reset()
	jobManifestSize, jobManifestFingerprint, _ := GenerateJobManifest(&buffer, name)

	jobManifestHeader := new(tar.Header)
	jobManifestHeader.Name = "./job.MF"
	jobManifestHeader.Mode = 100644
	jobManifestHeader.Size = int64(jobManifestSize)

	tw.WriteHeader(jobManifestHeader)
	tw.Write(buffer.Bytes())
	fingerprint.WriteString("spec") // job.MF is still named spec when the fingerprint is computed
	fingerprint.WriteString(jobManifestFingerprint)
	fingerprint.WriteString(strconv.FormatInt(jobManifestHeader.Mode, 10))

	sha1sum := fmt.Sprintf("%x", sha1.Sum(fingerprint.Bytes()))
	return sha1sum, nil
}

func GenerateReleaseManifest() error {
	return nil
}
