package simulator

import (
	"github.com/RyanCarrier/dijkstra"
)

func (sim *Simulator) GenerateNetworkGraph() {
	sim.log.Println("Building graph")
	sim.graph = dijkstra.NewGraph()
	sim.maps = make(map[string]int)
	for n := range sim.nodes {
		sim.maps[n] = sim.graph.AddMappedVertex(n)
	}
	for a, aa := range sim.wires {
		for b := range aa {
			if err := sim.graph.AddMappedArc(a, b, 1); err != nil {
				panic(err)
			}
			if err := sim.graph.AddMappedArc(b, a, 1); err != nil {
				panic(err)
			}
		}
	}
}
