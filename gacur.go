package main

import (
	"encoding/json"
	"math/rand/v2"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/ucukertz/hfs"
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
	if roll > 10 {
		return
	}

	hint := `Respond to user messages as memey and brief as possible. 
	Don't reiterate user's message. Put it in a json. 
	prompt [natural language prompt to generate the meme picture], top [top text]. bot [bottom text]. 
	Answer purely with the json, no code blocks.`

	aians, err := ChatGaiOneText(WaMsgStr(msg), hint)
	if err != nil {
		if strings.Contains(err.Error(), "overload") {
			WaReact(msg, "🤕")
		} else {
			WaSaad(msg, err)
		}
		return
	}
	log.Debug().Str("aians", aians).Msg("TOKKE")

	var ais struct {
		P string `json:"prompt"`
		T string `json:"top"`
		B string `json:"bot"`
	}

	if err := json.Unmarshal([]byte(aians), &ais); err != nil {
		WaReact(msg, "😭")
	}

	query := "Meme of " + ais.P + ". Top text '" + ais.T + "'. Bottom text '" + ais.B + "'."
	img, err := hfs.NewHfs[any, any]("mrfakename-z-image-turbo").
		WithBearerToken(ENV_TOKEN_HUGGING).
		WithTimeout(HFS_TIMEOUT).
		DoFD("/generate_image", query, 1024, 1024, 9, 42, true)
	if err != nil {
		WaReact(msg, "😢")
	}

	WaReplyImg(msg, img, "")
}

var CurHeheEnd time.Time
var CurHeheScope *events.Message

func CurHeheAct(msg *events.Message) {
	if msg == nil {
		return
	}

	WaReplyText(msg, "h3h3 or heehee 👿")
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
	WaReplyVid(msg, vid, "Kawaii 🤮", true)
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
			WaReplyImg(msg, img, "You know what time is it?")
		case 2:
			vid, err := os.ReadFile("assets/disco.mp4")
			if err != nil {
				WaSaad(msg, err)
				return
			}
			WaReplyVid(msg, vid, "Zeta ter00s 🥵", true)
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
