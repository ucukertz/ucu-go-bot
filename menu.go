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
		"*!img* Z-Image Turbo [ZeroGPU]",
		"*!img.z* Image gen with ZeroGPU",
		"*!img.m* Image gen with Modal",
		"*!img.e* Image editing",

		"*!what* More explanation for commands (ex: '!what ai')",
	}, "\n")
}

func MenuImgm() string {
	return strings.Join([]string{
		"*Image gen with Modal*",
		"",
		"*Stable Diffusion*",
		"*!m.fwa* Fuwa v7 Anthro",
		"*!m.fws* Fuwa v7 Super",
		"*!m.fwt* Fuwa v7 Turbo",
		"*!m.mei* Meina v6",
		"*!m.wai* Waifu v5",
		"Add .up at the end for upscaling `ex: !m.fws.up`",
		"",
		"_Utilities_",
		"*!reso* Set custom resolution",
		"*!resos* Show resolution choices",
		"*!den* Set denoise strength",
		"*!dens* Show denoise strength choices",
		"*!bluff* Bluff",
		"*!seed* Set or lock seed",
		"",
	}, "\n")
}

func MenuImgz() string {
	return strings.Join([]string{
		"*Image gen with ZeroGPU*",
		"",
		"*!z.zit* Z-Image Turbo",
		"*!z.wai* Wai Illu v12-16",
		"",
		"_Utilities_",
		"*!reso* Set custom resolution",
		"*!resos* Show resolution choices",
		"!z.tag Get booru tags for an image",
		"",
		"*Legacy Stable Diffusion*",
		"*!sxl* Stable Diffusion XL",
	}, "\n")
}

func MenuImge() string {
	return strings.Join([]string{
		"*Image editing*",
		"",
		"*Natural Language Editing*",
		"*!m.flx* Flux Kontext",
		"",
		"*Specialized editing*",
		"!*z.qma* Qwen Many Angles",
	}, "\n")
}

func MenuWhat(query string) string {
	switch query {
	case "ai":
		return "Gemini, ask anything. Can process pictures and documents."
	case "cai":
		return "[DEPRECATED] ChatGPT, Use !ai instead."
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
	case "bluff":
		return "[A1111] Next generation outputs bluff image."
	case "z.zit", "img":
		return "Z-Image Turbo txt2img. General purpose image gen with natural language."
	case "z.wai":
		return "Wai Illus v12-16 txt2img. Most popular public booru SDXL model. You can choose v12-16 (ex: !z.wai.15)."
	case "z.tag":
		return "Get booru tags for an image."
	case "z.qma":
		return "Qwen Many Angles img2img. Get different camera angles by setting rotation, elevation, distance."
	case "reso":
		return "[IMG] Set custom resolution for image generation."
	case "den":
		return "[IMG] Set denoise strength for image generation."
	case "seed":
		return "[IMG] Send only `!i.seed` to toggle seed randomness. Send `!i.seed <number>` to set a specific seed."
	case "l.sxl":
		return "Stable Diffusion XL txt2img. Massive breakthrough compared to earlier versions of SD."
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
	case "!img.m":
		WaReplyText(msg, MenuImgm())
		return true
	case "!img.z":
		WaReplyText(msg, MenuImgz())
		return true
	case "!img.e":
		WaReplyText(msg, MenuImge())
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
