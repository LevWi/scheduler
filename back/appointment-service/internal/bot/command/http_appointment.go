package command

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	swagger "scheduler/appointment-service/api/types"
	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/bot"
	"time"
)

type HttpAppointment struct {
	Connection *bot.SchedulerConnection
}

func (a *HttpAppointment) AddSlots(ctx context.Context, customer common.ID, slots []common.Slot) error {
	u, err := url.JoinPath(a.Connection.URL, "slots/bt")
	if err != nil {
		return err
	}

	jsonSlots := make([]swagger.Slot, 0, len(slots))
	for _, s := range slots {
		jsonSlots = append(jsonSlots, swagger.Slot{TpStart: s.Start, Len: int32(s.Dur / time.Minute)})
	}

	b, err := json.Marshal(jsonSlots)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", u, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Client-ID", a.Connection.ClientId)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.Connection.Token))

	q := req.URL.Query()
	q.Add("customer_id", customer)
	req.URL.RawQuery = q.Encode()

	// TODO with timeout
	// https://github.com/LevWi/scheduler/issues/19
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	return checkStatusCode(resp)
}

// TODO make function swagger.Slot -> common.Slot
func (p *HttpAppointment) AvailableSlotsInRange(ctx context.Context, interval common.Interval) ([]common.Slot, error) {
	u, err := url.JoinPath(p.Connection.URL, "slots", p.Connection.BusinessID)
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
