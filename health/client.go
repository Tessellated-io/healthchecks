package health

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/tessellated-io/pickaxe/log"
)

type HealthClient interface {
	UpsertCheck(slug string) error

	SendSuccess(slug string) error
	SendFailure(slug string) error
}

type CreateCheckPayload struct {
	Name     string   `json:"name"`
	Slug     string   `json:"slug"`
	Timeout  int      `json:"timeout"`
	Grace    int      `json:"grace"`
	Unique   []string `json:"unique"`
	Channels string   `json:"channels"`
	ApiKey   string   `json:"api_key"`
}

type healthClient struct {
	pingKey            string
	apiKey             string
	createNewChecks    bool
	timeoutSeconds     int
	gracePeriodSeconds int

	client *http.Client
	logger *log.Logger
}

var _ HealthClient = (*healthClient)(nil)

func NewHealthClient(logger *log.Logger, apiKey, pingKey string, createNewChecks bool, timeoutSeconds, gracePeriodSeconds int) HealthClient {
	prefixedLogger := logger.ApplyPrefix("[HEALTHCHECKS] 🩺")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	return &healthClient{
		pingKey:            pingKey,
		apiKey:             apiKey,
		createNewChecks:    createNewChecks,
		client:             client,
		logger:             prefixedLogger,
		timeoutSeconds:     timeoutSeconds,
		gracePeriodSeconds: gracePeriodSeconds,
	}
}

// HealthCheck Interface

func (hc *healthClient) SendSuccess(slug string) error {
	logger := hc.logger.With("slug", slug)
	logger.Info("sending success")

	if hc.createNewChecks {
		err := hc.UpsertCheck(slug)
		if err != nil {
			return err
		}
	}

	url := fmt.Sprintf("https://hc-ping.com/%s/%s", hc.pingKey, slug)

	resp, err := hc.client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	logger.Debug("got response from success call", "response", string(body), "status", resp.StatusCode)

	return nil
}

func (hc *healthClient) SendFailure(slug string) error {
	logger := hc.logger.With("slug", slug)
	logger.Info("sending failure")

	if hc.createNewChecks {
		err := hc.UpsertCheck(slug)
		if err != nil {
			return err
		}
	}
	url := fmt.Sprintf("https://hc-ping.com/%s/%s/fail", hc.pingKey, slug)

	resp, err := hc.client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	logger.Debug("got response from failure call", "response", string(body), "status", resp.StatusCode)

	return nil
}

func (hc *healthClient) UpsertCheck(slug string) error {
	url := "https://healthchecks.io/api/v3/checks/"
	createCheckPayload := &CreateCheckPayload{
		Name:     slug,
		Slug:     slug,
		Timeout:  hc.timeoutSeconds,
		Grace:    hc.gracePeriodSeconds,
		Unique:   []string{"name", "slug"},
		Channels: "*",
		ApiKey:   hc.apiKey,
	}

	payload, err := json.Marshal(createCheckPayload)
	if err != nil {
		return err
	}

	resp, err := hc.client.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected code from HTTP to %s: HTTP %d", url, resp.StatusCode)
	}
	return nil
}
