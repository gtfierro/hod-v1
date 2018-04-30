package bw2bind

import (
	"bufio"
	"net"
	"sync"
)

// BW2Client is a handle to your local BOSSWAVE router. It is obtained
// from Connect or ConnectOrExit
type BW2Client struct {
	c            net.Conn
	out          *bufio.Writer
	in           *bufio.Reader
	remotever    string
	seqnos       map[int]chan *frame
	olock        sync.Mutex
	curseqno     uint32
	defAutoChain *bool
	rHost        string
	wal          *wal
	errorHandler func(error)
}

func (cl *BW2Client) Close() error {
	return cl.c.Close()
}

func (cl *BW2Client) SetErrorHandler(f func(error)) {
	cl.errorHandler = f
}

//Sends a request frame and returns a  chan that contains all the responses.
//Automatically closes the returned channel when there are no more responses.
func (cl *BW2Client) transact(req *frame) chan *frame {
	seqno := req.SeqNo
	inchan := make(chan *frame, 3)
	outchan := make(chan *frame, 3)
	cl.olock.Lock()
	cl.seqnos[seqno] = inchan
	req.WriteToStream(cl.out)
	cl.olock.Unlock()
	go func() {
		for {
			fr, ok := <-inchan
			if !ok {
				close(outchan)
				return
			}
			outchan <- fr
			finished, ok := fr.GetFirstHeader("finished")
			if ok && finished == "true" {
				close(outchan)
				cl.closeSeqno(fr.SeqNo)
				return
			}
		}
	}()
	return outchan
}
func (cl *BW2Client) closeSeqno(seqno int) {
	cl.olock.Lock()
	ch, ok := cl.seqnos[seqno]
	if ok {
		close(ch)
		delete(cl.seqnos, seqno)
	}
	cl.olock.Unlock()
}
