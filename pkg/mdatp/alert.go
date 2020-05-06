package mdatp

import (
	"context"
	"net/url"
)

var (
	fetchEndpoint  = "alerts"
	datetimeFormat = "2006-01-02T15:04:05.999"
)

// AlertService .
type AlertService service

// List retrieves alerts using conditions.
func (s *AlertService) List(ctx context.Context, odataQueryFilter string) (*Response, *AlertListResponse, error) {
	queryParams := url.Values{}
	if odataQueryFilter != "" {
		queryParams.Set("$filter", odataQueryFilter)
	}
	req, err := s.client.newRequest("GET", fetchEndpoint, queryParams, nil)
	if err != nil {
		return nil, nil, err
	}
	var alert *AlertListResponse
	resp, err := s.client.do(ctx, req, &alert)
	return resp, alert, err
}

// AlertListResponse represents a JSON Object returned by
// the List Alerts endpoint.
type AlertListResponse struct {
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
