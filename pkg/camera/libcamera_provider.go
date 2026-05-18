package camera

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"time"
)

type LibcameraProvider struct {
	CommandPath string
	Interval    time.Duration
	Timeout     time.Duration
	Width       int
	Height      int
}

func (p *LibcameraProvider) Open(ctx context.Context) chan Frame {
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
			if ctx.Err() != nil {
				return
			}

			frame, err := p.Capture(ctx)
			if err != nil {
				log.Printf("camera capture failed: %v", err)
			} else {
				select {
				case c <- frame:
				case <-ctx.Done():
					return
				}
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

func (p *LibcameraProvider) Close(context.Context) error {
	return nil
}

func (p *LibcameraProvider) Capture(ctx context.Context) (Frame, error) {
	commandPath := p.CommandPath
	if commandPath == "" {
		commandPath = "libcamera-still"
	}

	timeout := p.Timeout
	if timeout == 0 {
		timeout = time.Second
	}

	captureCtx, cancel := context.WithTimeout(ctx, timeout+2*time.Second)
	defer cancel()

	args := []string{
		"--nopreview",
		"--timeout", strconv.Itoa(int(timeout.Milliseconds())),
		"--encoding", "jpg",
		"--output", "-",
	}
	if p.Width > 0 {
		args = append(args, "--width", strconv.Itoa(p.Width))
	}
	if p.Height > 0 {
		args = append(args, "--height", strconv.Itoa(p.Height))
	}

	var stderr bytes.Buffer
	cmd := exec.CommandContext(captureCtx, commandPath, args...)
	cmd.Stderr = &stderr

	output, err := cmd.Output()
	if err != nil {
		return Frame{}, fmt.Errorf("capture frame: %w: %s", err, stderr.String())
	}

	return Frame{
		Type:        "camera_frame",
		ContentType: "image/jpeg",
		Data:        base64.StdEncoding.EncodeToString(output),
		Timestamp:   time.Now(),
	}, nil
}
