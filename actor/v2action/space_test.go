package v2action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Space Actions", func() {
	var (
		actor                     Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		fakeConfig := new(v2actionfakes.FakeConfig)
		actor = NewActor(fakeCloudControllerClient, nil, fakeConfig)
	})

	Describe("GetOrganizationSpaces", func() {
		Context("when there are spaces in the org", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpacesReturns(
					[]ccv2.Space{
						{
							GUID:     "space-1-guid",
							Name:     "space-1",
							AllowSSH: true,
						},
						{
							GUID:     "space-2-guid",
							Name:     "space-2",
							AllowSSH: false,
						},
					},
					ccv2.Warnings{"warning-1", "warning-2"},
					nil)
			})

			It("returns all spaces and all warnings", func() {
				spaces, warnings, err := actor.GetOrganizationSpaces("some-org-guid")

				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				Expect(spaces).To(Equal(
					[]Space{
						{
							GUID:     "space-1-guid",
							Name:     "space-1",
							AllowSSH: true,
						},
						{
							GUID:     "space-2-guid",
							Name:     "space-2",
							AllowSSH: false,
						},
					}))

				Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetSpacesArgsForCall(0)).To(Equal(
					[]ccv2.Query{
						{
							Filter:   "organization_guid",
							Operator: ":",
							Value:    "some-org-guid",
						},
					}))
			})
		})

		Context("when an error is encountered", func() {
			var returnedErr error

			BeforeEach(func() {
				returnedErr = errors.New("cc-get-spaces-error")
				fakeCloudControllerClient.GetSpacesReturns(
					[]ccv2.Space{},
					ccv2.Warnings{"warning-1", "warning-2"},
					returnedErr,
				)
			})

			It("returns the error and all warnings", func() {
				_, warnings, err := actor.GetOrganizationSpaces("some-org-guid")

				Expect(err).To(MatchError(returnedErr))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})
	})

	Describe("GetSpaceByName", func() {
		Context("when the space exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpacesReturns(
					[]ccv2.Space{
						{
							GUID:     "some-space-guid",
							Name:     "some-space",
							AllowSSH: true,
						},
					},
					ccv2.Warnings{"warning-1", "warning-2"},
					nil)
			})

			It("returns the space and all warnings", func() {
				space, warnings, err := actor.GetSpaceByName("some-org-guid", "some-space")

				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				Expect(space).To(Equal(Space{
					GUID:     "some-space-guid",
					Name:     "some-space",
					AllowSSH: true,
				}))

				Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetSpacesArgsForCall(0)).To(ConsistOf(
					[]ccv2.Query{
						{
							Filter:   "organization_guid",
							Operator: ":",
							Value:    "some-org-guid",
						},
						{
							Filter:   "name",
							Operator: ":",
							Value:    "some-space",
						},
					}))
			})
		})

		Context("when an error is encountered", func() {
			var returnedErr error

			BeforeEach(func() {
				returnedErr = errors.New("cc-get-spaces-error")
				fakeCloudControllerClient.GetSpacesReturns(
					[]ccv2.Space{},
					ccv2.Warnings{"warning-1", "warning-2"},
					returnedErr,
				)
			})

			It("return the error and all warnings", func() {
				_, warnings, err := actor.GetSpaceByName("some-org-guid", "some-space")

				Expect(err).To(MatchError(returnedErr))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})

		Context("when the space does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpacesReturns(
					[]ccv2.Space{},
					nil,
					nil,
				)
			})

			It("returns SpaceNotFoundError", func() {
				_, _, err := actor.GetSpaceByName("some-org-guid", "some-space")

				Expect(err).To(MatchError(SpaceNotFoundError{
					SpaceName: "some-space",
				}))
			})
		})

		Context("when multiple spaces exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpacesReturns(
					[]ccv2.Space{
						{
							GUID:     "some-space-guid",
							Name:     "some-space",
							AllowSSH: true,
						},
						{
							GUID:     "another-space-guid",
							Name:     "another-space",
							AllowSSH: true,
						},
					},
					nil,
					nil,
				)
			})

			It("returns MultipleSpacesFoundError", func() {
				_, _, err := actor.GetSpaceByName("some-org-guid", "some-space")

				Expect(err).To(MatchError(MultipleSpacesFoundError{
					OrgGUID:   "some-org-guid",
					SpaceName: "some-space",
				}))
			})
		})
	})
})
