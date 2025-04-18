package main

import (
	"fmt"
	"strings"

	"go.mau.fi/whatsmeow/types/events"
)

func MenuTop() string {
	return strings.Join([]string{
		"*Ucukertz WA bot*",
		"*!ai* Youbot",
		"*!cai* ChatGPT",
		"*!gai* Gemini Pro",
		"*!img* Flux-schnell",
		"*!imgm* Advanced image gen menu",
		"*!what* More explanation for commands (ex: '!what ai')",
	}, "\n")
}

func MenuImage() string {
	return strings.Join([]string{
		"*Advanced image gen menu*",
		"*!.sxl* Stable Diffusion XL",
		"*!.std* Stable Diffusion",
		"*!.some* Something v2",
		"*!.cntr* Counterfeit",
		"*!.modi* Modern Disney",
		"*!.prot* Protogen",
		"*!.pix* PixelArt",
		"*!.logo* LogoRedmond",
		"*!.mid* OpenMidjourney",
		"",
		"*A1111* - Premium GPU 💪",
		"*!.reso* Set custom resolution",
		"*!.wai* Waifu XL v5",
		"*!.mei* Meina XL v6",
	}, "\n")
}

func MenuWhat(query string) string {
	switch query {
	case "ai":
		return "YouBot GPT4, ask anything. Capable of surfing the web (fresh info) but sometimes sleeps."
	case "cai":
		return "ChatGPT GPT4-turbo, ask anything. Dec 2023 training cutoff. Start prompt with /play to make it roleplay."
	case "gai":
		return "Gemini Pro, ask anything. Up-to-date info but may refuse to answer."
	case "img":
		return "Flux-schnell txt2img. Distilled version of model superior to SD3."
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
	case "i.reso":
		return "[A1111] Set custom resolution for image generation."
	case "i.wai":
		return "[A1111] Waifu XL v5 txt2img. Normal anime-style."
	case "i.mei":
		return "[A1111] Meina XL v6 txt2img. Cutesy anime-style."
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
