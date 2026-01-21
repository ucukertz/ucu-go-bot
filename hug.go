package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/ucukertz/hfs"
	"go.mau.fi/whatsmeow/types/events"
)

var (
	HFS_TIMEOUT = 300 * time.Second

	HUG_MAX_ATTEMPT = 10
	HUG_RETRY_SEC   = 20 * time.Second
)

type hugLegacyModel struct {
	url     string
	postpos string
}

func HugZit(msg *events.Message) {
	s := hfs.NewHfs[any, any]("mrfakename-z-image-turbo").
		WithBearerToken(ENV_TOKEN_HUGGING).
		WithTimeout(HFS_TIMEOUT)
	log.Info().Msg("HUG ZIT start")

	query := WaMsgPrompt(msg)
	t_start := time.Now()

	ucfg := GenGet(msg)
	reso := ucfg.Reso
	img, err := s.DoFD("/generate_image", query, reso.Height, reso.Width, 9, 42, true)
	if err != nil {
		WaSaad(msg, err)
		return
	}

	t_all := time.Since(t_start).Round(time.Second)
	dur := fmt.Sprintf("%s", t_all)
	WaReplyImg(msg, img, dur)
}

func HugQie(msg *events.Message, cmd string) {
	ucfg := GenGet(msg)
	iimg := WaMsgMedia(msg)
	if iimg == nil {
		iimg = WaMsgMediaQuoted(msg)
	}
	firstImg := iimg
	iimgarr := []hfs.FileData{}
	if iimg != nil {
		fd, _ := hfs.NewFileData("input").FromBytes(iimg)
		iimgarr = append(iimgarr, *fd)
	} else {
		if ucfg.Pics[0] == nil {
			WaReplyText(msg, "No image to edit ☹️")
			return
		}
		firstImg = ucfg.Pics[0]
		for _, iimg = range ucfg.Pics {
			if iimg == nil {
				break
			}
			fd, _ := hfs.NewFileData("input").FromBytes(iimg)
			iimgarr = append(iimgarr, *fd)
		}
	}

	query := WaMsgPrompt(msg)
	log.Info().Msg("HUG QIE start")
	t_start := time.Now()

	reso := ucfg.Reso
	target_w, target_h := PicAdjustReso(reso.Width, reso.Height, PIC_RESO_1k2)

	if !strings.HasSuffix(cmd, ".r") {
		imgimg, _ := PicByte2ImgImg(firstImg)
		target_w, target_h = PicAdjustReso(imgimg.Bounds().Dx(), imgimg.Bounds().Dy(), PIC_RESO_1k2)
	}

	s := hfs.NewHfs[any, any]("linoyts-qwen-image-edit-2511-fast").
		WithBearerToken(ENV_TOKEN_HUGGING).
		WithTimeout(HFS_TIMEOUT)
	rsp, err := s.Do("/infer",
		iimgarr,
		query,
		42,
		true,
		1,
		4,
		target_h,
		target_w,
		false,
	)
	if err != nil {
		WaSaad(msg, err)
		return
	}

	imgarr := rsp[0].([]any)
	json := imgarr[0].(map[string]any)
	fd, err := hfs.ParseFileData(json["image"])
	if err != nil {
		WaSaad(msg, err)
		return
	}

	fd.URL = strings.ReplaceAll(fd.URL, "space/ca", "space")
	img, err := hfs.FileDataDownload(fd, 60*time.Second)
	if err != nil {
		WaSaad(msg, err)
		return
	}

	t_all := time.Since(t_start).Round(time.Second)
	caption := fmt.Sprintf("%ss\n", t_all)
	WaReplyImg(msg, img, caption)
}

