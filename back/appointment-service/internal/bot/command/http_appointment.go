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

	//TODO with timeout?
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	return checkStatusCode(resp)
}
