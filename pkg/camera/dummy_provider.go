package camera

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"
)

type DummyProvider struct {
	Interval time.Duration
}

func (p *DummyProvider) Open(ctx context.Context) chan Frame {
	c := make(chan Frame)
	interval := p.Interval
	if interval == 0 {
		interval = time.Second
	}

	go func() {
		defer close(c)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			frame := dummyFrame(time.Now())
			select {
			case c <- frame:
			case <-ctx.Done():
				return
			}

			select {
			case <-ticker.C:
			case <-ctx.Done():
				return
			}
		}
	}()

	return c
}

func (p *DummyProvider) Close(context.Context) error {
	return nil
}

func dummyFrame(now time.Time) Frame {
	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="640" height="480" viewBox="0 0 640 480"><rect width="640" height="480" fill="#111827"/><circle cx="%d" cy="240" r="86" fill="#38bdf8"/><text x="24" y="52" font-family="Arial" font-size="28" fill="#e5e7eb">Rover camera dummy</text><text x="24" y="436" font-family="Arial" font-size="22" fill="#cbd5e1">%s</text></svg>`, 160+(now.Second()*5)%320, now.Format(time.RFC3339))
	return Frame{
		Type:        "camera_frame",
		ContentType: "image/svg+xml",
		Data:        base64.StdEncoding.EncodeToString([]byte(svg)),
		Timestamp:   now,
	}
}
