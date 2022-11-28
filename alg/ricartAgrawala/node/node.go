package node

import (
	"log"
	"math/rand"
	"mutualexclusion-project/alg/ricartAgrawala/node/utils"
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
func InitProc(hostname string, port int, time_setup int) {

	ml := joinGroup(hostname)                   //join al gruppo di mutua esclusione
	state := utils.NewState(hostname, port, ml) //init stato

	go func() {
		log.Println("Waiting for setup group..")
		time.Sleep(time.Duration(time_setup) * time.Second) // attesa fine setup del gruppo
		log.Println("Group creation terminated. Members are: ", ml.NumMembers())

		//ogni processo invia la propria richiesta dopo un certo intervallo di tempo randomico

		max := 15
		min := 5
		x := rand.Intn(max-min) + min
		time.Sleep(time.Duration(x) * time.Second)

		trying(state) //trying protocol
		if state.GetReplies() == len(state.GetMembers())-1 {
			state.SetStatus(utils.CS) //set status a CS
			enterCS()                 //entrata CS
			exit(state)               //exit protocol
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
se riceve REPLY incrementa il # di reply ricevuti
*/
func sendRequest(ml []*memberlist.Node, req RequestPayload, hostname string, state *utils.State) {
	log.Println("Sending request to other nodes: ", req)
	for _, m := range ml {
		if m.Name != hostname {
			addr := m.Addr.String() + ":9090"
			client, err := rpc.DialHTTP("tcp", addr)
			if err != nil {
				log.Fatal("Error in dialing: ", err)
			}
			reply := new(Reply)
			call := client.Go("MutualExclusion.AccessCS", req, reply, nil)
			call = <-call.Done

			if reply.SenderId != "" {
				state.UpdateClock(reply.Timestamp)
				log.Println("REPLY received ", *reply)
				state.IncreaseReplies()
			}
			if call.Error != nil {
				log.Fatal("Error in Access Critical Section: ", call.Error.Error())
			}
			client.Close()
		}
	}
}

/*
Funzione per effettuare una chiamata RPC di send REPLY che invia la risposta alle richieste in coda
*/
func sendReply(queue []utils.Request, ml []*memberlist.Node, rep Reply) {
	var res string
	log.Println("Sending REPLY ", rep, "to nodes in: ", queue)

	//init mappa di id dei processi in coda
	waiting_reply := map[string]int{}
	for _, req := range queue {
		waiting_reply[req.Sender] = 0
	}

	for _, m := range ml {
		if _, exists := waiting_reply[m.Name]; exists {
			addr := m.Addr.String() + ":9090"
			client, err := rpc.DialHTTP("tcp", addr)
			if err != nil {
				log.Fatal("Error in dialing: ", err)
			}
			call := client.Go("MutualExclusion.ReceiveReply", rep, &res, nil)

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
	state.SetStatus(utils.Requesting) //Status --> Requesting
	state.IncreaseClock()             // incremento clock prima dell'invio di una richiesta

	/*Test richieste concorrenti
	state.SetLastReq(1)
	sendRequest(state.GetMembers(), RequestPayload{SenderId: state.GetHostname(), Timestamp: 1}, state.GetHostname(), state)
	*/
	state.SetLastReq(state.GetClock())                                                                                                      // LastReq = Clock
	sendRequest(state.GetMembers(), RequestPayload{SenderId: state.GetHostname(), Timestamp: state.GetClock()}, state.GetHostname(), state) // Invio richiesta agli altri nodi

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
Istruzioni eseguite all'uscita dalla sezione critica:
Invio reply ai nodi nella coda delle richieste pendenti
Reset stato
*/
func exit(state *utils.State) {

	if len(state.GetQueue()) == 0 {
		log.Println("Empy queue ", state.GetQueue())
	} else {
		state.IncreaseClock() // incremento clock prima dell'invio di REPLY
		sendReply(state.GetQueue(), state.GetMembers(), Reply{SenderId: state.GetHostname(), Timestamp: state.GetClock()})
	}

	state.ResetState()
}

/*
Funzione che verifica se {LastReq, id} > {req.Timestamp, req.SenderId}
*/
func compareRequest(req RequestPayload, last_timestamp int, hostname string) bool {
	if req.Timestamp == last_timestamp {
		return req.SenderId > hostname
	}
	return req.Timestamp > last_timestamp
}
