package worker

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type Event struct {
	Event string
	Data  string
}

type Conn struct {
	events <-chan Event
}

func Dial(endpoint string) (*Conn, error) {
	// Connect once immediately to ensure the server is reachable
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	slog.Debug("Connecting to worker api")
	events, err := dial(ctx, endpoint+"/api/v1/jobs")
	if err != nil {
		return nil, err
	}

	aggregatedEvents := make(chan Event)
	go func() {
		defer close(aggregatedEvents)

		events := events
		var err error

		for {
			for event := range events {
				aggregatedEvents <- event
			}

			// Once the events channel is closed (connection closed), reconnect
			slog.Debug("Reconnecting to worker api")
			events, err = dial(ctx, endpoint+"/api/v1/jobs")
			if err != nil {
				// TODO: Fallback?
				slog.Warn("Failed to connect to worker API")
				time.Sleep(5 * time.Second)
				// Fallthrough
			}
		}
	}()

	return &Conn{
		events: aggregatedEvents,
	}, nil
}

func (c *Conn) Read() (Event, error) {
	event, ok := <-c.events
	if !ok {
		return Event{}, fmt.Errorf("closed")
	}

	return event, nil
}

func (c *Conn) Close() error {
	return fmt.Errorf("not implemented")
}

func dial(ctx context.Context, endpoint string) (<-chan Event, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "text/event-stream")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	events := make(chan Event)
	go func() {
		defer res.Body.Close()
		defer close(events)

		// TODO: Fault-tolerant parsing
		scanner := bufio.NewScanner(res.Body)
		for {
			var event Event
			for scanner.Scan() {
				line := scanner.Text()
				k, v, ok := strings.Cut(line, ": ")
				if ok {
					switch k {
					case "event":
						event.Event = v
					case "data":
						event.Data = v
					}
				} else {
					events <- event
				}
			}
		}
	}()

	return events, nil
}
