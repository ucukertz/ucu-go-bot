package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mdp/qrterminal/v3"
	"github.com/nfnt/resize"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

var syncing = true
var synctime = time.Now()
var meow *whatsmeow.Client

var WaSaadFallen = fmt.Errorf("The request has fallen. Megabytes must free.")

func WaSendText(chat types.JID, text string) {
	rmsg := waE2E.Message{
		Conversation: &text,
	}
	_, err := meow.SendMessage(context.Background(), chat, &rmsg)
	if err != nil {
		log.Error().Err(err).Msg("WA TEXT")
	} else {
		log.Debug().Msg("WA TEXT OK")
	}
}

func WaReplyText(msg *events.Message, text string) {
	participant := msg.Info.MessageSource.Sender.String()
	xmsg := waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: &text,
			ContextInfo: &waE2E.ContextInfo{
				StanzaID:      &msg.Info.ID,
				Participant:   &participant,
				QuotedMessage: msg.Message,
			},
		},
	}
	_, err := meow.SendMessage(context.Background(), msg.Info.Chat, &xmsg)
	if err != nil {
		log.Error().Err(err).Msg("WA TEXT REPLY")
	} else {
		log.Debug().Msg("WA TEXT REPLY OK")
		return
	}

	// Fallback to regular send
	WaSendText(msg.Info.Chat, text)
}

func WaByte2ImgImg(b []byte) (image.Image, error) {
	imgimg, _, err := image.Decode(bytes.NewReader(b))
	if err == nil {
		return imgimg, nil
	}
	pngimg, err := png.Decode(bytes.NewReader(b))
	if err == nil {
		return pngimg, nil
	}
	jpegimg, err := jpeg.Decode(bytes.NewReader(b))
	if err == nil {
		return jpegimg, nil
	}
	return nil, fmt.Errorf("unsupported image format")
}

type WaUploadedBytes struct {
	mime string
	upr  whatsmeow.UploadResponse
}

type WaUploadedImage struct {
	upb      WaUploadedBytes
	thumb    []byte
	thumbupr whatsmeow.UploadResponse
}

func WaImageUpload(img []byte) (WaUploadedImage, error) {
	mime := http.DetectContentType(img)
	upr, err := meow.Upload(context.Background(), img, whatsmeow.MediaImage)
	if err != nil {
		return WaUploadedImage{}, fmt.Errorf("IMGUP: %w", err)
	}

	// Create thumbnail
	imgimg, err := WaByte2ImgImg(img)
	if err != nil {
		return WaUploadedImage{}, fmt.Errorf("IMGCONV: %w", err)
	}

	thumbimgimg := resize.Thumbnail(72, 72, imgimg, resize.Lanczos3)
	thumbbuf := new(bytes.Buffer)
	err = jpeg.Encode(thumbbuf, thumbimgimg, &jpeg.Options{Quality: 80})
	if err != nil {
		return WaUploadedImage{}, fmt.Errorf("IMGTHUMBENC: %w", err)
	}
	thumb := thumbbuf.Bytes()
	thumbupr, err := meow.Upload(context.Background(), thumb, whatsmeow.MediaImage)
	if err != nil {
		return WaUploadedImage{}, fmt.Errorf("IMGTHUMBUP: %w", err)
	}

	OkUpload := WaUploadedImage{
		upb: WaUploadedBytes{
			mime: mime,
			upr:  upr,
		},
		thumb:    thumb,
		thumbupr: thumbupr,
	}

	return OkUpload, nil
}

func WaImageBuildE2e(isReply bool, upi WaUploadedImage, caption string, msg *events.Message) (*waE2E.Message, error) {
	if isReply && msg == nil {
		return nil, fmt.Errorf("no message to reply to")
	}

	res := upi.upb.upr
	thumbRes := upi.thumbupr
	imsg := &waE2E.ImageMessage{
		Caption:             &caption,
		Mimetype:            &upi.upb.mime,
		URL:                 &res.URL,
		DirectPath:          &res.DirectPath,
		MediaKey:            res.MediaKey,
		FileEncSHA256:       res.FileEncSHA256,
		FileSHA256:          res.FileSHA256,
		FileLength:          &res.FileLength,
		JPEGThumbnail:       upi.thumb,
		ThumbnailDirectPath: &thumbRes.DirectPath,
		ThumbnailEncSHA256:  thumbRes.FileEncSHA256,
		ThumbnailSHA256:     thumbRes.FileSHA256,
	}

	if isReply {
		participant := msg.Info.MessageSource.Sender.String()
		imsg.ContextInfo = &waE2E.ContextInfo{
			StanzaID:      &msg.Info.ID,
			Participant:   &participant,
			QuotedMessage: msg.Message,
		}
	}

	return &waE2E.Message{
		ImageMessage: imsg,
	}, nil
}

