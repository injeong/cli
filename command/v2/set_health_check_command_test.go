package v2_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/v2"
	"code.cloudfoundry.org/cli/command/v2/v2fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("set-health-check Command", func() {
	var (
		cmd             v2.SetHealthCheckCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v2fakes.FakeSetHealthCheckActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v2fakes.FakeSetHealthCheckActor)

		cmd = v2.SetHealthCheckCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		fakeConfig.TargetedOrganizationReturns(configv3.Organization{
			Name: "some-org",
		})
		fakeConfig.TargetedSpaceReturns(configv3.Space{
			GUID: "some-space-guid",
			Name: "some-space",
		})

		fakeConfig.CurrentUserReturns(configv3.User{Name: "some-user"}, nil)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when checking the target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(
				sharedaction.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			config, targetedOrganizationRequired, targetedSpaceRequired := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(config).To(Equal(fakeConfig))
			Expect(targetedOrganizationRequired).To(Equal(true))
			Expect(targetedSpaceRequired).To(Equal(true))

			Expect(executeErr).To(MatchError(
				command.NotLoggedInError{BinaryName: binaryName}))
		})
	})

	Context("when setting the application health check type returns an error", func() {
		var expectedErr error

		BeforeEach(func() {
			cmd.RequiredArgs.AppName = "some-app"
			cmd.RequiredArgs.HealthCheck.Type = "some-health-check-type"

			expectedErr = errors.New("set health check error")
			fakeActor.SetApplicationHealthCheckTypeByNameAndSpaceReturns(
				v2action.Warnings{"warning-1"}, expectedErr)
		})

		It("displays warnings and returns the error", func() {
			Expect(testUI.Err).To(Say("warning-1"))
			Expect(executeErr).To(MatchError(expectedErr))
		})
	})

	Context("when setting health check is successful", func() {
		BeforeEach(func() {
			cmd.RequiredArgs.AppName = "some-app"
			cmd.RequiredArgs.HealthCheck.Type = "some-health-check-type"

			fakeActor.SetApplicationHealthCheckTypeByNameAndSpaceReturns(
				v2action.Warnings{"warning-1"}, nil)
		})

		It("informs the user and displays warnings", func() {
			Expect(testUI.Out).To(Say("Updating health check type to 'some-health-check-type' for app some-app in org some-org / space some-space as some-user..."))
			Expect(testUI.Err).To(Say("warning-1"))
			Expect(testUI.Out).To(Say("OK"))
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(fakeActor.SetApplicationHealthCheckTypeByNameAndSpaceCallCount()).To(Equal(1))
			name, spaceGUID, healthCheckType := fakeActor.SetApplicationHealthCheckTypeByNameAndSpaceArgsForCall(0)
			Expect(name).To(Equal("some-app"))
			Expect(spaceGUID).To(Equal("some-space-guid"))
			Expect(healthCheckType).To(Equal("some-health-check-type"))
		})
	})
})
