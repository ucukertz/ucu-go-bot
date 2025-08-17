package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/samber/lo"
	"go.mau.fi/whatsmeow/types/events"
)

const (
	SDAPI_TIMEOUT     = 60 * time.Second
	SDAPI_MAX_ATTEMPT = 10
)

type SdCkpt struct {
	name      string
	sampler   string
	scheduler string
	n_sample  uint8
	cfg_scale float32
	prompt    SdPrompt
}

type SdPrompt struct {
	prepos  string
	postpos string
	preneg  string
	postneg string
}

func (ckpt *SdCkpt) Create(name string) *SdCkpt {
	ckpt.name = name
	return ckpt
}

func (ckpt *SdCkpt) Sampling(sampler string, scheduler string, n_sample uint8, cfg_scale float32) *SdCkpt {
	ckpt.sampler = lo.If(len(sampler) > 0, sampler).Else("Euler a")
	ckpt.scheduler = lo.If(len(scheduler) > 0, scheduler).Else("Automatic")
	ckpt.n_sample = lo.If(n_sample != 0, n_sample).Else(32)
	ckpt.cfg_scale = lo.If(cfg_scale != 0, cfg_scale).Else(7.0)
	return ckpt
}

func (ckpt *SdCkpt) AddPrompt(prepos string, postpos string, preneg string, postneg string) *SdCkpt {
	ckpt.prompt.prepos = fmt.Sprint(ckpt.prompt.prepos, prepos)
	ckpt.prompt.postpos = fmt.Sprint(ckpt.prompt.postpos, postpos)
	ckpt.prompt.preneg = fmt.Sprint(ckpt.prompt.preneg, preneg)
	ckpt.prompt.postneg = fmt.Sprint(ckpt.prompt.postneg, postneg)
	return ckpt
}

var xbpostpos_default = "BREAK vibrant_colors, colorful, masterpiece, best quality, amazing quality, very aesthetic, absurdres, newest, "
var xbpostneg_default = fmt.Sprintln("chibi, bald, bad anatomy, poorly drawn, deformed anatomy, deformed fingers, censored, mosaic_censoring, bar_censor, shota, white_pupils, empty_eyes, multicolored_hair,",
	"BREAK lowres, (bad quality, worst quality:1.2), sketch, jpeg artifacts, censor, blurry, watermark, ")

var SdCkpts = map[string]SdCkpt{
	"!i.wai": *new(SdCkpt).Create("wai").Sampling("", "", 0, 0).AddPrompt("", xbpostpos_default, "", xbpostneg_default),
	"!i.mei": *new(SdCkpt).Create("mei").Sampling("", "", 0, 0).AddPrompt("", xbpostpos_default, "", xbpostneg_default),
	"!i.fwa": *new(SdCkpt).Create("fuwa").Sampling("", "", 0, 0).AddPrompt("", xbpostpos_default, "", xbpostneg_default),
	"!i.fwt": *new(SdCkpt).Create("fuwa").Sampling("", "Normal", 16, 2).AddPrompt("", xbpostpos_default+SdTurbo, "", xbpostneg_default),
	"!i.fws": *new(SdCkpt).Create("fuwa").Sampling("", "Normal", 16, 2).AddPrompt(SdFws, xbpostpos_default+SdTurbo, "", xbpostneg_default),
}

