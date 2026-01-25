package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"go.mau.fi/whatsmeow/types/events"
)

type GenReso struct {
	Name   string
	Width  int
	Height int
}

var GenResos = map[string]GenReso{
	// SDXL standard resolutions
	"sq": {Name: "SDXL_SQ", Width: 1024, Height: 1024},
	"w1": {Name: "SDXL_W1", Width: 1152, Height: 896},
	"h1": {Name: "SDXL_H1", Width: 896, Height: 1152},
	"w2": {Name: "SDXL_W2", Width: 1216, Height: 832},
	"h2": {Name: "SDXL_H2", Width: 832, Height: 1216},
	"w3": {Name: "SDXL_W3", Width: 1344, Height: 768},
	"h3": {Name: "SDXL_H3", Width: 768, Height: 1344},

	// 1k
	"10sq": {Name: "1024_1:1", Width: 1024, Height: 1024},
	"10w1": {Name: "1024_9:7", Width: 1152, Height: 896},
	"10h1": {Name: "1024_7:9", Width: 896, Height: 1152},
	"10w2": {Name: "1024_4:3", Width: 1152, Height: 864},
	"10h2": {Name: "1024_3:4", Width: 864, Height: 1152},
	"10w3": {Name: "1024_3:2", Width: 1248, Height: 832},
	"10h3": {Name: "1024_2:3", Width: 832, Height: 1248},
	"10w4": {Name: "1024_16:9", Width: 1280, Height: 720},
	"10h4": {Name: "1024_9:16", Width: 720, Height: 1280},
	"10w5": {Name: "1024_21:9", Width: 1344, Height: 576},
	"10h5": {Name: "1024_9:21", Width: 576, Height: 1344},

	// 1.2k
	"12sq": {Name: "1280_1:1", Width: 1280, Height: 1280},
	"12w1": {Name: "1280_9:7", Width: 1440, Height: 1120},
	"12h1": {Name: "1280_7:9", Width: 1120, Height: 1440},
	"12w2": {Name: "1280_4:3", Width: 1472, Height: 1104},
	"12h2": {Name: "1280_3:4", Width: 1104, Height: 1472},
	"12w3": {Name: "1280_3:2", Width: 1536, Height: 1024},
	"12h3": {Name: "1280_2:3", Width: 1024, Height: 1536},
	"12w4": {Name: "1280_16:9", Width: 1536, Height: 864},
	"12h4": {Name: "1280_9:16", Width: 864, Height: 1536},
	"12w5": {Name: "1280_21:9", Width: 1680, Height: 720},
	"12h5": {Name: "1280_9:21", Width: 720, Height: 1680},

	// 1.5k
	"15sq": {Name: "1536_1:1", Width: 1536, Height: 1536},
	"15w1": {Name: "1536_9:7", Width: 1728, Height: 1344},
	"15h1": {Name: "1536_7:9", Width: 1344, Height: 1728},
	"15w2": {Name: "1536_4:3", Width: 1728, Height: 1296},
	"15h2": {Name: "1536_3:4", Width: 1296, Height: 1728},
	"15w3": {Name: "1536_3:2", Width: 1872, Height: 1248},
	"15h3": {Name: "1536_2:3", Width: 1248, Height: 1872},
	"15w4": {Name: "1536_16:9", Width: 2048, Height: 1152},
	"15h4": {Name: "1536_9:16", Width: 1152, Height: 2048},
	"15w5": {Name: "1536_21:9", Width: 2016, Height: 864},
	"15h5": {Name: "1536_9:21", Width: 864, Height: 2016},
}

type GenDen struct {
	name     string
	strength float64
}

var GenDens = map[string]GenDen{
	"1": {name: "near-exact", strength: 0.1},
	"2": {name: "regenerative", strength: 0.2},
	"3": {name: "creative", strength: 0.3},
	"4": {name: "transformative", strength: 0.4},
	"5": {name: "scramble", strength: 0.5},
	"6": {name: "wild", strength: 0.6},
	"7": {name: "chaotic", strength: 0.7},
	"8": {name: "abstract", strength: 0.8},
	"9": {name: "unrecognizable", strength: 0.9},
}

const MAX_USERPIC = 5

type GenUserConfig struct {
	Bluff   bool
	Reso    GenReso
	Seed    int64
	Denoise GenDen
	Pics    [MAX_USERPIC][]byte
}

var GenActiveUserConfig = map[string]GenUserConfig{}
var GenDefaultUserConfig = GenUserConfig{Bluff: false, Reso: GenResos["sq"], Seed: -1, Denoise: GenDens["3"]}

func GenSet(msg *events.Message, uconfig GenUserConfig) {
	user := WaMsgUser(msg)
	if _, ok := GenActiveUserConfig[user]; !ok {
		WaReplyText(msg, "Hi, user "+msg.Info.Sender.User+"!")
	}
	GenActiveUserConfig[user] = uconfig
}

func GenSetBluff(msg *events.Message, bluff bool) {
	user := WaMsgUser(msg)
	ucfg, ok := GenActiveUserConfig[user]
	if !ok {
		GenSet(msg, GenDefaultUserConfig)
		ucfg = GenDefaultUserConfig
	}
	ucfg.Bluff = bluff
	GenActiveUserConfig[user] = ucfg
}

func GenSetReso(msg *events.Message, reso string) {
	user := WaMsgUser(msg)
	ucfg, ok := GenActiveUserConfig[user]
	if !ok {
		GenSet(msg, GenDefaultUserConfig)
		ucfg = GenDefaultUserConfig
	}
	ucfg.Reso = GenResos[reso]
	GenActiveUserConfig[user] = ucfg
}