func WaSendImg(chat types.JID, img []byte, caption string) {
	up, err := WaImageUpload(img)
	if err != nil {
		log.Error().Err(err).Msg("IMG")
		return
	}

	WaSendImgUp(chat, up, caption)
}

func WaSendImgUp(chat types.JID, upi WaUploadedImage, caption string) {
	e2e, err := WaImageBuildE2e(false, upi, caption, nil)
	if err != nil {
		log.Error().Err(err).Msg("IMG BUILD")
		return
	}

	_, err = meow.SendMessage(context.Background(), chat, e2e)
	if err != nil {
		log.Error().Err(err).Msg("WA IMG")
	} else {
		log.Debug().Msg("WA IMG OK")
		return
	}
}

func WaReplyImg(msg *events.Message, img []byte, caption string) {
	up, err := WaImageUpload(img)
	if err != nil {
		log.Error().Err(err).Msg("IMG")
		WaReact(msg, "😢")
		return
	}

	e2e, err := WaImageBuildE2e(true, up, caption, msg)
	if err != nil {
		log.Error().Err(err).Msg("IMG BUILD")
		return
	}
	_, err = meow.SendMessage(context.Background(), msg.Info.Chat, e2e)
	if err != nil {
		log.Error().Err(err).Msg("IMG")
	} else {
		log.Debug().Msg("IMG OK")
		return
	}

	// Fallback to regular send
	WaSendImgUp(msg.Info.Chat, up, caption)
}

func WaBuildVidE2e(isReply bool, upb WaUploadedBytes, msg *events.Message, caption string, gif bool) (*waE2E.Message, error) {
	if isReply && msg == nil {
		return nil, fmt.Errorf("no message to reply to")
	}

	res := upb.upr
	vmsg := &waE2E.VideoMessage{
		Caption:       &caption,
		Mimetype:      &upb.mime,
		URL:           &res.URL,
		DirectPath:    &res.DirectPath,
		MediaKey:      res.MediaKey,
		FileEncSHA256: res.FileEncSHA256,
		FileSHA256:    res.FileSHA256,
		FileLength:    &res.FileLength,
		GifPlayback:   &gif,
	}

	if isReply {
		participant := msg.Info.MessageSource.Sender.String()
		vmsg.ContextInfo = &waE2E.ContextInfo{
			StanzaID:      &msg.Info.ID,
			Participant:   &participant,
			QuotedMessage: msg.Message,
		}
	}

	return &waE2E.Message{
		VideoMessage: vmsg,
	}, nil
}

func WaSendVid(chat types.JID, video []byte, caption string, gif bool) {
	upr, err := meow.Upload(context.Background(), video, whatsmeow.MediaVideo)
	if err != nil {
		log.Error().Err(err).Msg("VID")
		return
	}
	mime := http.DetectContentType(video)
	upb := WaUploadedBytes{
		mime: mime,
		upr:  upr,
	}
	WaSendVidUp(chat, upb, caption, gif)
}

func WaSendVidUp(chat types.JID, upb WaUploadedBytes, caption string, gif bool) {
	e2e, err := WaBuildVidE2e(false, upb, nil, caption, gif)
	if err != nil {
		log.Error().Err(err).Msg("VID BUILD")
		return
	}

	_, err = meow.SendMessage(context.Background(), chat, e2e)
	if err != nil {
		log.Error().Err(err).Msg("VID")
	} else {
		log.Debug().Msg("VID OK")
	}
}

func WaReplyVid(msg *events.Message, video []byte, caption string, gif bool) {
	upr, err := meow.Upload(context.Background(), video, whatsmeow.MediaVideo)
	if err != nil {
		log.Error().Err(err).Msg("VID")
		WaReact(msg, "😢")
		return
	}

	mime := http.DetectContentType(video)
	upb := WaUploadedBytes{
		mime: mime,
		upr:  upr,
	}

	e2e, err := WaBuildVidE2e(true, upb, msg, caption, gif)
	if err != nil {
		log.Error().Err(err).Msg("VID BUILD")
		return
	}

	_, err = meow.SendMessage(context.Background(), msg.Info.Chat, e2e)
	if err != nil {
		log.Error().Err(err).Msg("VID")
	} else {
		log.Debug().Msg("VID OK")
		return
	}

	// Fallback to regular send
	WaSendVidUp(msg.Info.Chat, upb, caption, gif)
}

func WaReact(msg *events.Message, emoji string) {
	chat := msg.Info.Chat
	sender := msg.Info.Sender
	target := msg.Info.ID
	_, err := meow.SendMessage(context.Background(), chat, meow.BuildReaction(chat, sender, target, emoji))
	if err != nil {
		log.Error().Err(err).Msg("REACT")
	} else {
		log.Debug().Msg("REACT OK")
	}
}

func WaSaad(msg *events.Message, err error) {
	saad := fmt.Sprint("Saad. Bot errored. -> ", err)
	WaReplyText(msg, saad)
}

func WaSaadStr(msg *events.Message, sad string) {
	WaSaad(msg, errors.New(sad))
}

