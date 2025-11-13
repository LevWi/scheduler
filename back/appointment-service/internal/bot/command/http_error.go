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
		return fmt.Errorf("http response: %w (%s)", common.ErrInvalidArgument, resp.Status)
	case http.StatusUnauthorized:
		return fmt.Errorf("http response: %w (%s)", common.ErrUnauthorized, resp.Status)
	default:
		return fmt.Errorf("http response: unexpected response (%s)", resp.Status)
	}
}
