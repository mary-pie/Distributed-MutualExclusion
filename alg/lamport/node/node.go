package node

import (
	"log"
	"math/rand"
	"mutualexclusion-project/alg/lamport/node/utils"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
	"time"

	"github.com/hashicorp/memberlist"
)

/*
Funzione di inizializzazione e esecuzione del processo
*/
func InitPro(hostname string, port int, time_setup int) {

	ml := joinGroup(hostname)                   //join al gruppo
	state := utils.NewState(hostname, port, ml) //init stato

	go func() {

		log.Println("Waiting for setup group..")
		time.Sleep(time.Duration(time_setup) * time.Second) //attesa fine setup del gruppo
		log.Println("Group creation terminated. Members are: ", ml.NumMembers())

		//ogni processo invia la propria richiesta dopo un certo intervallo di tempo randomico

		max := 15
		min := 5
		x := rand.Intn(max-min) + min
		time.Sleep(time.Duration(x) * time.Second)

		trying(state) //trying protocol
		if state.GetNextRequest().Sender == hostname && state.GetAcks() == len(state.GetMembers())-1 {
			enterCS()   //entrata CS
			exit(state) //exit protocol
		}

	}()
	initServerRpc(state) //init server RPC

}

/*
Funzione di Join al gruppo di mutua esclusione:
inizializza un server memberlist in ascolto sulla 7946
*/
func joinGroup(id string) *memberlist.Memberlist {
	nodes := []string{"node0"}

	config := memberlist.DefaultLANConfig()
	config.Name = id

	ml, err := memberlist.Create(config)

	if err != nil {
		panic(err)
	}

	_, err = ml.Join(nodes)
	if err != nil {
		panic("Failed to join cluster: " + err.Error())
	}

	log.Println("Group Join is successfully terminated")
	return ml
}

/*
Funzione di inizializzazione del server RPC
*/
func initServerRpc(s *utils.State) {
	me := new(MutualExclusion)

	server := rpc.NewServer()
	err := server.RegisterName("MutualExclusion", me)
	if err != nil {
		log.Fatal("Format of service Mutual Exclusion is not correct: ", err)
	}

	lis, err := net.Listen("tcp", ":"+strconv.Itoa(s.GetPort()))
	if err != nil {
		log.Fatal("Listen error: ", err)
	}
	server.HandleHTTP("/", "/debug")

	me.state = s
	log.Printf("RPC server is up. Listening on port %d\n", s.GetPort())
	err = http.Serve(lis, nil)
	if err != nil {
		log.Fatal("Serve error: ", err)
	}
}

/*
Funzione per effettuare una chiamata RPC di richiesta per accedere alla sezione critica:
quando riceve ACK incrementa il # di acks ricevuti
*/
func sendRequest(ml []*memberlist.Node, req RequestPayload, state *utils.State) {
	log.Println("Sending request to other nodes: ", req)
	for _, m := range ml {
		if m.Name != state.GetHostname() {
			addr := m.Addr.String() + ":9090"
			client, err := rpc.DialHTTP("tcp", addr)
			if err != nil {
				log.Fatal("Error in dialing: ", err)
			}

			ack := new(AckMsg)
			call := client.Go("MutualExclusion.AccessCS", req, &ack, nil)
			call = <-call.Done

			log.Println("ACK received from: ", m.Name)
			//aggiornamento clock = max{ack.Timestamp, clock}
			if maxClock(ack.Timestamp, state.GetClock()) {
				state.SetClock(req.Timestamp)
			}
			state.IncreaseClock() //clock += 1
			log.Println("Updated Clock: ", state.GetClock())
			state.IncreaseAcks() //num_acks += 1

			if call.Error != nil {
				log.Fatal("Error in Access Critical Section: ", call.Error.Error())
			}
			client.Close()
		}
	}
}

/*
Funzione per effettuare una chiamata RPC di RELEASE dopo l'uscita dalla sezione critica:
*/
func sendRelease(ml []*memberlist.Node, rel ReleasePayload) {
	log.Println("Sending RELEASE to other nodes")
	for _, m := range ml {
		if m.Name != rel.SenderId {
			addr := m.Addr.String() + ":9090"
			client, err := rpc.DialHTTP("tcp", addr)
			if err != nil {
				log.Fatal("Error in dialing: ", err)
			}
			var reply string
			call := client.Go("MutualExclusion.Release", rel, &reply, nil)

			if call.Error != nil {
				log.Fatal("Error in Access Critical Section: ", call.Error.Error())
			}
			client.Close()
		}
	}
}

/*
Istruzioni che precedono l'accesso alla sezione critica
*/
func trying(state *utils.State) {
	state.IncreaseClock() // clock += 1
	/*
			TEST RICHIESTE CONCORRENTI
		sendRequest(state.GetMembers(), RequestPayload{SenderId: state.GetHostname(), Timestamp: 1}, state)
		state.AddRequest(utils.Request{Sender: state.GetHostname(), Timestamp: 1})
	*/
	state.AddRequest(utils.Request{Sender: state.GetHostname(), Timestamp: state.GetClock()})                          // Aggiungo la mia richiesta nella coda
	sendRequest(state.GetMembers(), RequestPayload{SenderId: state.GetHostname(), Timestamp: state.GetClock()}, state) //Invio la richiesta agli altri nodi

}

/*
Simulazione entrata nella sezione critica:
il processo rimane in sleep per un intervallo di tempo randomico compreso tra min e max
*/
func enterCS() {
	max := 15
	min := 10
	timeInCS := rand.Intn(max-min) + min

	log.Println("Entering critical section for ", timeInCS, " seconds")
	time.Sleep(time.Duration(timeInCS) * time.Second)
	log.Println("Exit from Critical Section")

}

/*
Istruzioni eseguite una volta usciti dalla sezione critica
*/
func exit(state *utils.State) {
	state.DeleteReq(state.GetHostname()) //eliminazione richiesta dalla coda
	log.Println("Deleted my request from queue ", state.GetQueue())
	if (state.GetNextRequest() == utils.Request{}) {
		log.Println("Empty queue.. ")
	}
	state.IncreaseClock()                                                                                       //clock += 1
	sendRelease(state.GetMembers(), ReleasePayload{SenderId: state.GetHostname(), Timestamp: state.GetClock()}) // Invio RELEASE  a tutti gli altri nodi
}

/*
Calcolo max tra req.timestamp e clock
*/
func maxClock(timestamp int, clock int) bool {
	return timestamp > clock
}
