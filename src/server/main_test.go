package server

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"deployer/src/config"
)

// syncBuffer is a concurrency-safe buffer for capturing the server error log,
// which the net/http server may write from a separate goroutine.
type syncBuffer struct {
	mu     sync.Mutex
	buffer bytes.Buffer
}

func (syncBuffer *syncBuffer) Write(payload []byte) (int, error) {
	syncBuffer.mu.Lock()
	defer syncBuffer.mu.Unlock()

	return syncBuffer.buffer.Write(payload)
}

func (syncBuffer *syncBuffer) String() string {
	syncBuffer.mu.Lock()
	defer syncBuffer.mu.Unlock()

	return syncBuffer.buffer.String()
}

func newTestServer(t *testing.T, components map[string]config.ComponentConfig) (*httptest.Server, *syncBuffer) {
	t.Helper()

	originalConfig := config.Config

	t.Cleanup(func() {
		config.Config = originalConfig
	})

	config.Config = config.AppConfig{
		Timeout:    30 * time.Second,
		Components: components,
	}

	errorLog := &syncBuffer{}

	server := httptest.NewUnstartedServer(http.HandlerFunc(handler))
	server.Config.ErrorLog = log.New(errorLog, "", 0)
	server.Start()

	t.Cleanup(server.Close)

	return server, errorLog
}

func postDeploy(t *testing.T, serverURL string, form url.Values) (int, string) {
	t.Helper()

	//nolint:gosec // test code with httptest.Server, URL is always internal
	response, err := http.Post(
		serverURL,
		"application/x-www-form-urlencoded",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		t.Fatalf("post deploy: %s", err)
	}

	defer func() {
		_ = response.Body.Close()
	}()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read body: %s", err)
	}

	return response.StatusCode, string(body)
}

func TestHandlerValidationReturnsBadRequest(t *testing.T) {
	components := map[string]config.ComponentConfig{
		"app": {
			Command: []string{"/bin/sh", "-c", "echo hello"},
			Key:     "secret",
		},
	}

	tests := []struct {
		name string
		form url.Values
	}{
		{
			name: "unknown component",
			form: url.Values{"component": {"nope"}},
		},
		{
			name: "invalid component key",
			form: url.Values{"component": {"app"}, "key": {"wrong"}},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			server, errorLog := newTestServer(t, components)

			statusCode, body := postDeploy(t, server.URL, testCase.form)

			if statusCode != http.StatusBadRequest {
				t.Errorf("status: got %d, want %d", statusCode, http.StatusBadRequest)
			}

			if strings.Contains(body, "event:") {
				t.Errorf("expected no sse stream, got body: %q", body)
			}

			assertNoSuperfluousHeader(t, errorLog)
		})
	}
}

func TestHandlerSyncStreamsOutputAndExit(t *testing.T) {
	tests := []struct {
		name         string
		command      []string
		wantExitCode string
		wantEvent    string
		wantMessage  string
	}{
		{
			name:         "successful deploy",
			command:      []string{"/bin/sh", "-c", "echo hello"},
			wantExitCode: `"exit_code":0`,
			wantEvent:    "event: output",
			wantMessage:  "hello",
		},
		{
			name:         "non-zero exit",
			command:      []string{"/bin/sh", "-c", "echo oops 1>&2; exit 2"},
			wantExitCode: `"exit_code":2`,
			wantEvent:    "event: error",
			wantMessage:  "oops",
		},
		{
			name:         "command cannot start",
			command:      []string{"/nonexistent/deployer/binary"},
			wantExitCode: `"exit_code":1`,
			wantEvent:    "event: error",
			wantMessage:  "deploy err",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			components := map[string]config.ComponentConfig{
				"app": {Command: testCase.command},
			}

			server, errorLog := newTestServer(t, components)

			statusCode, body := postDeploy(t, server.URL, url.Values{"component": {"app"}})

			if statusCode != http.StatusOK {
				t.Errorf("status: got %d, want %d", statusCode, http.StatusOK)
			}

			if !strings.Contains(body, testCase.wantEvent) {
				t.Errorf("body missing %q, got: %q", testCase.wantEvent, body)
			}

			if !strings.Contains(body, testCase.wantMessage) {
				t.Errorf("body missing message %q, got: %q", testCase.wantMessage, body)
			}

			if !strings.Contains(body, "event: exit") {
				t.Errorf("body missing exit event, got: %q", body)
			}

			if !strings.Contains(body, testCase.wantExitCode) {
				t.Errorf("body missing %q, got: %q", testCase.wantExitCode, body)
			}

			assertNoSuperfluousHeader(t, errorLog)
		})
	}
}

func assertNoSuperfluousHeader(t *testing.T, errorLog *syncBuffer) {
	t.Helper()

	logged := errorLog.String()
	if strings.Contains(logged, "superfluous response.WriteHeader") {
		t.Errorf("superfluous WriteHeader warning logged: %q", logged)
	}
}
