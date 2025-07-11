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
	lock       chan struct{}
	reply      chan *events.Message
	textBuffer *string
}

func OutsDestMake() OutsDest {
	return OutsDest{lock: make(chan struct{}, 1), reply: make(chan *events.Message), textBuffer: new(string)}
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
	"15854968266": OutsDestMake(), // Youbot
	"18002428478": OutsDestMake(), // ChatGPT
}

func OutsFilterReply(msg *events.Message, reply *events.Message) (string, bool) {
	if strings.Contains(reply.Info.Sender.User, "15854968266") {
		if strings.Contains(WaMsgStr(reply), "great to meet you!") {
			WaReact(msg, "⏳")
			return "", false
		}
		if strings.Contains(WaMsgStr(reply), "The answer does not rely on search results.") {
			ans := strings.Split(WaMsgStr(reply), "The answer does not rely on search results.")[0]
			ans = strings.TrimRight(ans, " \n_")
			return ans, true
		}
		return WaMsgStr(reply), true
	} else if strings.Contains(reply.Info.Sender.User, "18002428478") {
		dest := OutsDests["18002428478"]
		*dest.textBuffer = WaMsgStr(reply)

		finished := false
		for !finished {
			select {
			case <-time.After(10 * time.Second):
				finished = true
			case reply = <-dest.reply:
				text := fmt.Sprintln(*dest.textBuffer, WaMsgStr(reply))
				*dest.textBuffer = text
			}
		}

		go func() {
			time.Sleep(10 * time.Second)
			*dest.textBuffer = ""
		}()
		return *dest.textBuffer, true
	}

	return WaMsgStr(reply), true
}

func OutsExec(msg *events.Message, phone string, req *waE2E.Message) {
	_, err := meow.SendMessage(context.Background(), types.NewJID(phone, types.DefaultUserServer), req)
	if err != nil {
		log.Error().Err(err).Msg("REQ TEXT")
	} else {
		log.Debug().Msg("REQ TEXT OK")
	}

	// Wait for reply
	dest := OutsDests[phone]
	waitDur := 1 * time.Minute
	final := ""
	finished := false

	for !finished {
		select {
		case reply := <-dest.reply:
			final, finished = OutsFilterReply(msg, reply)
			if finished {
				WaText(msg, final)
			}

		case <-time.After(waitDur):
			WaSaad(msg, WaFallen)
			return
		}
	}
}

func OutsText(msg *events.Message, phone string) {
	waitDur := 1 * time.Minute
	dest, ok := OutsDests[phone]
	if !ok {
		log.Error().Err(fmt.Errorf(phone)).Msg("INVALID OUTS")
	}
	if dest.WIP() {
		WaReact(msg, "🐢")
	}
	if !dest.Wait(waitDur) {
		WaSaad(msg, WaFallen)
		return
	}
	dest.Lock()
	defer dest.Unlock()

	query := WaMsgQry(msg)
	req := waE2E.Message{Conversation: &query}
	OutsExec(msg, phone, &req)
}

func OutsCheck(msg *events.Message) {
	dest, ok := OutsDests[msg.Info.Sender.User]
	if !ok {
		return
	}
	if dest.WIP() {
		dest.reply <- msg
	}
}

func OutsCmdChk(msg *events.Message, cmd string) bool {
	switch cmd {
	case "!yai":
		go OutsText(msg, "15854968266")
		return true
	case "!cai":
		go OutsText(msg, "18002428478")
		return true
	}

	return false
}
