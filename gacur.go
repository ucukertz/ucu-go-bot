package main

import (
	"math/rand/v2"
	"os"
	"regexp"
	"strings"
	"time"

	"go.mau.fi/whatsmeow/types/events"
)

func GachaTokke(msg *events.Message) {
	roll := rand.IntN(100)
	if roll > 1 {
		return
	}

	hint := "Respond to user messages with two sentences at most. Be as memey as possible."
	aians, err := GaiSingleText(WaMsgStr(msg), hint)
	if err != nil {
		WaSaad(msg, err)
	}
	log.Debug().Str("aians", aians).Msg("TOKKE")
	if len(strings.Split(aians, ".")) > 2 {
		aians = strings.Split(aians, ".")[0]
	}

	r, err := HttpcBase.R().
		SetBody(map[string]any{"text": aians, "safe": true, "redirect": true}).
		Post(ENV_BASEURL_MEMEGEN)
	if err != nil {
		WaSaad(msg, err)
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
			sec := rand.IntN(120)
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
			sec := rand.IntN(120)
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
		if CurZetaCntr == 1 {
			img, err := os.ReadFile("assets/disco.jpg")
			if err != nil {
				WaSaad(msg, err)
				return
			}
			WaImage(msg, img, "You know what time is it?")
		} else if CurZetaCntr == 2 {
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
	GachaTokke(msg)
}

func Gacur(msg *events.Message) {
	GachaRoll(msg)
	CurHehe(msg)
	CurTehe(msg)
	CurZeta(msg)
}

func GacurCmdChk(msg *events.Message, cmd string) bool {
	if cmd == "!jail" {
		CurHeheAct(CurHeheScope)
		return true
	} else if cmd == "!jwail" {
		CurTeheAct(CurTeheScope)
		return true
	}

	return false
}
