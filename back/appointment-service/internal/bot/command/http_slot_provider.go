package command

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	swagger "scheduler/appointment-service/api/types"
	common "scheduler/appointment-service/internal"
	"time"
)

type HttpSlotsProvider struct {
	baseURL string
}

func (p *HttpSlotsProvider) AvailableSlotsInRange(ctx context.Context, business_id common.ID, interval common.Interval) (common.Intervals, error) {
	u, err := url.JoinPath(p.baseURL, "slots", business_id)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}

	v := req.URL.Query()
	v.Set("date_start", interval.Start.Format(time.RFC3339))
	v.Set("date_end", interval.End.Format(time.RFC3339))
	req.URL.RawQuery = v.Encode()

	//TODO with timeout?
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		break
	case http.StatusBadRequest:
		return nil, fmt.Errorf("http: %w (%s)", common.ErrInvalidArgument, resp.Status)
	default:
		return nil, fmt.Errorf("http: unexpected response (%s)", resp.Status)
	}

	//TODO print request id in log
	var slots swagger.AvailableSlots
	err = json.NewDecoder(resp.Body).Decode(&slots)
	if err != nil {
		return nil, fmt.Errorf("http: unexpected response (%s)", resp.Status)
	}

	out := make(common.Intervals, 0, len(slots.Slots))
	for _, slot := range slots.Slots {
		var tmp common.Interval
		tmp.Start = slot.TpStart
		tmp.End = tmp.Start.Add(time.Minute * time.Duration(slot.Len))
		out = append(out, tmp)
	}
	return out, nil
}
