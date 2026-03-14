package sonar

import "context"

type DummyProvider struct{}

func (p *DummyProvider) Close(context.Context) error {
	return nil
}
