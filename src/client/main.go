package client

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
)

// Options holds the configuration for a client deployment request.
type Options struct {
	URL       string
	Component string
	Key       string
	Args      map[string]string
}

type outputEventData struct {
	Message string `json:"message"`
}

type exitEventData struct {
	ExitCode int `json:"exit_code"`
}

// Run sends a deployment request to the deployer server, streams output to stdout,
// and returns the remote process exit code.
func Run(ctx context.Context, logger logrus.FieldLogger, opts Options) (int, error) {
	req, err := buildRequest(ctx, opts)
	if err != nil {
		return 0, err
	}

	resp, err := executeRequest(req)
	if err != nil {
		return 0, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	exitCode, err := processResponse(logger, resp.Body)
	if err != nil {
		return 0, err
	}

	return exitCode, nil
}

func buildRequest(ctx context.Context, opts Options) (*http.Request, error) {
	formData := url.Values{}
	formData.Set("component", opts.Component)

	if opts.Key != "" {
		formData.Set("key", opts.Key)
	}

	for key, value := range opts.Args {
		formData.Set(key, value)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, opts.URL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return req, nil
}

func executeRequest(req *http.Request) (*http.Response, error) {
	httpClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// nolint:gosec // CheckRedirect prevents open redirect attacks
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("post request: %w", err)
	}

	return resp, nil
}

func processResponse(logger logrus.FieldLogger, body io.Reader) (int, error) {
	exitCode := 0

	scanner := bufio.NewScanner(body)

	var currentEvent string

	var currentData string

	for scanner.Scan() {
		line := scanner.Text()
		processLine(logger, line, &currentEvent, &currentData, &exitCode)
	}

	err := scanner.Err()
	if err != nil {
		return 0, fmt.Errorf("scan response: %w", err)
	}

	return exitCode, nil
}

func processLine(logger logrus.FieldLogger, line string, currentEventPtr, currentDataPtr *string, exitCodePtr *int) {
	value, ok := strings.CutPrefix(line, "event: ")
	if ok {
		*currentEventPtr = value
		return
	}

	value, ok = strings.CutPrefix(line, "data: ")
	if ok {
		*currentDataPtr = value
		return
	}

	if line == "" {
		*exitCodePtr = dispatchEvent(logger, *currentEventPtr, *currentDataPtr, *exitCodePtr)
		*currentEventPtr = ""
		*currentDataPtr = ""
	}
}

func dispatchEvent(logger logrus.FieldLogger, eventName string, rawData string, exitCode int) int {
	switch eventName {
	case "output":
		var data outputEventData

		err := json.Unmarshal([]byte(rawData), &data)
		if err != nil {
			logger.WithError(err).Error("unmarshal output event")
		} else {
			fmt.Print(data.Message)
		}

	case "exit":
		var data exitEventData

		err := json.Unmarshal([]byte(rawData), &data)
		if err != nil {
			logger.WithError(err).Error("unmarshal exit event")
		} else {
			exitCode = data.ExitCode
		}
	}

	return exitCode
}
