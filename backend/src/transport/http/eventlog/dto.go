package eventlog

import (
	"net/url"
	"strconv"
	"strings"
	"time"

	appeventlog "activist-base/src/application/eventlog"
	domaineventlog "activist-base/src/domain/eventlog"
	"activist-base/src/domain/shared"
)

type listRequest struct {
	EventType   *domaineventlog.EventType
	SubjectType *domaineventlog.SubjectType
	SubjectID   *string
	Limit       int
	Offset      int
}

type listResponse struct {
	Items  []eventItem `json:"items"`
	Total  int         `json:"total"`
	Limit  int         `json:"limit"`
	Offset int         `json:"offset"`
}

type eventItem struct {
	ID          string                 `json:"id"`
	EventType   string                 `json:"event_type"`
	ActorID     string                 `json:"actor_id"`
	SubjectType string                 `json:"subject_type"`
	SubjectID   string                 `json:"subject_id"`
	Payload     map[string]string      `json:"payload"`
	Timestamp   string                 `json:"timestamp"`
	CreatedAt   *string                `json:"created_at,omitempty"`
	Meta        map[string]interface{} `json:"meta,omitempty"`
}

func parseListRequest(query url.Values) (listRequest, error) {
	allowed := map[string]struct{}{
		"event_type":   {},
		"subject_type": {},
		"subject_id":   {},
		"limit":        {},
		"offset":       {},
	}
	for key := range query {
		if _, ok := allowed[key]; !ok {
			return listRequest{}, &shared.Error{
				Code:    "validation.query",
				Message: "unsupported query parameter: " + key,
			}
		}
	}

	request := listRequest{
		Limit:  50,
		Offset: 0,
	}

	if raw := strings.TrimSpace(query.Get("event_type")); raw != "" {
		typ := domaineventlog.EventType(raw)
		if !typ.IsValid() {
			return listRequest{}, &shared.Error{
				Code:    "validation.event_type",
				Message: "event_type is invalid",
			}
		}
		request.EventType = &typ
	}
	if raw := strings.TrimSpace(query.Get("subject_type")); raw != "" {
		typ := domaineventlog.SubjectType(raw)
		if !typ.IsValid() {
			return listRequest{}, &shared.Error{
				Code:    "validation.subject_type",
				Message: "subject_type is invalid",
			}
		}
		request.SubjectType = &typ
	}
	if raw := strings.TrimSpace(query.Get("subject_id")); raw != "" {
		if err := shared.RequireText("eventlog.subject_id", raw, shared.MaxEventSubjectID); err != nil {
			return listRequest{}, err
		}
		request.SubjectID = &raw
	}
	if raw := strings.TrimSpace(query.Get("limit")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value <= 0 {
			return listRequest{}, &shared.Error{
				Code:    "validation.limit",
				Message: "limit must be positive integer",
			}
		}
		request.Limit = value
	}
	if raw := strings.TrimSpace(query.Get("offset")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 {
			return listRequest{}, &shared.Error{
				Code:    "validation.offset",
				Message: "offset must be zero or positive",
			}
		}
		request.Offset = value
	}

	normalized, err := appeventlog.NormalizeListInput(request.toInput())
	if err != nil {
		return listRequest{}, err
	}
	request.Limit = normalized.Limit
	request.Offset = normalized.Offset
	return request, nil
}

func (r listRequest) toInput() appeventlog.ListInput {
	return appeventlog.ListInput{
		EventType:   r.EventType,
		SubjectType: r.SubjectType,
		SubjectID:   r.SubjectID,
		Limit:       r.Limit,
		Offset:      r.Offset,
	}
}

func newListResponse(result appeventlog.ListResult) listResponse {
	items := make([]eventItem, 0, len(result.Items))
	for _, entry := range result.Items {
		items = append(items, eventItem{
			ID:          string(entry.ID),
			EventType:   string(entry.EventType),
			ActorID:     string(entry.ActorID),
			SubjectType: string(entry.SubjectType),
			SubjectID:   entry.SubjectID,
			Payload:     entry.Payload,
			Timestamp:   entry.Timestamp.UTC().Format(time.RFC3339),
		})
	}

	return listResponse{
		Items:  items,
		Total:  result.Total,
		Limit:  result.Limit,
		Offset: result.Offset,
	}
}
