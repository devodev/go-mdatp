package mdatp

import (
	"fmt"
	"net/http"
	"net/url"

	"golang.org/x/oauth2/clientcredentials"
)

// Credentials are used by WithOAuthClient.
type Credentials struct {
	ClientID     string
	ClientSecret string
	TenantDomain string
	TenantID     string
}

// OAuthClient returns a HTTP client with a OAuth
// compatible TokenSource.
func (c *Credentials) OAuthClient() *http.Client {
	conf := &clientcredentials.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		TokenURL:     fmt.Sprintf(microsoftTokenURL, c.TenantDomain),
		EndpointParams: url.Values{
			"resource": []string{defaultBaseURLStr},
		},
	}
	httpClient := conf.Client(nil)
	httpClient.Timeout = defaultTimeout
	return httpClient
}
