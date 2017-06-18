package bosh_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
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

	Describe("Generate Release Archive", func() {
		var (
			fileHeader        *tar.Header
			fileContents      []byte
			fileToBeRead      string
			jobsToBeGenerated []string
			err               error
			buffer            *bytes.Buffer
		)

		BeforeEach(func() {
			buffer = new(bytes.Buffer)
			jobsToBeGenerated = []string{"random-job", "other-random-job"}
		})

		JustBeforeEach(func() {
			err = GenerateReleaseArchive(buffer, "release-name", jobsToBeGenerated)

			Expect(err).To(BeNil())

			fileHeader, fileContents = ReadFileFromArchive(fileToBeRead, buffer)
		})

		It("is a valid .tgz file", func() {
			Expect(fileHeader).To(BeNil())
			Expect(fileContents).To(BeNil())
		})

		Describe("./jobs/ dir", func() {
			BeforeEach(func() {
				fileToBeRead = "./jobs/"
			})

			It("exists", func() {
				Expect(fileHeader).NotTo(BeNil())
				Expect(fileContents).NotTo(BeNil())
			})

			It("is of type dir", func() {
				Expect(fileHeader.Typeflag).To(Equal(byte(tar.TypeDir)))
			})
		})

		Describe("./release.MF", func() {
			BeforeEach(func() {
				fileToBeRead = "./release.MF"
			})

			It("exists", func() {
				Expect(fileHeader).NotTo(BeNil())
				Expect(fileContents).NotTo(BeNil())
			})

			It("contains a release name", func() {
				Expect(string(fileContents)).To(ContainSubstring("name: release-name"))
			})

			It("contains a version", func() {
				Expect(string(fileContents)).To(ContainSubstring("version: stub-version"))
			})

			It("contains a commit_hash", func() {
				Expect(string(fileContents)).To(ContainSubstring("commit_hash: deadbeef"))
			})

			It("contains 2 jobs", func() {
				Expect(string(fileContents)).To(ContainSubstring("name: random-job"))
				Expect(string(fileContents)).To(ContainSubstring("name: other-random-job"))
			})
		})

		for _, jobGenerated := range []string{"random-job", "other-random-job"} {
			jobPath := fmt.Sprintf("./jobs/%s.tgz", jobGenerated)

			Describe(jobPath, func() {
				BeforeEach(func() {
					fileToBeRead = jobPath
				})

				It("exists", func() {
					Expect(fileHeader).NotTo(BeNil())
					Expect(fileContents).NotTo(BeNil())
				})
			})
		}

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
				Expect(fileHeader.Mode).To(Equal(int64(0644)))
				Expect(fileHeader.Typeflag).To(Equal(byte(tar.TypeReg)))
				Expect(fileHeader.Uname).To(Equal("root"))
				Expect(fileHeader.Gname).To(Equal("root"))
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
				Expect(fileHeader.Mode).To(Equal(int64(0644)))
				Expect(fileHeader.Typeflag).To(Equal(byte(tar.TypeReg)))
				Expect(fileHeader.Uname).To(Equal("root"))
				Expect(fileHeader.Gname).To(Equal("root"))
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
