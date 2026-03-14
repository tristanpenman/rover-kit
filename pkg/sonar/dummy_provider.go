package sonar

import "context"

type DummyProvider struct{}

func (p *DummyProvider) Sample(context.Context) (Reading, error) {
	return Reading{}, nil
}

func (p *DummyProvider) Close(context.Context) error {
	return nil
}
