package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

type OutsDest struct {
	send_phone string
	rcv_phone  string
	lock       chan struct{}
	reply      chan *events.Message
	textBuffer *string
}

func OutsDestMake(send_phone, rcv_phone string) OutsDest {
	return OutsDest{
		send_phone: send_phone,
		rcv_phone:  rcv_phone,
		lock:       make(chan struct{}, 1),
		reply:      make(chan *events.Message),
		textBuffer: new(string),
	}
}

func (o OutsDest) Lock() {
	o.lock <- struct{}{}
}

func (o OutsDest) Unlock() {
	<-o.lock
}

func (o OutsDest) WIP() bool {
	select {
	case o.lock <- struct{}{}:
		<-o.lock
		return false
	default:
		return true
	}
}

func (o OutsDest) Wait(t time.Duration) bool {
	select {
	case o.lock <- struct{}{}:
		<-o.lock
		return true
	case <-time.After(t):
		return false
	}
}

var OutsDests = map[string]OutsDest{
	"!cai": OutsDestMake("18002428478", "30722505109681"),
}

func OutsFilterReply(msg *events.Message, dest OutsDest, reply *events.Message) (string, bool) {
	if strings.Contains(reply.Info.Sender.User, dest.rcv_phone) {
		*dest.textBuffer = WaMsgStr(reply)

		finished := false
		for !finished {
			select {
			case <-time.After(3 * time.Second):
				finished = true
			case reply = <-dest.reply:
				text := fmt.Sprintln(*dest.textBuffer, WaMsgStr(reply))
				*dest.textBuffer = text
			}
		}

		go func() {
			time.Sleep(3 * time.Second)
			*dest.textBuffer = ""
		}()
		return *dest.textBuffer, true
	}

	return WaMsgStr(reply), true
}

func OutsExec(msg *events.Message, dest OutsDest, req *waE2E.Message) {
	_, err := meow.SendMessage(context.Background(), types.NewJID(dest.send_phone, types.DefaultUserServer), req)
	if err != nil {
		log.Error().Err(err).Msg("REQ TEXT")
	} else {
		log.Debug().Msg("REQ TEXT OK")
	}

	// Wait for reply
	waitDur := 1 * time.Minute
	final := ""
	finished := false

	for !finished {
		select {
		case reply := <-dest.reply:
			final, finished = OutsFilterReply(msg, dest, reply)
			if finished {
				WaReplyText(msg, final)
			}

		case <-time.After(waitDur):
			WaSaad(msg, WaSaadFallen)
			return
		}
	}
}

func OutsText(msg *events.Message, dest OutsDest) {
	waitDur := 1 * time.Minute
	if dest.WIP() {
		WaReact(msg, "🐢")
	}
	if !dest.Wait(waitDur) {
		WaSaad(msg, WaSaadFallen)
		return
	}
	dest.Lock()
	defer dest.Unlock()

	query := WaMsgPrompt(msg)
	req := waE2E.Message{Conversation: &query}
	OutsExec(msg, dest, &req)
}

func OutsCapture(msg *events.Message) {
	validDest := OutsDest{}
	for _, dest := range OutsDests {
		if strings.Contains(msg.Info.Sender.User, dest.rcv_phone) {
			validDest = dest
			break
		}
	}
	if (validDest == OutsDest{}) {
		return
	}
	if validDest.WIP() {
		validDest.reply <- msg
	}
}

func OutsCmdChk(msg *events.Message, cmd string) bool {
	dest, ok := OutsDests[cmd]
	if !ok {
		return false
	}
	go OutsText(msg, dest)
	return true
}
