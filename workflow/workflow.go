package workflow

import (
	"time"

	"github.com/AstraBert/multipilot/shared"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const CopilotTaskQueue string = "copilot-task-queue"

func CopilotWorkflow(ctx workflow.Context, input shared.CopilotInput) error {

	// RetryPolicy specifies how to automatically handle retries if an Activity fails.
	retrypolicy := &temporal.RetryPolicy{
		InitialInterval:    20 * time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    100 * time.Second,
		MaximumAttempts:    10, // 0 is unlimited retries
	}

	options := workflow.ActivityOptions{
		// Timeout options specify when to automatically timeout Activity functions.
		StartToCloseTimeout: 60 * time.Minute,
		// Optionally provide a customized RetryPolicy.
		// Temporal retries failed Activities by default.
		RetryPolicy: retrypolicy,
	}

	// Apply the options.
	ctx = workflow.WithActivityOptions(ctx, options)

	// Run Copilot
	var output error

	activityError := workflow.ExecuteActivity(ctx, RunCopilot, input).Get(ctx, &output)
	if activityError != nil {
		return activityError
	}
	if output != nil {
		return output
	}
	return nil
}
