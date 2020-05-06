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
func (s *AlertService) Fetch(ctx context.Context, p *AlertRequestParams) (*Response, *AlertResponse, error) {
	values, err := p.Values()
	if err != nil {
		return nil, nil, err
	}
	req, err := s.client.newRequest("GET", fetchEndpoint, values, nil)
	if err != nil {
		return nil, nil, err
	}
	var alert *AlertResponse
	resp, err := s.client.do(ctx, req, &alert)
	return resp, alert, err
}

// AlertResponse represents a JSON Object returned by
// the List Alerts endpoint.
type AlertResponse struct {
	ODataContext string `json:"@odata.context"`
	Value        []Alert
}

// Alert represents a Microsoft Defender ATP Alert type.
type Alert struct {
	ID                 *string        `json:"id"`
	Title              *string        `json:"title"`
	Description        *string        `json:"description"`
	AlertCreationTime  *string        `json:"alertCreationTime"`
	LastEventTime      *string        `json:"lastEventTime"`
	FirstEventTime     *string        `json:"firstEventTime"`
	LastUpdateTime     *string        `json:"lastUpdateTime"`
	ResolvedTime       *string        `json:"resolvedTime"`
	IncidentID         *int           `json:"incidentId"`
	InvestigationID    *int           `json:"investigationId"`
	InvestigationState *string        `json:"investigationState"`
	AssignedTo         *string        `json:"assignedTo"`
	Severity           *string        `json:"severity"`
	Status             *string        `json:"status"`
	Classification     *string        `json:"classification"`
	Determination      *string        `json:"determination"`
	Category           *string        `json:"category"`
	DetectionSource    *string        `json:"detectionSource"`
	ThreatFamilyName   *string        `json:"threatFamilyName"`
	MachineID          *string        `json:"machineId"`
	Comments           []AlertComment `json:"comments"`
}

// AlertComment is an object contained in Alert.
type AlertComment struct {
	Comment     *string `json:"comment"`
	CreatedBy   *string `json:"createdBy"`
	CreatedTime *string `json:"createdTime"`
}
