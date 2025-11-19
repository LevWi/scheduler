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
	baseURL    string
	businessID common.ID
}

// TODO make function swagger.Slot -> common.Slot
func (p *HttpSlotsProvider) AvailableSlotsInRange(ctx context.Context, interval common.Interval) ([]common.Slot, error) {
	u, err := url.JoinPath(p.baseURL, "slots", p.businessID)
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

	err = checkStatusCode(resp)
	if err != nil {
		return nil, err
	}

	//TODO print request id in log
	var slots swagger.AvailableSlots
	err = json.NewDecoder(resp.Body).Decode(&slots)
	if err != nil {
		return nil, fmt.Errorf("http: unexpected response (%s)", resp.Status)
	}

	out := make([]common.Slot, 0, len(slots.Slots))
	for _, slot := range slots.Slots {
		var tmp common.Slot
		tmp.Start = slot.TpStart
		tmp.Dur = time.Minute * time.Duration(slot.Len)
		out = append(out, tmp)
	}
	return out, nil
}