func GenSetDenoise(msg *events.Message, denoise string) {
	user := WaMsgUser(msg)
	ucfg, ok := GenActiveUserConfig[user]
	if !ok {
		GenSet(msg, GenDefaultUserConfig)
		ucfg = GenDefaultUserConfig
	}
	ucfg.Denoise = GenDens[denoise]
	GenActiveUserConfig[user] = ucfg
}

func GenSetSeed(msg *events.Message, seed int64) {
	user := WaMsgUser(msg)
	ucfg, ok := GenActiveUserConfig[user]
	if !ok {
		GenSet(msg, GenDefaultUserConfig)
		ucfg = GenDefaultUserConfig
	}
	ucfg.Seed = seed
	GenActiveUserConfig[user] = ucfg
}

func GenSetPic(msg *events.Message, num int, pic []byte) {
	user := WaMsgUser(msg)
	ucfg, ok := GenActiveUserConfig[user]
	if !ok {
		GenSet(msg, GenDefaultUserConfig)
		ucfg = GenDefaultUserConfig
	}
	ucfg.Pics[num-1] = pic
	GenActiveUserConfig[user] = ucfg
}

func GenSetPicClear(msg *events.Message) {
	user := WaMsgUser(msg)
	ucfg, ok := GenActiveUserConfig[user]
	if !ok {
		GenSet(msg, GenDefaultUserConfig)
		ucfg = GenDefaultUserConfig
	}
	for i := range MAX_USERPIC {
		ucfg.Pics[i] = nil
	}
	GenActiveUserConfig[user] = ucfg
}

func GenGet(msg *events.Message) GenUserConfig {
	user := WaMsgUser(msg)
	ucfg, ok := GenActiveUserConfig[user]
	if !ok {
		GenSet(msg, GenDefaultUserConfig)
		return GenDefaultUserConfig
	}
	return ucfg
}

func GenCmdChk(msg *events.Message, cmd string) bool {
	ucfg := GenGet(msg)
	switch cmd {
	case "!reso":
		if reso, ok := GenResos[WaMsgPrompt(msg)]; ok {
			GenSetReso(msg, WaMsgPrompt(msg))
			WaReplyText(msg, fmt.Sprintf("Resolution set to *%s* for you 🫶", reso.Name))
		} else {
			WaReplyText(msg, "Resolution not found. Choices: \nsend !resos\n\nExample: `!reso sq`")
		}
		return true
	case "!resos":
		choices := "Available resolutions:\n"
		// Collect keys to have a consistent order
		keys := []string{}
		for k := range GenResos {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		resos := []GenReso{}
		for _, key := range keys {
			resos = append(resos, GenResos[key])
		}
		for i, reso := range resos {
			choices = fmt.Sprint(choices, fmt.Sprintf("- *%s* -> %s (%dx%d)\n", keys[i], reso.Name, reso.Width, reso.Height))
		}
		WaReplyText(msg, choices)
		return true
	case "!den":
		if denoise, ok := GenDens[WaMsgPrompt(msg)]; ok {
			GenSetDenoise(msg, WaMsgPrompt(msg))
			WaReplyText(msg, fmt.Sprintf("Denoise set to *%.1f (%s)* for you 🫶", denoise.strength, denoise.name))
		} else {
			WaReplyText(msg, "Denoise strength not found. Choices: \nsend !dens\n\nExample: `!den 3`")
		}
		return true
	case "!dens":
		choices := "Available denoise strengths:\n"
		// Collect keys to have a consistent order
		keys := []string{}
		for k := range GenDens {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		denoises := []GenDen{}
		for _, key := range keys {
			denoises = append(denoises, GenDens[key])
		}
		for i, denoise := range denoises {
			choices = fmt.Sprint(choices, fmt.Sprintf("- *%s* -> %s (%.1f)\n", keys[i], denoise.name, denoise.strength))
		}
		WaReplyText(msg, choices)
		return true
	case "!bluff":
		GenSetBluff(msg, true)
		WaReact(msg, "😏")
		return true
	case "!seed":
		var seed int64 = ucfg.Seed
		qry := WaMsgPrompt(msg)
		if parsed, err := strconv.ParseInt(qry, 10, 64); err == nil {
			seed = parsed
			WaReact(msg, "🔒")
		} else if seed == -1 {
			seed = GachaRand64(1e9, 9e9)
			WaReact(msg, "🔒")
		} else if seed != -1 {
			seed = -1
			WaReact(msg, "🎲")
		}
		GenSetSeed(msg, seed)
		return true
	}

	if strings.HasPrefix(cmd, "!pic") {
		pic := WaMsgMedia(msg)
		if pic == nil && !strings.HasSuffix(cmd, "0") {
			pic = WaMsgMediaQuoted(msg)
			if pic == nil {
				WaReplyText(msg, "No image to set ☹️")
				return true
			}
		}
		if strings.HasSuffix(cmd, "0") {
			GenSetPicClear(msg)
			WaReplyText(msg, "Your pics are cleared ✨")
			return true
		} else {
			numStr := strings.TrimPrefix(cmd, "!pic")
			num, err := strconv.Atoi(numStr)

			if err != nil {
				text := fmt.Sprintf("Invalid format. Use !picN (N=0-%d)", MAX_USERPIC)
				WaReplyText(msg, text)
				return true
			}
			if num < 1 || num > MAX_USERPIC {
				text := fmt.Sprintf("Invalid pic number (only 1-%d)", MAX_USERPIC)
				WaReplyText(msg, text)
				return true
			}
			GenSetPic(msg, num, pic)
			WaReplyText(msg, fmt.Sprintf("Pic%d is set 🫶", num))
		}
		return true
	}
	return false
}
