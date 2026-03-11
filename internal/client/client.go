package client

import (
	"net/http"
	"strings"
	"time"

	"github.com/junghoonkye/toss-investment-cli/internal/session"
)

const defaultAPIBaseURL = "https://wts-api.tossinvest.com"
const defaultInfoBaseURL = "https://wts-info-api.tossinvest.com"
const defaultCertBaseURL = "https://wts-cert-api.tossinvest.com"

type Config struct {
	HTTPClient  *http.Client
	APIBaseURL  string
	InfoBaseURL string
	CertBaseURL string
	Session     *session.Session
}

type Client struct {
	httpClient  *http.Client
	apiBaseURL  string
	infoBaseURL string
	certBaseURL string
	session     *session.Session
}

func New(cfg Config) *Client {
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}

	apiBaseURL := strings.TrimRight(cfg.APIBaseURL, "/")
	infoBaseURL := strings.TrimRight(cfg.InfoBaseURL, "/")
	certBaseURL := strings.TrimRight(cfg.CertBaseURL, "/")
	if apiBaseURL == "" {
		apiBaseURL = defaultAPIBaseURL
	}
	if infoBaseURL == "" {
		infoBaseURL = defaultInfoBaseURL
	}
	if certBaseURL == "" {
		certBaseURL = defaultCertBaseURL
	}

	return &Client{
		httpClient:  httpClient,
		apiBaseURL:  apiBaseURL,
		infoBaseURL: infoBaseURL,
		certBaseURL: certBaseURL,
		session:     cfg.Session,
	}
}
