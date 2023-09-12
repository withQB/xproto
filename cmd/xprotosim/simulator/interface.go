package simulator

import (
	"math"
)

func (sim *Simulator) ReportDistance(a, b string, l int64) {
	sim.distsMutex.Lock()
	defer sim.distsMutex.Unlock()

	if _, ok := sim.dists[a]; !ok {
		sim.dists[a] = map[string]*Distance{}
	}

	if _, ok := sim.dists[a][b]; !ok {
		sim.dists[a][b] = &Distance{}
	}

	sim.dists[a][b].Observed = l
}

func (sim *Simulator) UpdateRealDistances() {
	sim.distsMutex.Lock()
	defer sim.distsMutex.Unlock()

	for from := range sim.nodes {
		for to := range sim.nodes {
			if _, ok := sim.dists[from]; !ok {
				sim.dists[from] = map[string]*Distance{}
			}

			if _, ok := sim.dists[from][to]; !ok {
				sim.dists[from][to] = &Distance{}
			}

			a, _ := sim.graph.GetMapping(from)
			b, _ := sim.graph.GetMapping(to)
			if a != -1 && b != -1 {
				path, err := sim.graph.Shortest(a, b)

				if err == nil {
					sim.dists[from][to].Real = path.Distance
				} else {
					sim.dists[from][to].Real = math.MaxInt64
				}
			}
		}
	}

	sim.State.Act(nil, func() {
		sim.distsMutex.Lock()
		defer sim.distsMutex.Unlock()
		sim.State._updateExpectedBroadcasts(sim.dists)
	})
}
