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
	"minami rio": *new(SdChara).Create(
		SdTunes["rio"]["base"],
		"[:grey skirt, pleated skirt,:0.4] school uniform, white shirt, serafuku, yellow neckerchief, ",
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
		"[:purple_skirt, pleated_skirt, plaid_skirt, :0.3] white_shirt, short_sleeves, plaid_bowtie, purple bowtie, ",
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
	qry := WaMsgQry(msg)
	pow := int(lo.RandomString(1, lo.AllCharset)[0])
	cube := SdCube(qry, pow)
	WaText(msg, fmt.Sprintf("[%d] %s", pow, cube))
}

var SdBaker = 0

func SdBake(msg *events.Message) {
	qry := WaMsgQry(msg)
	baking, err := strconv.Atoi(qry)
	if err != nil {
		WaReact(msg, "❌")
		return
	}
	SdBaker = baking
	WaReact(msg, "👍")
}

func SdTake(msg *events.Message) {
	qry := WaMsgQry(msg)
	box := SdBox(qry, SdBaker)
	WaText(msg, box)
}
