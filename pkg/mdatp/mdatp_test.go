package mdatp

import "testing"

func TestClientDefault(t *testing.T) {
	tenantID := ""

	client := NewClient(tenantID)
	if client == nil {
		t.Fatal("something went terribly wrong")
	}

	if client.httpClient.Timeout != defaultTimeout {
		t.Errorf(
			"timeout is not default value. got: %v want: %v",
			client.httpClient.Timeout, defaultTimeout,
		)
	}
	baseURL := client.BaseURL.String()
	if baseURL != defaultBaseURLStr {
		t.Errorf("baseURL is not default value. got: %v want: %v", baseURL, defaultBaseURLStr)
	}
	version := client.Version()
	if version != defaultVersion {
		t.Errorf("Version is not default value. got: %v want: %v", version, defaultVersion)
	}
}
