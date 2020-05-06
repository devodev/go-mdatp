package mdatp

import (
	"fmt"

	"golang.org/x/oauth2/clientcredentials"
)

var (
	mdatpScope        = "https://securitycenter.onmicrosoft.com/windowsatpservice/.default"
	microsoftTokenURL = "https://login.windows.net/%s/oauth2/v2.0/token"
)

// OAuthConfig .
func OAuthConfig(clientID, clientSecret, tenantID string) *clientcredentials.Config {
	return &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     fmt.Sprintf(microsoftTokenURL, tenantID),
		Scopes:       []string{mdatpScope},
	}
}
