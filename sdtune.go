package main

import (
	"encoding/base64"
	"fmt"
	"strconv"

	"github.com/samber/lo"
	"go.mau.fi/whatsmeow/types/events"
)

type SdChara struct {
	traits  string
	clothes string
	postpos string
	postneg string
}

func (chara *SdChara) Create(traits string, clothes string, postpos string, postneg string) *SdChara {
	chara.traits = traits
	chara.clothes = clothes
	chara.postpos = postpos
	chara.postneg = postneg
	return chara
}

var SdTunes = map[string]map[string]string{
	"elsie": {
		"base": SdBox("GSYhJR3YKxsqHR0mGxko5NjgHSQxKyEZ2Bwd2CQtLB3YISUZ8unm6eHk2BokGRsj2CAZISrk", 72),
	},
	"rio": {
		"base": SdBox("MhFQPENDRk73PFA8SgP3EQcFCjT3REBFOERA90lARgP3Tj9ASzz3PzhASQP3Q0ZFPvc/OEBJA/c/QEQ89zpMSwP3SkA7PENGOkJKAw==", 41),
	},
	"lucy": {
		"base":   SdEdge("/AXzCfP4Av/+//EC+9n8/AUDBAL5/wUD6Nw=", 1.0, 112) + SdBox("OUNGRTs8Nj84QEkD9zIROEhMODY8UDxKA/cRBwUKNA==", 41),
		"loop":   SdBox("EhsJHwkH0sYIEhsLBQsfCxnSxh0ODxoLBRkJBxgM0sYB4BIHIB8FCx8L0sbg1tTZAw==", 90),
		"cage":   SdBox("NDc2LycwKTE69OgwKTE6Jzc6Nik1LTY89OgwKTE6KzQxOPToNS0sMT01Jyo6LSk7PDv0", 56),
		"summer": SdBox("/AXzCQMEvLDryvzxCgnv9Qn1vLDKwL7D7bAH+PkE9e/y/we8sAD5/vvv8/j/+/UCvLAE+f4E9fTv9Qn1B/XxArywB/j5BPW99gLx/fX07/UJ9Qf18QK8", 112),
		"roll":   SdBox("GPcpHjc2HCI2Iund9+3r8BrdKTIgNiUv6Q==", 67),
	},
}

var SdChars = map[string]SdChara{
	"elsie": *new(SdChara).Create(
		SdTunes["elsie"]["base"],
		"dark_purple_japanese_clothes, sleeveless, [:purple_short_yukata, fingerless_gloves, white_elbow_gloves, white_thighhighs, zettai_ryouiki, thighs, :0.1] purple_choker, ",
		"",
		"alternate_skin_color, sleeves, side_slit, school_emblem, ",
	),
	"minami rio": *new(SdChara).Create(
		SdTunes["rio"]["base"],
		"school uniform, white shirt, serafuku, [:grey skirt, pleated skirt,:0.4] yellow neckerchief, ",
		"",
		"",
	),
	"lucy (loop)": *new(SdChara).Create(
		SdTunes["lucy"]["base"]+" "+SdTunes["lucy"]["loop"],
		"",
		"",
		"green_eyes, wavy_hair, ",
	),
	"lucy (cage)": *new(SdChara).Create(
		SdTunes["lucy"]["base"]+" "+SdTunes["lucy"]["cage"],
		"white_shirt, short_sleeves, plaid_bowtie, [:purple_skirt, pleated_skirt, plaid_skirt, :0.3] purple bowtie, ",
		"",
		"lazy_eye, green_eyes, purple_eyes, large_breasts, breast_pocket, ",
	),
	"lucy (summer)": *new(SdChara).Create(
		SdTunes["lucy"]["base"]+" "+SdTunes["lucy"]["summer"],
		"",
		"",
		"green_eyes, jewelry, gem, ",
	),
	"lucy (roll)": *new(SdChara).Create(
		SdTunes["lucy"]["base"]+" "+SdTunes["lucy"]["roll"],
		"",
		"",
		"green_eyes, ",
	),
}

var SdTurbo = SdEdge("MjMwIC0=", 0.7, 66)
var SdFws = SdBox("+e8C7eq28AO7+QP8/Oq3uq4B9vP8/O/37fv3Afbvuq77/QD3+QMA7+3z/LquAfb3/AftAfn3/Lqu/u/68+0B+ff8uq6Y0ODTz9k=", 114)
var SdFwa = SdBox("MjQ3LjA6NyYkKjPx5TgtLjM+JDgwLjPx5c8HFwoGEA==", 59)
var SdSugar = SdBox("wwoM/vgJ9hAGDMM=", 105)

func SdCube(base string, pow int) string {
	enc := []byte(base)
	for i := range enc {
		pos := (int(enc[i]) - int(pow)) % 256
		if pos < 0 {
			pos += 256
		}
		enc[i] = byte(pos)
	}
	return base64.StdEncoding.EncodeToString(enc)
}

func SdBox(base string, pow int) string {
	dec, _ := base64.StdEncoding.DecodeString(base)
	for i := range dec {
		pos := (int(dec[i]) + int(pow)) % 256
		if pos < 0 {
			pos += 256
		}
		dec[i] = byte(pos)
	}
	return string(dec)
}

func SdEdge(edge string, amp float32, pow int) string {
	sd := SdBox(edge, pow)
	return fmt.Sprintf("<%s:%s:%f>, ", SdBox("ISQnFg==", 75), sd, amp)
}

func SdTune(msg *events.Message) {
	prompt := WaMsgPrompt(msg)
	pow := int(lo.RandomString(1, lo.AllCharset)[0])
	cube := SdCube(prompt, pow)
	WaReplyText(msg, fmt.Sprintf("[%d] %s", pow, cube))
}

var SdBaker = 0

func SdBake(msg *events.Message) {
	prompt := WaMsgPrompt(msg)
	baking, err := strconv.Atoi(prompt)
	if err != nil {
		WaReact(msg, "❌")
		return
	}
	SdBaker = baking
	WaReact(msg, "👍")
}

func SdTake(msg *events.Message) {
	prompt := WaMsgPrompt(msg)
	box := SdBox(prompt, SdBaker)
	WaReplyText(msg, box)
}
