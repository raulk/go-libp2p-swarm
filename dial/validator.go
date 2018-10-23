package dial

import (
	"errors"
)

import lgbl "github.com/libp2p/go-libp2p-loggables"

// ErrDialToSelf is returned if we attempt to dial our own peer
var ErrDialToSelf = errors.New("dial to self attempted")

type validator struct {
	bestConnFn BestConnFn
}

var _ RequestPreparer = (*validator)(nil)

func NewValidator(bestConnFn BestConnFn) RequestPreparer {
	return &validator{bestConnFn: bestConnFn}
}

func (v *validator) Prepare(req *Request) {
	id := req.id
	log.Debugf("[%s] swarm dialing peer [%s]", req.net.LocalPeer(), id)

	var logdial = lgbl.Dial("swarm", req.net.LocalPeer(), id, nil, nil)
	if err := id.Validate(); err != nil {
		req.Complete(nil, err)
		return
	}

	if id == req.net.LocalPeer() {
		log.Event(req.ctx, "swarmDialSelf", logdial)
		req.Complete(nil, ErrDialToSelf)
		return
	}

	defer log.EventBegin(req.ctx, "swarmDialAttemptSync", id).Done()

	// check if we already have an open connection first
	if conn := v.bestConnFn(id); conn != nil {
		req.Complete(conn, nil)
	}
}
