package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/png"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/nfnt/resize"
	"github.com/samber/lo"
	"go.mau.fi/whatsmeow/types/events"
)

const (
	SDAPI_TIMEOUT     = 240 * time.Second
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
var xbpostneg_default = fmt.Sprintln("zombie, chibi, bald, bad anatomy, poorly drawn, deformed anatomy, deformed fingers, censored, mosaic_censoring, bar_censor, shota, white_pupils, empty_eyes, multicolored_hair," +
	"BREAK lowres, (bad quality, worst quality:1.2), sketch, jpeg artifacts, censor, blurry, watermark, ")

var SdCkpts = map[string]SdCkpt{
	"!m.wai": *new(SdCkpt).Create("wai").Sampling("", "", 0, 0).AddPrompt("", xbpostpos_default, "", xbpostneg_default),
	"!m.mei": *new(SdCkpt).Create("mei").Sampling("", "", 0, 0).AddPrompt("", xbpostpos_default, "", xbpostneg_default),
	"!m.fwt": *new(SdCkpt).Create("fuwa").Sampling("", "Normal", 16, 2).AddPrompt("", xbpostpos_default+SdTurbo, "", xbpostneg_default),
	"!m.fws": *new(SdCkpt).Create("fuwa").Sampling("", "Normal", 16, 2).AddPrompt(SdFws, xbpostpos_default+SdTurbo, "", xbpostneg_default),
	"!m.fwa": *new(SdCkpt).Create("fuwa").Sampling("", "", 0, 0).AddPrompt(SdFwa, SdSugar+xbpostpos_default, "", xbpostneg_default),
}

func SdApi(msg *events.Message, cmd string) {
	ucfg := GenGet(msg)

	bluff := ucfg.Bluff
	if bluff {
		GenSetBluff(msg, false)

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
		WaReplyImg(msg, img, "")
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

	pos, neg := SdProcessPrompt(msg, ckpt)
	ready := SdWarmup(msg)
	if !ready {
		return
	}

	t_cold := time.Since(t_start).Round(time.Second)

	WaReact(msg, "⏳")
	if !strings.Contains(ucfg.Reso.Name, "SDXL") {
		warn := fmt.Sprint("Using non-SDXL reso ", ucfg.Reso.Name, ". Image quality may be affected.")
		WaReplyText(msg, warn)
	}
	log.Info().Msg("[SD POSITIVE] " + pos)
	log.Info().Msg("[SD NEGATIVE] " + neg)

	// Start generation
	attempt := 0
	SdHttpc := HttpcBase().SetBaseURL(ENV_BASEURL_SDAPI).SetBasicAuth(ENV_BAUTH_SDAPI_USER, ENV_BAUTH_SDAPI_PASS).SetTimeout(SDAPI_TIMEOUT)
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
				WaReplyImg(msg, image, caption)
			}
			return
		}
	}
}

