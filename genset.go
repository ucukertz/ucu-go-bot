package main

import (
	"fmt"
	"sort"
	"strconv"

	"go.mau.fi/whatsmeow/types/events"
)

type GenReso struct {
	name   string
	Width  int
	Height int
}

var GenResos = map[string]GenReso{
	// SDXL standard resolutions
	"sq": {name: "SDXL_SQ", Width: 1024, Height: 1024},
	"w1": {name: "SDXL_W1", Width: 1152, Height: 896},
	"h1": {name: "SDXL_H1", Width: 896, Height: 1152},
	"w2": {name: "SDXL_W2", Width: 1216, Height: 832},
	"h2": {name: "SDXL_H2", Width: 832, Height: 1216},
	"w3": {name: "SDXL_W3", Width: 1344, Height: 768},
	"h3": {name: "SDXL_H3", Width: 768, Height: 1344},

	// 1k
	"10sq": {name: "1024_1x1", Width: 1024, Height: 1024},
	"10w1": {name: "1024_9x7", Width: 1152, Height: 896},
	"10h1": {name: "1024_7x9", Width: 896, Height: 1152},
	"10w2": {name: "1024_4x3", Width: 1152, Height: 864},
	"10h2": {name: "1024_3x4", Width: 864, Height: 1152},
	"10w3": {name: "1024_3x2", Width: 1248, Height: 832},
	"10h3": {name: "1024_2x3", Width: 832, Height: 1248},
	"10w4": {name: "1024_16x9", Width: 1280, Height: 720},
	"10h4": {name: "1024_9x16", Width: 720, Height: 1280},
	"10w5": {name: "1024_21x9", Width: 1344, Height: 576},
	"10h5": {name: "1024_9x21", Width: 576, Height: 1344},

	// 1.2k
	"12sq": {name: "1280_1x1", Width: 1280, Height: 1280},
	"12w1": {name: "1280_9x7", Width: 1440, Height: 1120},
	"12h1": {name: "1280_7x9", Width: 1120, Height: 1440},
	"12w2": {name: "1280_4x3", Width: 1472, Height: 1104},
	"12h2": {name: "1280_3x4", Width: 1104, Height: 1472},
	"12w3": {name: "1280_3x2", Width: 1536, Height: 1024},
	"12h3": {name: "1280_2x3", Width: 1024, Height: 1536},
	"12w4": {name: "1280_16x9", Width: 1536, Height: 864},
	"12h4": {name: "1280_9x16", Width: 864, Height: 1536},
	"12w5": {name: "1280_21x9", Width: 1680, Height: 720},
	"12h5": {name: "1280_9x21", Width: 720, Height: 1680},

	// 1.5k
	"15sq": {name: "1536x1536", Width: 1536, Height: 1536},
	"15w1": {name: "1536_9x7", Width: 1728, Height: 1344},
	"15h1": {name: "1536_7x9", Width: 1344, Height: 1728},
	"15w2": {name: "1536_4x3", Width: 1728, Height: 1296},
	"15h2": {name: "1536_3x4", Width: 1296, Height: 1728},
	"15w3": {name: "1536_3x2", Width: 1872, Height: 1248},
	"15h3": {name: "1536_2x3", Width: 1248, Height: 1872},
	"15w4": {name: "1536_16x9", Width: 2048, Height: 1152},
	"15h4": {name: "1536_9x16", Width: 1152, Height: 2048},
	"15w5": {name: "1536_21x9", Width: 2016, Height: 864},
	"15h5": {name: "1536_9x21", Width: 864, Height: 2016},
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

type GenUserConfig struct {
	Bluff   bool
	Reso    GenReso
	Seed    int64
	Denoise GenDen
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
			WaReplyText(msg, fmt.Sprintf("Resolution set to *%s* for you 🫶", reso.name))
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
			choices = fmt.Sprint(choices, fmt.Sprintf("- *%s* -> %s (%dx%d)\n", keys[i], reso.name, reso.Width, reso.Height))
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
	return false
}