func HugQma(msg *events.Message, cmd string) {
	iimg := WaMsgMedia(msg)
	if iimg == nil {
		iimg = WaMsgMediaQuoted(msg)
		if iimg == nil {
			WaReplyText(msg, "No image to edit ☹️")
			return
		}
	}

	// cmd is expected with this format: !x.qma.rot.elv.dst
	rotmap := map[string]float64{"f": 0, "fr": 45, "r": 90, "br": 135, "b": 180, "bl": 225, "l": 270, "fl": 315}
	elvmap := map[string]float64{"lo": -30, "md": 0, "h1": 30, "h2": 60}
	dstmap := map[string]float64{"0": 0.6, "1": 1.0, "2": 1.4}

	parts := strings.Split(cmd, ".")
	if len(parts) < 5 {
		WaReplyText(msg, strings.Join([]string{
			"Invalid format. Expected: `!z.qma.rot.elv.dst`. Example: `!z.qma.r.h1.1`",
			"",
			"*Rotations*",
			"_f_ -> Front (0°)",
			"_fr_ -> Front-Right (45°)",
			"_r_ -> Right (90°)",
			"_br_ -> Back-Right (135°)",
			"_b_ -> Back (180°)",
			"_bl_ -> Back-Left (225°)",
			"_l_ -> Left (270°)",
			"_fl_ -> Front-Left (315°)",
			"",
			"*Elevations*",
			"_lo_ -> Low (-30°)",
			"_md_ -> Medium (0°)",
			"_h1_ -> High (30°)",
			"_h2_ -> Very High (60°)",
			"",
			"*Distances*",
			"_0_ -> Close (0.6x)",
			"_1_ -> Normal (1.0x)",
			"_2_ -> Far (1.4x)",
		}, "\n"))
		return
	}
	rot := lo.ValueOr(rotmap, parts[2], 0)
	elv := lo.ValueOr(elvmap, parts[3], 0)
	dst := lo.ValueOr(dstmap, parts[4], 1.0)

	log.Info().Msg("HUG QMA start")
	t_start := time.Now()
	fd, err := hfs.NewFileData("input").FromBytes(iimg)
	if err != nil {
		WaSaad(msg, err)
		return
	}

	imgimg, _ := PicByte2ImgImg(iimg)
	w, h := PicAdjustReso(imgimg.Bounds().Dx(), imgimg.Bounds().Dy(), PIC_RESO_1k)

	s := hfs.NewHfs[any, any]("multimodalart-qwen-image-multiple-angles-3d-camera").
		WithBearerToken(ENV_TOKEN_HUGGING).
		WithTimeout(HFS_TIMEOUT)

	img, err := s.DoFD("/infer_camera_edit", fd, rot, elv, dst, 42, true, 1, 4, h, w)
	if err != nil {
		WaSaad(msg, err)
		return
	}
	t_all := time.Since(t_start).Round(time.Second)
	caption := fmt.Sprintf("%s\nRot %.0f° Elv %.0f° Dist x%.1f\n%dx%d", t_all, rot, elv, dst, w, h)
	WaReplyImg(msg, img, caption)
}

func HugWai(msg *events.Message, cmd string) {
	s := hfs.NewHfsRaw[any, any]("https://ibarakidouji-wai-nsfw-illustrious-sdxl.hf.space/call").
		WithBearerToken(ENV_TOKEN_HUGGING).
		WithTimeout(HFS_TIMEOUT)

	query := WaMsgPrompt(msg)

	v := "16"
	// Take string after last dot, use as version if integer
	if parts := strings.Split(cmd, "."); len(parts) > 1 {
		lastPart := parts[len(parts)-1]
		if _, err := strconv.Atoi(lastPart); err == nil {
			// clamp version to 11-16
			if ver, _ := strconv.Atoi(lastPart); ver < 11 {
				v = "11"
			} else if ver > 16 {
				v = "16"
			} else {
				v = lastPart
			}
		}
	}

	t_start := time.Now()
	log.Info().Str("V", v).Msg("HUG WAI start")

	ucfg := GenGet(msg)
	reso := ucfg.Reso

	w, h := PicExpandLow(reso.Width, reso.Height, 1024)
	w, h = Pic2DSnap16(w, h)

	pospos := strings.ReplaceAll(xbpostpos_default, "BREAK", "")

	rsp, err := s.Do("/generate",
		query+", "+pospos,
		strings.ReplaceAll(xbpostneg_default, "BREAK", ""),
		GachaRand64(1, 1e7),
		w,
		h,
		7,
		28,
		"Euler a",
		"v"+v,
		"Custom",
		false,
		0,
		1,
		false,
	)
	if err != nil {
		WaSaad(msg, err)
		return
	}

	imgarr := rsp[0].([]any)
	json := imgarr[0].(map[string]any)
	fd, err := hfs.ParseFileData(json["image"])
	if err != nil {
		WaSaad(msg, err)
		return
	}

	fd.URL = strings.ReplaceAll(fd.URL, "space/ca", "space")
	img, err := hfs.FileDataDownload(fd, 60*time.Second)
	if err != nil {
		WaSaad(msg, err)
		return
	}

	t_all := time.Since(t_start).Round(time.Second)
	caption := fmt.Sprintf("%ss\nv%s", t_all, v)
	WaReplyImg(msg, img, caption)
}