func WaMsgUser(msg *events.Message) string {
	return msg.Info.Sender.User
}

func WaMsgStr(msg *events.Message) string {
	if msg == nil {
		return ""
	}

	if conversation := msg.Message.GetConversation(); len(conversation) > 0 {
		return conversation
	}
	if extendedMsg := msg.Message.GetExtendedTextMessage().GetText(); len(extendedMsg) > 0 {
		return extendedMsg
	}
	if imageMsg := msg.Message.GetImageMessage(); imageMsg != nil && imageMsg.Caption != nil {
		return *imageMsg.Caption
	}
	if videoMsg := msg.Message.GetVideoMessage(); videoMsg != nil && videoMsg.Caption != nil {
		return *videoMsg.Caption
	}
	if documentMsg := msg.Message.GetDocumentMessage(); documentMsg != nil && documentMsg.Caption != nil {
		return *documentMsg.Caption
	}

	return ""
}

func WaMsgPrompt(msg *events.Message) string {
	split := strings.Split(WaMsgStr(msg), " ")[1:]
	return strings.Join(split, " ")
}

func WaE2eMedia(e2e *waE2E.Message) []byte {
	if img := e2e.GetImageMessage(); img != nil {
		res, err := meow.Download(context.Background(), img)
		if err != nil {
			return nil
		}
		return res
	}
	if video := e2e.GetVideoMessage(); video != nil {
		res, err := meow.Download(context.Background(), video)
		if err != nil {
			return nil
		}
		return res
	}
	if audio := e2e.GetAudioMessage(); audio != nil {
		res, err := meow.Download(context.Background(), audio)
		if err != nil {
			return nil
		}
		return res
	}
	if document := e2e.GetDocumentMessage(); document != nil {
		res, err := meow.Download(context.Background(), document)
		if err != nil {
			return nil
		}
		return res
	}
	return nil
}

func WaMsgMedia(msg *events.Message) []byte {
	e2e := msg.Message
	return WaE2eMedia(e2e)
}

func WaMsgMediaQuoted(msg *events.Message) []byte {
	if msg.Message.GetExtendedTextMessage() == nil {
		return nil
	}
	quoted := msg.Message.GetExtendedTextMessage().GetContextInfo().GetQuotedMessage()
	if quoted == nil {
		return nil
	}

	return WaE2eMedia(quoted)
}

func cmdHandler(msg *events.Message) {
	cmd := strings.Split(WaMsgStr(msg), " ")[0]

	if ok := SdCmdChk(msg, cmd); ok {
		return
	} else if ok := HugLegacyCmdChk(msg, cmd); ok {
		return
	} else if ok := ChatCmdChk(msg, cmd); ok {
		return
	} else if ok := OutsCmdChk(msg, cmd); ok {
		return
	} else if ok := MenuCmdChk(msg, cmd); ok {
		return
	} else if ok := GacurCmdChk(msg, cmd); ok {
		return
	} else if ok := AdminCmdChk(msg, cmd); ok {
		return
	}
}

func eventHandler(evt any) {
	switch v := evt.(type) {
	case *events.Message:
		if syncing {
			if time.Now().Before(synctime.Add(5 * time.Second)) {
				log.Trace().Msg("Syncing...")
				synctime = time.Now()
				return
			} else {
				syncing = false
				log.Info().Msg("Sync done!")
			}
		}
		if v.Info.IsFromMe {
			return
		}

		OutsCapture(v)
		if ENV_DEV_MODE == "1" && !IsAdmin(v) {
			log.Info().Str("user", WaMsgUser(v)).Msg("DEV MODE: Skipping message from user")
			return
		}

		msgstr := WaMsgStr(v)
		log.Debug().Str("from", v.Info.Sender.User).Str("msg", msgstr).Msg("Received a message!")

		if len(WaMsgStr(v)) > 0 {
			Gacur(v)

			if strings.HasPrefix(msgstr, "!") {
				cmdHandler(v)
			}
		}

	case *events.PermanentDisconnect:
		panic("PERMANENT DISCONNECT")
	}
}

func WaInit() {
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	container, err := sqlstore.New(context.Background(), "sqlite3", "./wa-login.db?_foreign_keys=on", dbLog)
	if err != nil {
		panic(err)
	}
	deviceStore, err := container.GetFirstDevice(context.Background())
	if err != nil {
		panic(err)
	}
	clientLog := waLog.Stdout("Client", "INFO", true)
	meow = whatsmeow.NewClient(deviceStore, clientLog)
	meow.AddEventHandler(eventHandler)

	if meow.Store.ID == nil {
		// No ID stored, new login
		qrChan, _ := meow.GetQRChannel(context.Background())
		err = meow.Connect()
		if err != nil {
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				// Render the QR code here
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				fmt.Println("QR code:", evt.Code)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		// Already logged in, just connect
		err = meow.Connect()
		if err != nil {
			panic(err)
		}
	}
}
