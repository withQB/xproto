package simulator

import (
	"context"
	"fmt"
	"time"
)

func (sim *Simulator) Ping(from, to string) (uint16, time.Duration, error) {
	fromnode := sim.nodes[from]
	tonode := sim.nodes[to]
	success := false

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	defer func() {
		sim.pathConvergenceMutex.Lock()
		if _, ok := sim.pathConvergence[from]; !ok {
			sim.pathConvergence[from] = map[string]bool{}
		}
		sim.pathConvergence[from][to] = success
		sim.pathConvergenceMutex.Unlock()
	}()

	hops, rtt, err := fromnode.Ping(ctx, tonode.PublicKey())
	if err != nil {
		return 0, 0, fmt.Errorf("fromnode.Ping: %w", err)
	}

	success = true
	sim.ReportDistance(from, to, int64(hops))
	return hops, rtt, nil
}
