package balancer

//go:generate mockgen -source=./balancer.go -destination=../mocks/balancer/balancer.go

import (
	"context"

	"github.com/livepeer/catalyst-api/cluster"
)

type BalancerStub struct {
	config *Config
}

// create a new load balancer instance
func NewBalancerStub(config *Config) Balancer {
	return &BalancerStub{config: config}
}

// start this load balancer instance, execing MistUtilLoad if necessary
func (b *BalancerStub) Start(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (b *BalancerStub) UpdateMembers(members []cluster.Member) error {
	return nil
}

// always returns local node
func (b *BalancerStub) GetBestNode(redirectPrefixes []string, playbackID, lat, lon, fallbackPrefix string) (string, string, error) {
	return "localhost", playbackID, nil
}

func (b *BalancerStub) QueryMistForClosestNodeSource(playbackID, lat, lon, prefix string, source bool) (string, error) {
	return "dtsc://localhost", nil
}