package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"net/http"
	"strings"
	"time"

	"github.com/nfnt/resize"
	"github.com/ucukertz/hfs"
	"go.mau.fi/whatsmeow/types/events"
)

const MODAL_TIMEOUT = 300 * time.Second

func ModQie(msg *events.Message, cmd string) {
	prompt := WaMsgPrompt(msg)
	if len(prompt) == 0 {
		WaReplyText(msg, "Prompt where?")
		return
	}

	s := hfs.NewHfsRaw[any, any](strings.TrimRight(ENV_BASEURL_QIE, "/") + "/gradio_api/call").
		WithTimeout(MODAL_TIMEOUT)

	ucfg := GenGet(msg)
	iimg, ierr := WaMsgMediaOrQuoted(msg)

	imgs := [][]byte{}
	if iimg != nil {
		imgs = append(imgs, iimg)
	} else {
		for _, p := range ucfg.Pics {
			if p == nil {
				break
			}
			imgs = append(imgs, p)
		}
	}
	if len(imgs) == 0 {
		if ierr != nil {
			WaSaad(msg, WaSaadExpired)
			return
		}
		WaReplyText(msg, "No image to edit ☹️")
		return
	}

	iimgarr := []any{}
	for _, p := range imgs {
		fd, err := hfs.NewFileData("input.jpg").FromUpload(*s, p)
		if err != nil {
			WaSaad(msg, err)
			return
		}
		iimgarr = append(iimgarr, map[string]any{"image": fd, "caption": nil})
	}

	log.Info().Msg("MOD QIE start")
	t_start := time.Now()

	const max_reso_dim = 1168
	reso := ucfg.Reso
	target_w, target_h := PicResoAdjust(reso.Width, reso.Height, max_reso_dim)

	// Plain !m.qie follows the input image's ratio; !m.qie.r uses the !reso setting.
	if !strings.HasSuffix(cmd, ".r") {
		imgimg, _ := PicImgImgFromBytes(imgs[0])
		target_w, target_h = PicResoAdjust(imgimg.Bounds().Dx(), imgimg.Bounds().Dy(), max_reso_dim)
	}

	rsp, err := s.Do("/infer",
		iimgarr,
		prompt,
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

	img, err := hfs.GetGalleryImage(rsp[0], 0)
	if err != nil {
		WaSaad(msg, err)
		return
	}

	t_all := time.Since(t_start).Round(time.Second)
	caption := fmt.Sprintf("%s\n%dx%d", t_all, target_w, target_h)
	WaReplyImg(msg, img, caption)
}

func ModKontext(msg *events.Message) {
	t_start := time.Now()
	defer func() {
		log.Info().Str("took", fmt.Sprintf("%s", time.Since(t_start).Round(time.Second))).Msg("KONTEXT END")
	}()

	var reqbody struct {
		Prompt string `json:"prompt,omitempty"`
		Img    string `json:"input_image,omitempty"`
	}

	img, ierr := WaMsgMediaOrQuoted(msg)
	if img == nil {
		if ierr != nil {
			WaSaad(msg, WaSaadExpired)
			return
		}
		WaSaadStr(msg, "No image to edit ☹️")
		return
	}

	imgimg, err := PicImgImgFromBytes(img)
	if err != nil {
		log.Error().Err(err).Msg("IMGCONV")
		WaSaadStr(msg, "IMGCONV: "+err.Error())
		return
	}

	imgr := resize.Thumbnail(1344, 1344, imgimg, resize.Lanczos3)
	thumbbuf := new(bytes.Buffer)
	if err := png.Encode(thumbbuf, imgr); err != nil {
		log.Error().Err(err).Msg("PNG ENCODE")
		WaSaadStr(msg, "PNG ENCODE: "+err.Error())
		return
	}

	reqbody.Img = base64.StdEncoding.EncodeToString(thumbbuf.Bytes())
	reqbody.Prompt = WaMsgPrompt(msg)

	r, err := HttpcBase().SetTimeout(MODAL_TIMEOUT).SetBasicAuth(ENV_BAUTH_SDAPI_USER, ENV_BAUTH_SDAPI_PASS).
		R().SetBody(reqbody).Post(ENV_BASEURL_KONTEXT)
	if err != nil {
		WaSaadStr(msg, "KONTEXT: "+err.Error())
		return
	}
	if r.StatusCode() != http.StatusOK {
		if r.StatusCode() == http.StatusTooManyRequests {
			WaReact(msg, "💸")
			return
		} else {
			WaSaadStr(msg, "KONTEXT DED "+r.Status())
			return
		}
	}

	image, err := base64.StdEncoding.DecodeString(r.String())
	if err != nil {
		WaSaadStr(msg, "KONTEXT DECODE: "+err.Error())
		return
	}
	t_all := time.Since(t_start).Round(time.Second)
	t_str := fmt.Sprintf("%s", t_all)
	WaReplyImg(msg, image, t_str)
}

func ModCmdChk(msg *events.Message, cmd string) bool {
	switch cmd {
	case AdminDevDiff("!x.flx", "!m.flx"):
		go ModKontext(msg)
		return true
	}

	if strings.HasPrefix(cmd, AdminDevDiff("!x.qie", "!m.qie")) {
		go ModQie(msg, cmd)
		return true
	}

	return false
}