func SdApi(msg *events.Message, cmd string) {
	user := WaMsgUser(msg)
	ucfg := lo.ValueOr(SdActiveUcfg, user, SdDefaultUcfg)

	bluff := ucfg.Bluff
	if bluff {
		ucfg.Bluff = false
		SdSetUcfg(msg, user, ucfg)

		sec := GachaRand64(30, 50)
		time.Sleep(time.Duration(sec) * time.Second)
		WaReact(msg, "⏳")
		sec = GachaRand64(10, 20)
		time.Sleep(time.Duration(sec) * time.Second)
		img, err := os.ReadFile("assets/bluff.jpg")
		if err != nil {
			WaSaad(msg, err)
			return
		}
		WaImage(msg, img, "")
		return
	}

	ckpt, ok := SdCkpts[cmd]
	if !ok {
		WaSaadStr(msg, "SD CKPT NOT FOUND")
		return
	}

	log.Info().Str("ckpt", ckpt.name).Msg("SD START")
	t_start := time.Now()
	defer func() {
		log.Info().Str("took", fmt.Sprintf("%s", time.Since(t_start).Round(time.Second))).Msg("SD END")
	}()

	prompt := new(SdPrompt)
	prompt.prepos = WaMsgQry(msg)
	prompt.prepos = strings.ReplaceAll(prompt.prepos, "\n", " ")
	prompt.prepos = strings.ToLower(prompt.prepos)

	// Handle tuning
	for _, chara := range lo.Keys(SdChars) {
		if strings.Contains(prompt.prepos, chara) {
			log.Info().Str("chara", chara).Msg("SD TUNE")
			prompt.prepos = strings.ReplaceAll(prompt.prepos, chara, "")

			prompt.prepos = lo.If(strings.Contains(prompt.prepos, "cosplay"), strings.ReplaceAll(prompt.prepos, "cosplay", "")).
				Else(fmt.Sprint(SdChars[chara].clothes, ", ", prompt.prepos))
			prompt.prepos = fmt.Sprint(SdChars[chara].traits, ", ", prompt.prepos)

			prompt.postpos = fmt.Sprint(prompt.postpos, ", ", SdChars[chara].postpos)
			prompt.postneg = fmt.Sprint(prompt.postneg, ", ", SdChars[chara].postneg)
		}
	}

	// Negative prompt
	if strings.Contains(prompt.prepos, "nega") {
		split := strings.Split(prompt.prepos, "nega")
		prompt.prepos = split[0] + ", "
		prompt.preneg = split[1] + ", "
	} else {
		prompt.prepos = prompt.prepos + ", "
	}

	// Handle prompt
	pos := strings.Join([]string{ckpt.prompt.prepos, prompt.prepos, prompt.postpos, ckpt.prompt.postpos}, " ")
	neg := strings.Join([]string{ckpt.prompt.preneg, prompt.preneg, prompt.postneg, ckpt.prompt.postneg}, " ")

	for range 3 {
		pos = strings.ReplaceAll(pos, "\n", ", ")
		neg = strings.ReplaceAll(neg, "\n", ", ")
		pos = strings.ReplaceAll(pos, "  ", " ")
		neg = strings.ReplaceAll(neg, "  ", " ")
		pos = strings.ReplaceAll(pos, " , ", ", ")
		neg = strings.ReplaceAll(neg, " , ", ", ")
		pos = strings.ReplaceAll(pos, ", , ,", ",")
		neg = strings.ReplaceAll(neg, ", , ,", ",")
		pos = strings.ReplaceAll(pos, ", ,", ",")
		neg = strings.ReplaceAll(neg, ", ,", ",")
		pos = strings.ReplaceAll(pos, ",,", ",")
		neg = strings.ReplaceAll(neg, ",,", ",")
		pos = strings.TrimLeft(pos, ", ")
		neg = strings.TrimLeft(neg, ", ")
		pos = strings.TrimRight(pos, ", ")
		neg = strings.TrimRight(neg, ", ")
	}

	attempt := 0
	SdHttpc := HttpcBase().SetBaseURL(ENV_BASEURL_SDAPI).SetBasicAuth(ENV_BAUTH_SDAPI_USER, ENV_BAUTH_SDAPI_PASS).SetTimeout(SDAPI_TIMEOUT)

	// Check server readiness
	for {
		if attempt > SDAPI_MAX_ATTEMPT {
			WaSaadStr(msg, "SD DED")
			return
		}
		r, err := SdHttpc.R().Get("/sdapi/v1/prompt-styles")
		if err != nil {
			attempt++
		} else if r.StatusCode() != http.StatusOK {
			if r.StatusCode() == http.StatusTooManyRequests {
				WaText(msg, "MODAL ZERO")
				return
			}

			attempt++
		} else if r.StatusCode() == http.StatusOK {
			break
		}
	}

	t_cold := time.Since(t_start).Round(time.Second)

	WaReact(msg, "⏳")
	log.Info().Msg("[SD POSITIVE] " + pos)
	log.Info().Msg("[SD NEGATIVE] " + neg)

	// Start generation
	seed := lo.If(ucfg.Seed != -1, ucfg.Seed).Else(GachaRand64(1e9, 9e9))
	for {
		if attempt > SDAPI_MAX_ATTEMPT {
			WaSaadStr(msg, "SD CANNOT REAL GEN")
			return
		}
		body := map[string]any{
			"prompt":            pos,
			"negative_prompt":   neg,
			"sampler_name":      ckpt.sampler,
			"scheduler":         ckpt.scheduler,
			"steps":             ckpt.n_sample,
			"cfg_scale":         ckpt.cfg_scale,
			"width":             ucfg.Reso.Width,
			"height":            ucfg.Reso.Height,
			"seed":              seed,
			"override_settings": map[string]any{"sd_model_checkpoint": ckpt.name, "CLIP_stop_at_last_layers": 2},
		}
		r, err := SdHttpc.R().SetBody(body).Post("/sdapi/v1/txt2img")
		if err != nil {
			attempt++
		} else if r.StatusCode() != http.StatusOK {
			attempt++
		} else if r.StatusCode() == http.StatusOK {
			type SdImages struct {
				Images []string `json:"images"`
			}

			var res SdImages
			err = json.Unmarshal(r.Body(), &res)
			if err != nil {
				WaSaad(msg, err)
			}

			if len(res.Images) > 0 {
				image, err := base64.StdEncoding.DecodeString(res.Images[0])
				if err != nil {
					WaSaadStr(msg, "SD INVALID IMG")
					return
				}

				t_all := time.Since(t_start).Round(time.Second)
				t_gen := t_all - t_cold

				caption := fmt.Sprintf("%s | G %s | C %s\n%d", t_all, t_gen, t_cold, seed)
				WaImage(msg, image, caption)
			}
			return
		}
	}
}

