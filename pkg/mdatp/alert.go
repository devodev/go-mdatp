package mdatp

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	duration "github.com/ChannelMeter/iso8601duration"
)

var (
	fetchEndpoint  = "alerts"
	datetimeFormat = "2006-01-02T15:04:05.999"
)

// AlertService .
type AlertService service

// AlertRequestParams represents available query params
// to be passed to Fetch.
type AlertRequestParams struct {
	SinceTimeUTC             time.Time
	UntilTimeUTC             time.Time
	Ago                      string
	Limit                    int
	Machinegroups            []string
	DeviceCreatedMachineTags string
	CloudCreatedMachineTags  []string
}

// Values validates attribute values and returns a Values map.
func (p *AlertRequestParams) Values() (url.Values, error) {
	if p == nil {
		return nil, fmt.Errorf("AlertRequestParams is nil")
	}
	if p.Ago != "" {
		if !p.SinceTimeUTC.IsZero() || !p.UntilTimeUTC.IsZero() {
			return nil, fmt.Errorf("Ago and SinceTimeUTC or UntilTimeUTC are provided but Ago and *TimeUTC are mutually exclusives")
		}
		if _, err := duration.FromString(p.Ago); err != nil {
			return nil, err
		}
	}

	values := make(url.Values)
	if !p.SinceTimeUTC.IsZero() {
		values.Add("sinceTimeUtc", p.SinceTimeUTC.Format(datetimeFormat))
	}
	if !p.UntilTimeUTC.IsZero() {
		values.Add("untilTimeUtc", p.UntilTimeUTC.Format(datetimeFormat))
	}
	if p.Ago != "" {
		values.Add("ago", p.Ago)
	}
	if p.Limit != 0 {
		values.Add("limit", strconv.Itoa(p.Limit))
	}
	if len(p.Machinegroups) > 0 {
		for _, group := range p.Machinegroups {
			values.Add("machinegroups", group)
		}
	}
	if p.DeviceCreatedMachineTags != "" {
		values.Add("DeviceCreatedMachineTags", p.DeviceCreatedMachineTags)
	}
	if len(p.CloudCreatedMachineTags) > 0 {
		for _, tag := range p.CloudCreatedMachineTags {
			values.Add("CloudCreatedMachineTags", tag)
		}
	}
	return values, nil
}

// Fetch retrieves alerts using conditions.
func (s *AlertService) Fetch(ctx context.Context, p *AlertRequestParams) (*Response, []Alert, error) {
	values, err := p.Values()
	if err != nil {
		return nil, nil, err
	}
	req, err := s.client.newRequest("GET", fetchEndpoint, values, nil)
	if err != nil {
		return nil, nil, err
	}
	var alerts []Alert
	resp, err := s.client.do(ctx, req, &alerts)
	if err != nil {
		return nil, nil, err
	}
	return resp, alerts, nil
}

// Alert represents a Microsoft Defender ATP Alert type.
type Alert struct {
	Actor                     string `json:"Actor"`
	AlertID                   string `json:"AlertId"`
	AlertPart                 string `json:"AlertPart"`
	AlertTime                 string `json:"AlertTime"`
	AlertTitle                string `json:"AlertTitle"`
	Category                  string `json:"Category"`
	CloudCreatedMachineTags   string `json:"CloudCreatedMachineTags"`
	CommandLine               string `json:"CommandLine"`
	ComputerDNSName           string `json:"ComputerDnsName"`
	CreatorIocName            string `json:"CreatorIocName"`
	CreatorIocValue           string `json:"CreatorIocValue"`
	Description               string `json:"Description"`
	DeviceCreatedMachineTags  string `json:"DeviceCreatedMachineTags"`
	DeviceID                  string `json:"DeviceID"`
	ExternalID                string `json:"ExternalId"`
	FileHash                  string `json:"FileHash"`
	FileName                  string `json:"FileName"`
	FilePath                  string `json:"FilePath"`
	FullID                    string `json:"FullId"`
	IncidentLinkToWDATP       string `json:"IncidentLinkToWDATP"`
	InternalIPv4List          string `json:"InternalIPv4List"`
	InternalIPv6List          string `json:"InternalIPv6List"`
	IoaDefinitionID           string `json:"IoaDefinitionId"`
	IocName                   string `json:"IocName"`
	IocUniqueID               string `json:"IocUniqueId"`
	IocValue                  string `json:"IocValue"`
	IPAddress                 string `json:"IpAddress"`
	LastProcessedTimeUtc      string `json:"LastProcessedTimeUtc"`
	LinkToWDATP               string `json:"LinkToWDATP"`
	LogOnUsers                string `json:"LogOnUsers"`
	MachineDomain             string `json:"MachineDomain"`
	MachineGroup              string `json:"MachineGroup"`
	MachineName               string `json:"MachineName"`
	Md5                       string `json:"Md5"`
	RemediationAction         string `json:"RemediationAction"`
	RemediationIsSuccess      string `json:"RemediationIsSuccess"`
	ReportID                  string `json:"ReportID"`
	Severity                  string `json:"Severity"`
	Sha1                      string `json:"Sha1"`
	Sha256                    string `json:"Sha256"`
	Source                    string `json:"Source"`
	ThreatCategory            string `json:"ThreatCategory"`
	ThreatFamily              string `json:"ThreatFamily"`
	ThreatName                string `json:"ThreatName"`
	URL                       string `json:"Url"`
	UserDomain                string `json:"UserDomain"`
	UserName                  string `json:"UserName"`
	WasExecutingWhileDetected string `json:"WasExecutingWhileDetected"`
}
