package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/AstraBert/multipilot/components"
	"github.com/AstraBert/multipilot/worker"
	"github.com/a-h/templ"
	"github.com/spf13/cobra"
)

const DefaultConfigFile = "multipilot.config.json"

var configFile string
var showHelp bool

var rootCmd = &cobra.Command{
	Use:   "multipilot",
	Short: "multipilot is a simple orchestration layer to run multiple GitHub Copilot tasks.",
	Long:  "multipilot is a simple orchestration layer based on Temporal workflows designed to run multiple Copilot instances concurrently on different projects",
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
		var failureReasons string
		switch len(reasonsFailed) {
		case 0:
			failureReasons = "\n"
		default:
			failureReasons = "Failure reasons:\n- " + strings.Join(reasonsFailed, "\n- ") + "\n"
		}
		fmt.Printf("Successfull tasks: %d\nFailed tasks: %d\n%s", success, failed, failureReasons)
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

var port int
var host string
var fileToRender string

var renderCmd = &cobra.Command{
	Use:   "render",
	Short: "Render logs from a MultiPilot session",
	Long:  "Render the logs from a MultiPilot session within a HTML file served locally on your browser.",
	Run: func(cmd *cobra.Command, args []string) {
		if fileToRender == "" {
			log.Println("required option `--input/-i` is missing")
			return
		}
		events, err := LoadEvents(fileToRender)
		if err != nil {
			log.Printf("An error occurred while loading the events from the log file: %s\n", err.Error())
			return
		}
		addr := fmt.Sprintf("%s:%d", host, port)
		server := http.NewServeMux()
		component := components.Home(events)
		server.Handle("GET /", templ.Handler(component))
		log.Printf("starting server on :%s\n", addr)

		if err := http.ListenAndServe(addr, server); err != nil {
			log.Fatal(err)
		}
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

	renderCmd.Flags().StringVarP(&fileToRender, "input", "i", "", "File with the JSON log records to render")
	renderCmd.Flags().IntVarP(&port, "port", "p", 8000, "Port where to serve the rendered logs")
	renderCmd.Flags().StringVarP(&host, "bind", "b", "0.0.0.0", "Host where to bind the port for logs rendering")
	_ = renderCmd.MarkFlagRequired("input")

	rootCmd.AddCommand(workerCmd)
	rootCmd.AddCommand(renderCmd)
}