type SdReso struct {
	name   string
	Width  int
	Height int
}

var SdResos = map[string]SdReso{
	"sq": {name: "1024x1024", Width: 1024, Height: 1024},
	"w1": {name: "1152x896", Width: 1152, Height: 896},
	"h1": {name: "896x1152", Width: 896, Height: 1152},
	"w2": {name: "1216x832", Width: 1216, Height: 832},
	"h2": {name: "832x1216", Width: 832, Height: 1216},
	"w3": {name: "1344x768", Width: 1344, Height: 768},
	"h3": {name: "768x1344", Width: 768, Height: 1344},
}

type SdUcfg struct {
	Bluff bool
	Reso  SdReso
	Seed  int64
}

var SdActiveUcfg = map[string]SdUcfg{}
var SdDefaultUcfg = SdUcfg{Bluff: false, Reso: SdResos["sq"], Seed: -1}

func SdSetUcfg(msg *events.Message, user string, uconfig SdUcfg) {
	if _, ok := SdActiveUcfg[user]; !ok {
		WaText(msg, "Hi, user "+msg.Info.Sender.User+"!")
	}
	SdActiveUcfg[user] = uconfig
}

func SdCmdChk(msg *events.Message, cmd string) bool {
	if _, ok := SdCkpts[cmd]; ok {
		go SdApi(msg, cmd)
		return true
	}

	user := WaMsgUser(msg)
	ucfg := lo.ValueOr(SdActiveUcfg, user, SdDefaultUcfg)

	switch cmd {
	case "!i.reso":
		if reso, ok := SdResos[WaMsgQry(msg)]; ok {

			ucfg.Reso = reso
			SdSetUcfg(msg, user, ucfg)
			WaText(msg, fmt.Sprintf("A1111 resolution set to %s for you 🫶", reso.name))
		} else {
			WaText(msg, "Resolution not found. Choices: \nsq, h1, h2, h3, w1, w2, w3\n\nExample: `!i.reso sq`")
		}
		return true
	case "!i.bluff":
		ucfg.Bluff = true
		SdSetUcfg(msg, user, ucfg)
		WaReact(msg, "😏")
		return true
	case "!i.seed":
		var seed int64 = ucfg.Seed
		qry := WaMsgQry(msg)
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
		ucfg.Seed = seed
		SdSetUcfg(msg, user, ucfg)
		return true
	case "!s.tune":
		go SdTune(msg)
		return true
	case "!s.bake":
		go SdBake(msg)
		return true
	case "!s.take":
		go SdTake(msg)
		return true
	}

	return false
}
