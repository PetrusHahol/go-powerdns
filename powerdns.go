package powerdns

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/dghubble/sling"
)

// Error structure with JSON API metadata
type Error struct {
	Status  string
	Message string `json:"error"`
}

func (e Error) Error() string {
	return fmt.Sprintf("%v", e.Message)
}

func handleAPIClientError(slingResponse *http.Response, slingError *error, failureV *Error) error {
	if slingError == nil && slingResponse.StatusCode >= 400 {
		return &Error{
			Status:  slingResponse.Status,
			Message: failureV.Message,
		}
	}
	return nil
}

// PowerDNS configuration structure
type PowerDNS struct {
	Scheme     string
	Hostname   string
	Port       string
	VHost      string
	Headers    map[string]string
	httpClient *http.Client
}

// NewClient initializes a new PowerDNS client configuration
func NewClient(baseURL string, vhost string, headers map[string]string, httpClient *http.Client) *PowerDNS {
	scheme, hostname, port, err := parseBaseURL(baseURL)
	if err != nil {
		log.Fatalf("%s is not a valid url: %v", baseURL, err)
	}

	return &PowerDNS{
		Scheme:     scheme,
		Hostname:   hostname,
		Port:       port,
		VHost:      parseVhost(vhost),
		Headers:    headers,
		httpClient: httpClient,
	}
}

func parseBaseURL(baseURL string) (string, string, string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", "", "", err
	}
	hp := strings.Split(u.Host, ":")
	hostname := hp[0]
	var port string
	if len(hp) > 1 {
		port = hp[1]
	} else {
		if u.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}

	return u.Scheme, hostname, port, nil
}

func parseVhost(vhost string) string {
	if vhost == "" {
		vhost = "localhost"
	}

	return vhost
}

func (p *PowerDNS) makeSling() *sling.Sling {
	u := url.URL{}
	u.Host = p.Hostname + ":" + p.Port
	u.Scheme = p.Scheme
	u.Path = "/api/v1/"

	mySling := sling.New()
	mySling.Base(u.String())

	for key, value := range p.Headers {
		mySling.Set(key, value)
	}

	if p.httpClient != nil {
		mySling.Client(p.httpClient)
	}

	return mySling
}