func HugTag(msg *events.Message) {
	iimg := WaMsgMedia(msg)
	if iimg == nil {
		iimg = WaMsgMediaQuoted(msg)
		if iimg == nil {
			WaReplyText(msg, "No image to tag ☹️")
			return
		}
	}

	fd, _ := hfs.NewFileData("input").FromBytes(iimg)

	s := hfs.NewHfs[any, any]("johnny-z-dan-tagger").
		WithBearerToken(ENV_TOKEN_HUGGING).
		WithTimeout(HFS_TIMEOUT)
	rsp, err := s.Do("/process_image", fd)
	if err != nil {
		WaSaad(msg, err)
		return
	}
	tags := rsp[0].(string)
	WaReplyText(msg, tags)
}

func HugRbg(msg *events.Message) {
	iimg := WaMsgMedia(msg)
	if iimg == nil {
		iimg = WaMsgMediaQuoted(msg)
		if iimg == nil {
			WaReplyText(msg, "No image to edit ☹️")
			return
		}
	}

	fd, _ := hfs.NewFileData("input").FromBytes(iimg)

	s := hfs.NewHfs[any, any]("not-lain-background-removal").
		WithBearerToken(ENV_TOKEN_HUGGING).
		WithTimeout(HFS_TIMEOUT)
	rsp, err := s.Do("/image", fd)
	if err != nil {
		WaSaad(msg, err)
		return
	}

	imgarr := rsp[0].([]any)
	img, err := hfs.GetFileData(imgarr[0])
	if err != nil {
		WaSaad(msg, err)
		return
	}

	WaReplyImg(msg, img, "")
}

func HugCmdChk(msg *events.Message, cmd string) bool {
	switch cmd {
	case AdminDevDiff("!x.zit", "!z.zit"), AdminDevDiff("!ximg", "!img"):
		go HugZit(msg)
		return true
	case AdminDevDiff("!x.qie", "!z.qie"):
		go HugQie(msg, cmd)
		return true
	case AdminDevDiff("!x.tag", "!z.tag"):
		go HugTag(msg)
		return true
	case AdminDevDiff("!x.rbg", "!z.rbg"):
		go HugRbg(msg)
		return true
	}

	if strings.HasPrefix(cmd, AdminDevDiff("!x.wai", "!z.wai")) {
		go HugWai(msg, cmd)
		return true
	} else if strings.HasPrefix(cmd, AdminDevDiff("!x.qma", "!z.qma")) {
		go HugQma(msg, cmd)
		return true
	}

	return false
}

var HugLgc = map[string]hugLegacyModel{
	"!l.sxl": {"stabilityai/stable-diffusion-xl-base-1.0", ""},
}

func HugLegacy(msg *events.Message, model string, query string, attempt int) ([]byte, error) {
	r, err := HttpcBase().
		SetBaseURL(ENV_BASEURL_HUGGINGFACE).
		SetAuthToken(ENV_TOKEN_HUGGING).
		SetTimeout(HUG_RETRY_SEC).
		R().SetBody(map[string]any{"inputs": fmt.Sprint(query, ",", HugLgc[model].postpos, lo.RandomString(6, lo.NumbersCharset))}).
		Post(HugLgc[model].url)

	if err != nil || r.StatusCode() != http.StatusOK {
		if attempt == 0 {
			WaReact(msg, "⏳")
		}
		if err != nil {
			return nil, err
		}
		if r.StatusCode() != http.StatusOK {
			return nil, fmt.Errorf("HUG %s", r.Status())
		}
	}
	return r.Body(), nil
}

func HuggingLegacy(msg *events.Message, model string, query string) {
	attempt := 0
	img := []byte{}
	err := fmt.Errorf("HUG ERRINIT")
	for attempt < HUG_MAX_ATTEMPT && err != nil {
		img, err = HugLegacy(msg, model, query, 0)
		attempt++
	}

	if err != nil {
		WaSaad(msg, err)
		return
	}
	WaReplyImg(msg, img, "")
}

func HugLegacyCmdChk(msg *events.Message, cmd string) bool {
	if _, ok := HugLgc[cmd]; ok {
		go HuggingLegacy(msg, cmd, WaMsgPrompt(msg))
		return true
	}

	return false
}
