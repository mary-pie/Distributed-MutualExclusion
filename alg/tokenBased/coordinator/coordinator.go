package coordinator

import (
	"log"
	"mutualexclusion-project/alg/tokenBased/coordinator/utils"
	"net"
	"net/http"
	"net/rpc"
	"time"

	"github.com/hashicorp/memberlist"
)

/*
Funzione di inizializzazione del coordinatore:
*/
func InitCoord(hostname string, port int, time_setup int) {

	ml := initGroup(hostname) // inizializzazione gruppo di mutua esclusione

	state := utils.NewState(hostname, port, ml) //init stato

	log.Println("Waiting for setup group..")
	time.Sleep(time.Duration(time_setup) * time.Second) // attesa setup gruppo
	log.Println("Group creation terminated. Members are: ", ml.NumMembers())

	state.SetUpV() //setup V
	log.Println("V: ", state.GetV())
	initServerRpc(state) //init server RPC
}

/*
Funzione di inizializzazione del cluster --> server in ascolto sulla 7946
*/
func initGroup(hostname string) *memberlist.Memberlist {

	config := memberlist.DefaultLANConfig()
	config.Name = hostname
	ml, err := memberlist.Create(config)
	if err != nil {
		panic("Failed to create memberlist: " + err.Error())
	}

	if err != nil {
		panic(err)
	}
	log.Println("Group successfully initialized!")
	return ml
}

/*
Funzione di inizializzazione del server RPC --> in ascolto sulla 9090
*/
func initServerRpc(s *utils.State) {

	me := new(MutualExclusion)

	server := rpc.NewServer()
	err := server.RegisterName("MutualExclusion", me)
	if err != nil {
		log.Fatal("Format of service Mutual Exclusion is not correct: ", err)
	}

	server.HandleHTTP("/", "/debug")

	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatal("Listen error: ", err)
	}
	me.state = s
	log.Printf("RPC server is up. Listening on port %d\n", 9090)

	err = http.Serve(lis, nil)
	if err != nil {
		log.Fatal("Serve error: ", err)
	}
}

/*
Funzione di invio del token a "host":
set isTokenFree a false
*/
func sendToken(host string, s *utils.State) {
	log.Println("Sending token ", token, "to ", host)
	var res string

	s.SetStateToken(false)
	addr := s.SearchMember(host).String() + ":9090"
	client, err := rpc.DialHTTP("tcp", addr)
	if err != nil {
		log.Fatal("Error in dialing: ", err)
	}
	defer client.Close()

	call := client.Go("MutualExclusion.ReceiveToken", token, &res, nil)
	if call.Error != nil {
		log.Fatal("Error in Access Critical Section: ", call.Error.Error())
	}
}
