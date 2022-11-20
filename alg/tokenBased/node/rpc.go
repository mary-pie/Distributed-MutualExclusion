package node

import (
	"log"
	"mutualexclusion-project/alg/tokenBased/node/utils"

	"github.com/DistributedClocks/GoVector/govec/vclock"
)

type RequestPayload struct {
	SenderId  string
	Timestamp vclock.VClock
}

type Token string

type PMsg struct {
	SenderId    string
	SenderClock vclock.VClock
}

type TokenBack struct {
	SenderId string
	T        Token
}

type ListenerRPC struct {
	state *utils.State
}

/*
Handler che gestisce ricezione del token inviato dal coordinatore:
entre nella sezione criticca
reinvia il token al coordinatore
*/
func (l *ListenerRPC) ReceiveToken(token Token, res *TokenBack) error {
	log.Println("Token received ", token)
	enterCS()
	resendToken(TokenBack{SenderId: l.state.GetHostname(), T: token})
	return nil
}

/*
Handler che gestisce la ricezione dei messaggi di programma da parte degli altri processi nel gruppo:
aggiorno eventualmente il clock
*/
func (l *ListenerRPC) ProgramMessage(pmsg PMsg, res *string) error {
	log.Println("Program Message received ", pmsg)
	processMsgProgram(pmsg.SenderClock, l.state)
	return nil
}
