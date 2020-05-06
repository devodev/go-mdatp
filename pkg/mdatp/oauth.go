package mdatp

import (
	"fmt"
	"net/url"

	"golang.org/x/oauth2/clientcredentials"
)

var (
	mdatpResourceURL  = "https://graph.windows.net"
	microsoftTokenURL = "https://login.windows.net/%s/oauth2/token"
)

// OAuthConfig .
func OAuthConfig(clientID, clientSecret, tenantID string) *clientcredentials.Config {
	return &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     fmt.Sprintf(microsoftTokenURL, tenantID),
		EndpointParams: url.Values{
			"resource": []string{mdatpResourceURL},
		},
	}
}
