package node

import (
	"log"
	"math/rand"
	"mutualexclusion-project/alg/tokenBased/node/utils"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
	"time"

	"github.com/DistributedClocks/GoVector/govec/vclock"
	"github.com/hashicorp/memberlist"
)

/*
Funzione di inizializzazione e esecuzione del processo:
1. join al gruppo di mutua esclusione
2. attesa del completamento del setup del gruppo
3. inizializzazione server RPC
4. setup clock vettoriale
5. setup stato
6. invio richiesta, entrata in CS, reinvio del token e invio msg di programma
*/
func InitProcess(hostname string, port int, time_setup int) {

	ml := joinCluster(hostname)
	state := utils.NewState(hostname, port, ml)

	go func() {
		log.Println("Waiting for setup group..")
		time.Sleep(time.Duration(time_setup) * time.Second)
		log.Println("Group creation terminated. Members are: ", ml.NumMembers())
		state.SetUpClock()
		log.Println("Vector Clock: ", state.GetClock())

		//ogni processo invia la propria richiesta dopo un certo intervallo di tempo randomico
		max := 15
		min := 5
		x := rand.Intn(max-min) + min
		time.Sleep(time.Duration(x) * time.Second)

		//trying protocol
		state.IncreaseClock()
		log.Println("Updating Clock before send request: ", state.GetClock())
		sendProgramMsg(PMsg{SenderClock: state.GetClock(), SenderId: state.GetHostname()}, state.GetMembers(), state.GetHostname())
		res := sendRequestRPC(state, RequestPayload{state.GetHostname(), state.GetClock()})
		if res != "" {
			log.Println("Token received: ", res)
			enterCS()                                                     //entrata nella sezione critica
			resendToken(TokenBack{SenderId: state.GetHostname(), T: res}) //reinvio del token
		}

	}()
	initServerRpc(state)
}

/*
Funzione di Join al gruppo di mutua esclusione inizializzato dal coordinatore:
inizializza un server memberlist in ascolto sulla 7946
*/
func joinCluster(hostaname string) *memberlist.Memberlist {
	config := memberlist.DefaultLANConfig()
	config.Name = hostaname

	ml, err := memberlist.Create(config)

	if err != nil {
		panic(err)
	}

	_, err = ml.Join([]string{"coord"})
	if err != nil {
		panic("Failed to join cluster: " + err.Error())
	}

	log.Println("Group Join is successfully terminated")
	return ml
}

/*
Inizializzione Server RPC per ricevere il Token e i Messaggi di Programma
*/
func initServerRpc(s *utils.State) {
	me := new(ListenerRPC)

	server := rpc.NewServer()
	err := server.RegisterName("MutualExclusion", me)
	if err != nil {
		log.Fatal("Format of service Mutual Exclusion is not correct: ", err)
	}

	server.HandleHTTP("/", "/debug")

	lis, err := net.Listen("tcp", ":"+strconv.Itoa(s.GetPort()))
	if err != nil {
		log.Fatal("Listen error: ", err)
	}

	me.state = s
	log.Printf("RPC server is up. Listening on port %d\n", s.GetPort())

	err = http.Serve(lis, nil)
	if err != nil {
		log.Fatal("Serve error: ", err)
	}
}

/*
Funzione per effettuare una chiamata RPC di richiesta per accedere alla sezione critica:
se riceve il token allora entra nella SC
*/
func sendRequestRPC(s *utils.State, req RequestPayload) Token {
	log.Println("Sending request ", req)
	var res Token
	addr := "coord:9090"

	client, err := rpc.DialHTTP("tcp", addr)
	if err != nil {
		log.Fatal("Error in dialing: ", err)
	}
	defer client.Close()
	call := client.Go("MutualExclusion.AccessCS", req, &res, nil)
	call = <-call.Done

	if call.Error != nil {
		log.Fatal("Error in Access Critical Section: ", call.Error.Error())
	}
	return res
}

/*
Funzione di reinvio del token al coordinatore
*/
func resendToken(token_back TokenBack) {
	log.Println("Resending token back to coordinator ", token_back)
	var res string
	addr := "coord:9090"

	client, err := rpc.DialHTTP("tcp", addr)
	if err != nil {
		log.Fatal("Error in dialing: ", err)
	}
	defer client.Close()

	call := client.Go("MutualExclusion.ResendToken", token_back, &res, nil)
	if call.Error != nil {
		log.Fatal("Error in Access Critical Section: ", call.Error.Error())
	}
}

/*
Funzione che effettua una chiamata RPC sugli altri nodi per inviare in messaggio di programma
*/
func sendProgramMsg(msg PMsg, ml []*memberlist.Node, hostname string) {
	log.Println("Sending program message ", msg)
	var response string
	for _, m := range ml {
		if m.Name != "coordinator" && m.Name != hostname {
			addr := m.Addr.String() + ":9090"
			client, err := rpc.DialHTTP("tcp", addr)
			if err != nil {
				log.Fatal("Error in dialing: ", err)
			}
			defer client.Close()

			call := client.Go("MutualExclusion.ProgramMessage", msg, &response, nil)
			if call.Error != nil {
				log.Fatal("Error in Access Critical Section: ", call.Error.Error())
			}
		}
	}
}

/*
Funzione che aggiorna il clock del processo in funzione del msg di programma ricevuto
*/
func processMsgProgram(sender_clock vclock.VClock, state *utils.State) {
	for i, clock := range sender_clock {
		if clock > state.GetClock()[i] {
			state.UpdateClock(i, clock)
		}
	}
	log.Println("New clock after received program msg: ", state.GetClock())
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
	log.Println("Exit from Critical Section. Resend token to coordinator")
}
