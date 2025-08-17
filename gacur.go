package main

import (
	"math/rand/v2"
	"os"
	"regexp"
	"strings"
	"time"

	"go.mau.fi/whatsmeow/types/events"
)

func GachaRand64(min int64, max int64) int64 {
	return rand.Int64N(max-min+1) + min
}

func GachaTokke(msg *events.Message) {
	if len(WaMsgStr(msg)) == 0 {
		return
	}
	if strings.HasPrefix(WaMsgStr(msg), "!") {
		return
	}

	roll := GachaRand64(0, 1000)
	if roll > 25 {
		return
	}

	hint := "Respond to user messages with two sentences at most. Be as memey as possible."
	aians, err := ChatGaiOneText(WaMsgStr(msg), hint)
	if err != nil {
		WaSaad(msg, err)
	}
	log.Debug().Str("aians", aians).Msg("TOKKE")
	if len(strings.Split(aians, ".")) > 2 {
		aians = strings.Split(aians, ".")[0]
	}

	r, err := HttpcBase().R().
		SetBody(map[string]any{"text": aians, "safe": true, "redirect": true}).
		Post(ENV_BASEURL_MEMEGEN)
	if err != nil {
		WaText(msg, aians)
		return
	}

	WaImage(msg, r.Body(), "")
}

var CurHeheEnd time.Time
var CurHeheScope *events.Message

func CurHeheAct(msg *events.Message) {
	if msg == nil {
		return
	}

	WaText(msg, "h3h3 or heehee 👿")
	CurHeheEnd = time.Now().Add(7 * 24 * time.Hour)
}

func CurHehe(msg *events.Message) {
	inmate := "7102"
	chk := func(str string) bool {
		regex := regexp.MustCompile(`(?i)h\s*e\s*h\s*e`)
		return regex.MatchString(str)
	}

	if ok := strings.Contains(msg.Info.Sender.String(), inmate); !ok {
		return
	}
	CurHeheScope = msg

	if CurHeheEnd.After(time.Now()) {
		go func(m *events.Message) {
			sec := GachaRand64(20, 120)
			time.Sleep(time.Duration(sec) * time.Second)
			WaReact(msg, "😡")
		}(msg)
	}

	if ok := chk(WaMsgStr(msg)); !ok {
		return
	}
	CurHeheAct(msg)
}

var CurTeheEnd time.Time
var CurTeheScope *events.Message

func CurTeheAct(msg *events.Message) {
	if msg == nil {
		return
	}

	vid, err := os.ReadFile("assets/disco.mp4")
	if err != nil {
		WaSaad(msg, err)
		return
	}
	WaVideo(msg, vid, "Kawaii 🤮", true)
	CurTeheEnd = time.Now().Add(7 * 24 * time.Hour)
}

func CurTehe(msg *events.Message) {
	inmate := "7095"
	chk := func(str string) bool {
		regex := regexp.MustCompile(`(?i)t\s*e+[e\s]*\s*h\s*e`)
		return regex.MatchString(str)
	}

	if ok := strings.Contains(msg.Info.Sender.String(), inmate); !ok {
		return
	}
	CurTeheScope = msg

	if CurTeheEnd.After(time.Now()) {
		go func(m *events.Message) {
			sec := GachaRand64(20, 120)
			time.Sleep(time.Duration(sec) * time.Second)
			WaReact(msg, "😛")
		}(msg)
	}

	if ok := chk(WaMsgStr(msg)); !ok {
		return
	}
	CurTeheAct(msg)
}

var CurZetaDay = time.Now().Weekday()
var CurZetaCntr = 0

func CurZeta(msg *events.Message) {
	inmate := "7095"
	if time.Now().Weekday() != CurZetaDay {
		CurZetaCntr = 0
		CurZetaDay = time.Now().Weekday()
	}
	if CurZetaCntr == 2 {
		return
	}

	if ok := strings.Contains(msg.Info.Sender.String(), inmate); !ok {
		return
	}
	if strings.Contains(strings.ToLower(WaMsgStr(msg)), "zeta") {
		CurZetaCntr++
		switch CurZetaCntr {
		case 1:
			img, err := os.ReadFile("assets/disco.jpg")
			if err != nil {
				WaSaad(msg, err)
				return
			}
			WaImage(msg, img, "You know what time is it?")
		case 2:
			vid, err := os.ReadFile("assets/disco.mp4")
			if err != nil {
				WaSaad(msg, err)
				return
			}
			WaVideo(msg, vid, "Zeta ter00s 🥵", true)
		}
		return
	}
}

func GachaRoll(msg *events.Message) {
	go GachaTokke(msg)
}

func Gacur(msg *events.Message) {
	GachaRoll(msg)
	CurHehe(msg)
	CurTehe(msg)
	CurZeta(msg)
}

func GacurCmdChk(msg *events.Message, cmd string) bool {
	switch cmd {
	case "!jail":
		CurHeheAct(CurHeheScope)
		return true
	case "!jwail":
		CurTeheAct(CurTeheScope)
		return true
	}

	return false
}
