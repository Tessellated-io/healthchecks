package health

import (
	"fmt"
	"io/ioutil"
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
	var client = &http.Client{
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
	shouldCreateNewChecks := hc.createNewChecksValue()
	url := fmt.Sprintf("https://hc-ping.com/%s/%s?create=%d", hc.pingKey, slug, shouldCreateNewChecks)
	fmt.Println("url")
	fmt.Println(url)

	resp, err := hc.client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println("got a response")
	fmt.Println(string(body))
	// hc.logger.Info().Str("ressponse", string(body))

	return nil
}

func (hc *healthClient) SendFailure(slug string) error {
	shouldCreateNewChecks := hc.createNewChecksValue()
	url := fmt.Sprintf("https://hc-ping.com/%s/%s/fail?create=%d", hc.pingKey, slug, shouldCreateNewChecks)
	_, err := hc.client.Get(url)
	if err != nil {
		return err
	}
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
