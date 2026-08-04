package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestRunIncludesServerErrorMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		http.Error(writer, "deploy err: get component config: component not found: missing", http.StatusBadRequest)
	}))
	t.Cleanup(server.Close)

	_, err := Run(context.Background(), logrus.New(), Options{
		URL:       server.URL,
		Component: "missing",
	})

	want := "unexpected status: 400: deploy err: get component config: component not found: missing"
	if err == nil || err.Error() != want {
		t.Fatalf("error: got %v, want %q", err, want)
	}
}

func TestRunKeepsStatusErrorWhenResponseBodyIsEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusBadGateway)
	}))
	t.Cleanup(server.Close)

	_, err := Run(context.Background(), logrus.New(), Options{
		URL:       server.URL,
		Component: "app",
	})

	want := "unexpected status: 502"
	if err == nil || err.Error() != want {
		t.Fatalf("error: got %v, want %q", err, want)
	}
}
