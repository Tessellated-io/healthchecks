package health

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/tessellated-io/pickaxe/log"
)

type HealthClient interface {
	SendSuccess(slug string) error
	SendFailure(slug string) error
}

type healthClient struct {
	pingKey         string
	createNewChecks bool

	client *http.Client
	logger *log.Logger
}

var _ HealthClient = (*healthClient)(nil)

func NewHealthClient(logger *log.Logger, pingKey string, createNewChecks bool) HealthClient {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	return &healthClient{
		pingKey:         pingKey,
		createNewChecks: createNewChecks,
		client:          client,
		logger:          logger,
	}
}

// HealthCheck Interface

func (hc *healthClient) SendSuccess(slug string) error {
	hc.logger.Info().Str("slug", slug).Msg("sending success")

	shouldCreateNewChecks := hc.createNewChecksValue()
	url := fmt.Sprintf("https://hc-ping.com/%s/%s?create=%d", hc.pingKey, slug, shouldCreateNewChecks)

	resp, err := hc.client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	hc.logger.Debug().Str("response", string(body)).Int("status", resp.StatusCode).Str("slug", slug).Msg("got response from success call")

	return nil
}

func (hc *healthClient) SendFailure(slug string) error {
	hc.logger.Info().Str("slug", slug).Msg("sending failure")

	shouldCreateNewChecks := hc.createNewChecksValue()
	url := fmt.Sprintf("https://hc-ping.com/%s/%s/fail?create=%d", hc.pingKey, slug, shouldCreateNewChecks)

	resp, err := hc.client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	hc.logger.Debug().Str("response", string(body)).Int("status", resp.StatusCode).Str("slug", slug).Msg("got response from failure call")

	return nil
}

// Private methods

func (hc *healthClient) createNewChecksValue() int {
	if hc.createNewChecks {
		return 1
	} else {
		return 0
	}
}
