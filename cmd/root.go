package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/AstraBert/multipilot/worker"
	"github.com/spf13/cobra"
)

const DefaultConfigFile = "multipilot.config.json"

var configFile string
var showHelp bool

var rootCmd = &cobra.Command{
	Use:   "mp",
	Short: "mp is a simple orchestration layer to run multiple GitHub Copilot tasks.",
	Long:  "mp (multipilot) is a simple orchestration layer based on Temporal workflows designed to run multiple Copilot instances concurrently on different projects",
	Run: func(cmd *cobra.Command, args []string) {
		if showHelp {
			_ = cmd.Help()
			return
		}
		tasks, err := ReadConfigToTasks(configFile)
		if err != nil {
			log.Println("An error occurred while loading the configuration: ", err)
			return
		}
		var wg sync.WaitGroup
		errChan := make(chan error)
		for _, task := range tasks.Tasks {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := RunCopilotWorkflow(task)
				errChan <- err
			}()
		}
		go func() {
			wg.Wait()
			close(errChan)
		}()

		success := 0
		failed := 0
		reasonsFailed := []string{}

		for e := range errChan {
			if e != nil {
				failed += 1
				reasonsFailed = append(reasonsFailed, e.Error())
			} else {
				success += 1
			}
		}
		fmt.Printf("Successfull tasks: %d\nFailed tasks: %d\nFailure reasons:\n- %s", success, failed, strings.Join(reasonsFailed, "\n- ")+"\n")
	},
}

var workerCmd = &cobra.Command{
	Use:   "start-worker",
	Short: "Start the Temporal worker responsible for the execution of Copilot tasks",
	Long:  "Start the Temporal worker that, polling from the task queue, orchestrates the execution of Copilot tasks",
	Run: func(cmd *cobra.Command, args []string) {
		worker.StartWorker()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Oops. An error while executing scpr '%s'\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&configFile, "config", "c", DefaultConfigFile, "Path to the JSON file where the config for multipilot is stored. Defaults to: multipilot.config.json")
	rootCmd.Flags().BoolVarP(&showHelp, "help", "h", false, "Show the help message and exit.")

	rootCmd.AddCommand(workerCmd)
}
