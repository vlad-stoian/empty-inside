package bosh_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/vlad-stoian/empty-inside/bosh"
)

var _ = Describe("Bosh Tests", func() {

	Describe("Generate Release Manifest", func() {
		var (
			buffer            *bytes.Buffer
			releaseToGenerate ReleaseManifest
			bytesWritten      int
			err               error
		)

		BeforeEach(func() {
			releaseToGenerate = ReleaseManifest{
				Name: "random-release",
			}

			buffer = new(bytes.Buffer)
		})

		JustBeforeEach(func() {
			bytesWritten, err = GenerateReleaseManifest(buffer, releaseToGenerate)
			Expect(err).To(BeNil())
		})

		It("should contain yaml ---", func() {
			Expect(buffer.String()).To(ContainSubstring("---"))
		})

		It("should contain release name", func() {
			Expect(buffer.String()).To(ContainSubstring("name: random-release"))
		})

		It("should have size != -1", func() {
			Expect(bytesWritten).NotTo(Equal(-1))
		})
	})

	Describe("Generate Job Manifest", func() {
		var (
			fingerprint  string
			bytesWritten int
			err          error
			buffer       *bytes.Buffer
		)

		BeforeEach(func() {
			buffer = new(bytes.Buffer)
		})

		JustBeforeEach(func() {
			bytesWritten, fingerprint, err = GenerateJobManifest(buffer, "random-job-name")
			Expect(err).To(BeNil())
		})

		It("should contain yaml ---", func() {
			Expect(buffer.String()).To(ContainSubstring("---"))
		})

		It("should contain job name", func() {
			Expect(buffer.String()).To(ContainSubstring("name: random-job-name"))
		})

		It("should not contain other fields", func() {
			Expect(buffer.String()).NotTo(ContainSubstring("packages"))
			Expect(buffer.String()).NotTo(ContainSubstring("templates"))
			Expect(buffer.String()).NotTo(ContainSubstring("properties"))
		})

		It("should have size != -1", func() {
			Expect(bytesWritten).NotTo(Equal(-1))
		})

		It("should return the correct sha1 sum", func() {
			Expect(fingerprint).To(Equal("3b4346b4c483e8cae92e019acdae42243a8bee11"))
		})
	})

	Describe("Generate Monit File", func() {
		var (
			fingerprint  string
			bytesWritten int
			err          error
			buffer       *bytes.Buffer
		)

		BeforeEach(func() {
			buffer = new(bytes.Buffer)
		})

		JustBeforeEach(func() {
			bytesWritten, fingerprint, err = GenerateMonitFile(buffer)

			Expect(err).To(BeNil())
		})

		It("does nothing", func() {
			Expect(buffer.Len()).To(Equal(0))
		})

		It("should have size != -1", func() {
			Expect(bytesWritten).NotTo(Equal(-1))
		})

		It("should return the correct sha1 sum", func() {
			Expect(string(fingerprint)).To(Equal("da39a3ee5e6b4b0d3255bfef95601890afd80709"))
		})
	})

	Describe("Generate Job Archive", func() {
		var (
			fileHeader   *tar.Header
			fileContents []byte
			fileToBeRead string
			fingerprint  string
			err          error
			buffer       *bytes.Buffer
		)

		BeforeEach(func() {
			buffer = new(bytes.Buffer)
		})

		JustBeforeEach(func() {
			fingerprint, err = GenerateJobArchive(buffer, "random-job-name")

			Expect(err).To(BeNil())

			fileHeader, fileContents = ReadFileFromArchive(fileToBeRead, buffer)
		})

		It("is a valid .tgz file", func() {
			Expect(fileHeader).To(BeNil())
			Expect(fileContents).To(BeNil())
		})

		It("returns the correct fingerprint", func() {
			Expect(fingerprint).To(Equal("a5aee13168a3aac83734f8bbb1d292fe704aef2f"))

		})

		Describe("./job.MF", func() {
			BeforeEach(func() {
				fileToBeRead = "./job.MF"
			})

			It("exists", func() {
				Expect(fileHeader).NotTo(BeNil())
				Expect(fileContents).NotTo(BeNil())
			})

			It("has the correct permissions", func() {
				Expect(fileHeader.Mode).To(Equal(int64(100644)))
				Expect(fileHeader.Typeflag).To(Equal(byte(0)))
				// Expect(fileHeader.Uname).To(Equal("root"))
				// Expect(fileHeader.Gname).To(Equal("root"))
			})

			It("is not empty", func() {
				Expect(fileHeader.Size).To(Equal(int64(26)))
				Expect(fileContents).To(HaveLen(26))
			})
		})

		Describe("./monit", func() {
			BeforeEach(func() {
				fileToBeRead = "./monit"
			})

			It("exists", func() {
				Expect(fileHeader).NotTo(BeNil())
				Expect(fileContents).NotTo(BeNil())
			})

			It("has the correct permissions", func() {
				Expect(fileHeader.Mode).To(Equal(int64(100644)))
				Expect(fileHeader.Typeflag).To(Equal(byte(0)))
				// Expect(fileHeader.Uname).To(Equal("root"))
				// Expect(fileHeader.Gname).To(Equal("root"))
			})

			It("is empty", func() {
				Expect(fileHeader.Size).To(Equal(int64(0)))
				Expect(fileContents).To(HaveLen(0))
			})
		})
	})
})

func ReadFileFromArchive(fileName string, buffer *bytes.Buffer) (*tar.Header, []byte) {
	gr, liberr := gzip.NewReader(buffer)
	Expect(liberr).To(BeNil())
	tr := tar.NewReader(gr)

	for true {
		header, err := tr.Next()

		if err == io.EOF {
			break
		}

		if header.Name == fileName {
			contents, _ := ioutil.ReadAll(tr)
			return header, contents
		}
	}

	return nil, nil
}
