package mdatp

import "testing"

func TestClientDefault(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("error occured creating client: %v", err)
	}

	if client.httpClient.Timeout != defaultTimeout {
		t.Errorf(
			"timeout is not default value. got: %v want: %v",
			client.httpClient.Timeout, defaultTimeout,
		)
	}
	baseURL := client.BaseURL.String()
	defaultURL := defaultBaseURL.String()
	if baseURL != defaultURL {
		t.Errorf("baseURL is not default value. got: %v want: %v", baseURL, defaultURL)
	}
	version := client.Version()
	if version != defaultVersion {
		t.Errorf("Version is not default value. got: %v want: %v", version, defaultVersion)
	}
}