func SdUpscale(msg *events.Message, cmd string) {
	cmd = strings.TrimSuffix(cmd, ".up")
	ckpt, ok := SdCkpts[cmd]
	if !ok {
		WaSaadStr(msg, "SD UP CKPT NOT FOUND")
		return
	}

	log.Info().Str("ckpt", ckpt.name).Msg("SD UP START")
	t_start := time.Now()
	defer func() {
		log.Info().Str("took", fmt.Sprintf("%s", time.Since(t_start).Round(time.Second))).Msg("SD UP END")
	}()

	pos, neg := SdProcessPrompt(msg, ckpt)
	ready := SdWarmup(msg)
	if !ready {
		return
	}

	t_cold := time.Since(t_start).Round(time.Second)
	WaReact(msg, "⏳")

	// Prepare image and config
	ucfg := GenGet(msg)
	img := WaMsgMedia(msg)
	if img == nil {
		img = WaMsgMediaQuoted(msg)
		if img == nil {
			WaSaadStr(msg, "No image to upscale ☹️")
			return
		}
	}

	imgimg, err := PicByte2ImgImg(img)
	if err != nil {
		WaSaad(msg, err)
		return
	}
	imgimg = resize.Thumbnail(1344, 1344, imgimg, resize.Lanczos3)
	init_image := new(bytes.Buffer)
	if err := png.Encode(init_image, imgimg); err != nil {
		WaSaad(msg, err)
		return
	}
	target_w := imgimg.Bounds().Dx() * 2
	target_h := imgimg.Bounds().Dy() * 2

	log.Info().Msg("[SD POSITIVE] " + pos)
	log.Info().Msg("[SD NEGATIVE] " + neg)

	upcfg := fmt.Sprintf("%dx%d -> %dx%d (%.1f)", imgimg.Bounds().Dx(), imgimg.Bounds().Dy(), target_w, target_h, ucfg.Denoise.strength)
	log.Info().Msgf("[SD UPSCALE] %s", upcfg)

	// Start generation
	attempt := 0
	SdHttpc := HttpcBase().SetBaseURL(ENV_BASEURL_SDAPI).SetBasicAuth(ENV_BAUTH_SDAPI_USER, ENV_BAUTH_SDAPI_PASS).SetTimeout(SDAPI_TIMEOUT)
	seed := lo.If(ucfg.Seed != -1, ucfg.Seed).Else(GachaRand64(1e9, 9e9))
	for {
		if attempt > SDAPI_MAX_ATTEMPT {
			WaSaadStr(msg, "SD CANNOT REAL GEN")
			return
		}
		body := map[string]any{
			"init_images":        []string{base64.StdEncoding.EncodeToString(init_image.Bytes())},
			"prompt":             pos,
			"negative_prompt":    neg,
			"sampler_name":       ckpt.sampler,
			"scheduler":          ckpt.scheduler,
			"steps":              ckpt.n_sample,
			"cfg_scale":          ckpt.cfg_scale,
			"width":              target_w,
			"height":             target_h,
			"seed":               seed,
			"denoising_strength": ucfg.Denoise.strength,
			"resize_mode":        0,
			"override_settings":  map[string]any{"sd_model_checkpoint": ckpt.name, "CLIP_stop_at_last_layers": 2},
		}
		r, err := SdHttpc.R().SetBody(body).Post("/sdapi/v1/img2img")
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

				caption := fmt.Sprintf("%s | G %s | C %s\n%s\n%d", t_all, t_gen, t_cold, upcfg, seed)
				WaReplyImg(msg, image, caption)
			}
			return
		}
	}
}

func SdProcessPrompt(msg *events.Message, ckpt SdCkpt) (string, string) {
	prompt := new(SdPrompt)
	prompt.prepos = WaMsgPrompt(msg)
	prompt.prepos = strings.ReplaceAll(prompt.prepos, "\n", " ")
	prompt.prepos = strings.ToLower(prompt.prepos)
	prompt.prepos = strings.ReplaceAll(prompt.prepos, "break", "BREAK")

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

	return pos, neg
}

func SdWarmup(msg *events.Message) bool {
	attempt := 0
	SdHttpc := HttpcBase().SetBaseURL(ENV_BASEURL_SDAPI).SetBasicAuth(ENV_BAUTH_SDAPI_USER, ENV_BAUTH_SDAPI_PASS).SetTimeout(SDAPI_TIMEOUT)

	// Check server readiness
	for {
		if attempt > SDAPI_MAX_ATTEMPT {
			WaSaadStr(msg, "SD DED")
			return false
		}
		r, err := SdHttpc.R().Get("/sdapi/v1/prompt-styles")
		if err != nil {
			attempt++
		} else if r.StatusCode() != http.StatusOK {
			if r.StatusCode() == http.StatusTooManyRequests {
				WaReact(msg, "💸")
				return false
			}

			attempt++
		} else if r.StatusCode() == http.StatusOK {
			break
		}
	}
	return true
}

func SdCmdChk(msg *events.Message, cmd string) bool {
	if !strings.HasSuffix(cmd, ".up") {
		if _, ok := SdCkpts[cmd]; ok {
			go SdApi(msg, cmd)
			return true
		}
	} else {
		basecmd := strings.TrimSuffix(cmd, ".up")
		if _, ok := SdCkpts[basecmd]; ok {
			go SdUpscale(msg, cmd)
			return true
		}
	}
	return false
}
