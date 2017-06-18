package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/vlad-stoian/empty-inside"
)

var _ = Describe("Empty Inside CLI", func() {
	Describe("Create Manifest", func() {
		var (
			manifestBytes []byte
			manifest      Manifest
		)
		BeforeEach(func() {
			manifestBytes = []byte(
				"---\n" +
					"instance_groups:\n" +
					"- name: group-1\n" +
					"  jobs:\n" +
					"  - name: group-1-job-1\n" +
					"    release: group-1-release-1\n" +
					"  - name: group-1-job-2\n" +
					"    release: group-1-release-2\n" +
					"- name: group-2\n" +
					"  jobs:\n" +
					"  - name: group-2-job-1\n" +
					"    release: group-2-release-1\n")
		})

		JustBeforeEach(func() {
			manifest = CreateManifest(manifestBytes)
		})

		It("Unmarshals a manifest correctly", func() {
			Expect(manifest).NotTo(BeNil())
		})

		It("Unmarshals instance_groups correctly", func() {
			Expect(manifest.InstanceGroups).To(HaveLen(2))
			Expect(manifest.InstanceGroups[0].Name).To(Equal("group-1"))
			Expect(manifest.InstanceGroups[1].Name).To(Equal("group-2"))
		})

		It("Unmarshals jobs correctly", func() {
			Expect(manifest.InstanceGroups[0].Jobs).To(HaveLen(2))
			Expect(manifest.InstanceGroups[1].Jobs).To(HaveLen(1))

			Expect(manifest.InstanceGroups[0].Jobs[0].Name).To(Equal("group-1-job-1"))
			Expect(manifest.InstanceGroups[0].Jobs[1].Name).To(Equal("group-1-job-2"))
			Expect(manifest.InstanceGroups[1].Jobs[0].Name).To(Equal("group-2-job-1"))

			Expect(manifest.InstanceGroups[0].Jobs[0].Release).To(Equal("group-1-release-1"))
			Expect(manifest.InstanceGroups[0].Jobs[1].Release).To(Equal("group-1-release-2"))
			Expect(manifest.InstanceGroups[1].Jobs[0].Release).To(Equal("group-2-release-1"))
		})
	})

	Describe("Create Deployment", func() {
		var (
			manifest   Manifest
			deployment Deployment
		)

		BeforeEach(func() {
			manifest = Manifest{
				InstanceGroups: []InstanceGroup{
					InstanceGroup{
						Name: "group-1",
						Jobs: []Job{
							Job{
								Name:    "job-1",
								Release: "release-1",
							},
							Job{
								Name:    "job-2",
								Release: "release-1",
							},
							Job{
								Name:    "job-3",
								Release: "release-1",
							},
						},
					},
					InstanceGroup{
						Name: "group-2",
						Jobs: []Job{
							Job{
								Name:    "job-3",
								Release: "release-1",
							},
							Job{
								Name:    "job-4",
								Release: "release-2",
							},
						},
					},
				},
			}
		})

		JustBeforeEach(func() {
			deployment = CreateDeployment(manifest)
		})

		It("reorganizes the manifest", func() {
			Expect(deployment).NotTo(BeNil())

			Expect(deployment.Releases).To(HaveLen(2))

			Expect(deployment.Releases[0].Name).To(Equal("release-1"))
			Expect(deployment.Releases[0].Jobs).To(HaveLen(3))
			Expect(deployment.Releases[0].Jobs[0]).To(Equal("job-1"))
			Expect(deployment.Releases[0].Jobs[1]).To(Equal("job-2"))
			Expect(deployment.Releases[0].Jobs[2]).To(Equal("job-3"))

			Expect(deployment.Releases[1].Name).To(Equal("release-2"))
			Expect(deployment.Releases[1].Jobs).To(HaveLen(1))
			Expect(deployment.Releases[1].Jobs[0]).To(Equal("job-4"))
		})

	})
})
