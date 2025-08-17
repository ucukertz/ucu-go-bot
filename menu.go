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
		"*!yai* YouBot",
		"*!img* SDXL",
		"*!imgm* Advanced image gen menu",
		"*!what* More explanation for commands (ex: '!what ai')",
	}, "\n")
}

func MenuImage() string {
	return strings.Join([]string{
		"*Advanced image gen menu*",
		"*A1111* - [Modal]",
		"*!i.fws* Fuwa v7 Super",
		"*!i.fwt* Fuwa v7 Turbo",
		"*!i.fwa* Fuwa v7",
		"*!i.mei* Meina v6",
		"*!i.wai* Waifu v5",
		"",
		"_A1111 utilities_",
		"*!i.reso* Set custom resolution",
		"*!i.bluff* Bluff",
		"*!i.seed* Set or lock seed",
		"",
		"*Natural Language Editing* - [Modal]",
		"*!i.flx* Flux Kontext",
		"",
		"*Legacy Stable Diffusion* - Free",
		"*!i.sxl* Stable Diffusion XL",
		"*!i.std* Stable Diffusion",
		"*!i.some* Something v2",
		"*!i.cntr* Counterfeit",
		"*!i.modi* Modern Disney",
		"*!i.prot* Protogen",
		"*!i.pix* PixelArt",
		"*!i.logo* LogoRedmond",
		"*!i.mid* OpenMidjourney",
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
	case "i.sxl":
		return "Stable Diffusion XL txt2img. Massive breakthrough compared to earlier versions of SD."
	case "i.std":
		return "Stable Diffusion V2.1 txt2img."
	case "i.some":
		return "[SD] Something V2.2 txt2img. Illust anime-style."
	case "i.cntr":
		return "[SD] Counterfeit V2.5 txt2img. Eerie anime-style."
	case "i.modi":
		return "[SD] Modern Disney Diffusion txt2img. Disney-style."
	case "i.prot":
		return "[SD] Protogen x3.4 txt2img. Tuned for photorealism."
	case "i.pix":
		return "[SDXL] Pixel Art LoRa txt2img"
	case "i.logo":
		return "[SDXL] Logo Redmond txt2img. Specializes in creating logo images."
	case "i.mid":
		return "Open source version of Midjourney V4 txt2img."
	case "i.wai":
		return "[A1111] Waifu v5 txt2img. Anime-style. Booru average style for chars."
	case "i.mei":
		return "[A1111] Meina v6 txt2img. Cutesy anime-style."
	case "i.fwa":
		return "[A1111] Fuwa v7 txt2img. Stable clean cutesy anime-style."
	case "i.fwt":
		return "[A1111] Fuwa v7 txt2img. Stable cutesy anime-style, vibrant colors, 2x faster."
	case "i.fws":
		return "[A1111] Fuwa v7 txt2img. Super stable cutesy anime-style, vibrant colors, 2x faster."
	case "i.reso":
		return "[A1111] Set custom resolution for image generation."
	case "i.bluff":
		return "[A1111] Next generation outputs bluff image."
	case "i.flx":
		return "Flux Kontext txtimg2img. Natural Language Editing. [WIP]"
	case "what":
		return "*U wot m8?*"
	default:
		return fmt.Sprint(query, " (2)", "\nSend `!menu` for command list")
	}
}

func MenuCmdChk(msg *events.Message, cmd string) bool {
	switch cmd {
	case "!menu":
		WaText(msg, MenuTop())
		return true
	case "!imgm":
		WaText(msg, MenuImage())
		return true
	case "!what":
		WaText(msg, MenuWhat(WaMsgQry(msg)))
		return true
	case "!ping":
		WaReact(msg, "🆙")
		return true
	}
	return false
}
