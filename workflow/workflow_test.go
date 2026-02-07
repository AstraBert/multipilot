package workflow

import (
	"context"
	"errors"
	"testing"

	"github.com/AstraBert/multipilot/shared"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
)

type UnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	env *testsuite.TestWorkflowEnvironment
}

func (s *UnitTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
}

func (s *UnitTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}

func (s *UnitTestSuite) Test_CopilotWorkflow_RunCopilotFails() {
	s.env.OnActivity(RunCopilot, mock.Anything, mock.Anything).Return(errors.New("activity failure"))
	s.env.ExecuteWorkflow(CopilotWorkflow, shared.CopilotInput{LogFile: "hello.jsonl"})

	s.True(s.env.IsWorkflowCompleted())

	err := s.env.GetWorkflowError()
	s.Error(err)
	var applicationErr *temporal.ApplicationError
	s.True(errors.As(err, &applicationErr))
	s.Equal("activity failure", applicationErr.Error())
}

func (s *UnitTestSuite) Test_CopilotWorkflow_RunCopilotSuccess() {
	s.env.OnActivity(RunCopilot, mock.Anything, mock.Anything).Return(nil)
	s.env.ExecuteWorkflow(CopilotWorkflow, shared.CopilotInput{LogFile: "hello.jsonl"})

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *UnitTestSuite) Test_CopilotWorkflow_CorrectParam() {
	s.env.OnActivity(RunCopilot, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, inpt shared.CopilotInput) error {
			s.Equal("hello.jsonl", inpt.LogFile)
			s.Equal("/test/hello", inpt.Cwd)
			s.Equal("Say hello and exit", inpt.Prompt)
			s.Equal("gpt-5.1", inpt.AiModel)
			return nil
		})
	s.env.ExecuteWorkflow(CopilotWorkflow, shared.CopilotInput{LogFile: "hello.jsonl", Cwd: "/test/hello", Prompt: "Say hello and exit", AiModel: "gpt-5.1"})

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}
