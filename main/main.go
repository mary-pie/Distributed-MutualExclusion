package main

import (
	"flag"
	nodeL "mutualexclusion-project/alg/lamport/node"
	nodeRA "mutualexclusion-project/alg/ricartAgrawala/node"
	"mutualexclusion-project/alg/tokenBased/coordinator"
	"mutualexclusion-project/alg/tokenBased/node"
)

type config struct {
	portRPC    int
	setup_time int
}

func main() {

	conf := config{portRPC: 9090, setup_time: 10}

	algorithm := flag.String("alg", "", "algoritmo")
	mode := flag.String("node-mode", "", "modalita")
	id := flag.String("node-id", "", "id")
	flag.Parse()

	if *algorithm == "token-centr" {
		if *mode == "coord" {
			coordinator.InitCoord("coordinator", conf.portRPC, conf.setup_time)
		}
		if *mode == "node" {
			node.InitProcess(*id, conf.portRPC, conf.setup_time)
		}
	}
	if *algorithm == "lamport" {
		nodeL.InitPro(*id, conf.portRPC, conf.setup_time)
	}
	if *algorithm == "ricartagrawala" {
		nodeRA.InitProc(*id, conf.portRPC, conf.setup_time)
	}
}
