package main

import (
	"fmt"
	"strings"

	"go.mau.fi/whatsmeow/types/events"
)

func MenuTop() string {
	return strings.Join([]string{
		"*Ucukertz WA bot*",
		"*!ai* Gemini",
		"*!cai* ChatGPT",
		"*!img* SDXL",
		"*!imgm* Advanced image gen menu",
		"*!what* More explanation for commands (ex: '!what ai')",
	}, "\n")
}

func MenuImage() string {
	return strings.Join([]string{
		"*Advanced image gen menu*",
		"*A1111* - [Modal]",
		"*!m.fwa* Fuwa v7 Advanced",
		"*!m.fws* Fuwa v7 Super",
		"*!m.fwt* Fuwa v7 Turbo",
		"*!m.mei* Meina v6",
		"*!m.wai* Waifu v5",
		"",
		"_A1111 utilities_",
		"*!reso* Set custom resolution",
		"*!bluff* Bluff",
		"*!seed* Set or lock seed",
		"",
		"*Natural Language Editing*",
		"*!m.flx* Flux Kontext",
		"",
		"*Legacy Stable Diffusion*",
		"*!sxl* Stable Diffusion XL",
	}, "\n")
}

func MenuWhat(query string) string {
	switch query {
	case "ai":
		return "Gemini, ask anything. Can process pictures and documents."
	case "cai":
		return "ChatGPT, ask anything. Capable of browsing the web. Slightly slower to response."
	case "yai":
		return "YouBot, ask anything. Frequently out of service."
	case "img":
		return "Stable Diffusion XL txt2img. Massive breakthrough compared to earlier versions of SD."
	case "l.sxl":
		return "Stable Diffusion XL txt2img. Massive breakthrough compared to earlier versions of SD."
	case "m.wai":
		return "[A1111] Waifu v5 txt2img. Anime-style. Booru average style for chars. 2x slower than Fuwa."
	case "m.mei":
		return "[A1111] Meina v6 txt2img. Cutesy anime-style. 2x slower than Fuwa."
	case "m.fwt":
		return "[A1111] Fuwa v7 txt2img. Stable cutesy anime-style, vibrant colors."
	case "m.fws":
		return "[A1111] Fuwa v7 txt2img. Super stable cutesy anime-style, vibrant colors."
	case "m.fwa":
		return "[A1111] Fuwa v7 txt2img. Stable clean anime-style, better anatomy."
	case "reso":
		return "[A1111] Set custom resolution for image generation."
	case "bluff":
		return "[A1111] Next generation outputs bluff image."
	case "seed":
		return "[A1111] Send only `!i.seed` to toggle seed randomness. Send `!i.seed <number>` to set a specific seed."
	case "m.flx":
		return "Flux Kontext txtimg2img. Natural Language Editing."
	case "what":
		return "*U wot m8?*"
	default:
		return fmt.Sprint(query, " (2)", "\nSend `!menu` for command list")
	}
}

func MenuCmdChk(msg *events.Message, cmd string) bool {
	switch cmd {
	case "!menu":
		WaReplyText(msg, MenuTop())
		return true
	case "!imgm":
		WaReplyText(msg, MenuImage())
		return true
	case "!what":
		WaReplyText(msg, MenuWhat(WaMsgPrompt(msg)))
		return true
	case "!ping":
		WaReact(msg, "🆙")
		return true
	}
	return false
}
