package command

import (
	"fmt"
	"net/http"
	common "scheduler/appointment-service/internal"
)

func checkStatusCode(resp *http.Response) error {
	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusBadRequest:
		return fmt.Errorf("http: %w (%s)", common.ErrInvalidArgument, resp.Status)
	default:
		return fmt.Errorf("http: unexpected response (%s)", resp.Status)
	}
}
