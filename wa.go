package main

import (
	"bytes"
	"context"
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
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

var syncing = true
var synctime = time.Now()
var meow *whatsmeow.Client

var WaFallen = fmt.Errorf("The request has fallen. Megabytes must free.")

func WaText(msg *events.Message, text string) {
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
		log.Error().Err(err).Msg("TEXT")
	} else {
		log.Debug().Msg("TEXT OK")
		return
	}

	// Fallback to regular send
	rmsg := waE2E.Message{
		Conversation: &text,
	}
	_, err = meow.SendMessage(context.Background(), msg.Info.Chat, &rmsg)
	if err != nil {
		log.Error().Err(err).Msg("TEXT FB")
	} else {
		log.Debug().Msg("TEXT FB OK")
	}
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

func WaImage(msg *events.Message, img []byte, caption string) {
	mime := http.DetectContentType(img)
	upr, err := meow.Upload(context.Background(), img, whatsmeow.MediaImage)
	if err != nil {
		log.Error().Err(err).Msg("IMGUP")
		WaReact(msg, "😢")
		return
	}

	// Create thumbnail
	imgimg, err := WaByte2ImgImg(img)
	if err != nil {
		log.Error().Err(err).Msg("IMGCONV")
		return
	}
	thumbimgimg := resize.Thumbnail(72, 72, imgimg, resize.Lanczos3)
	thumbbuf := new(bytes.Buffer)
	err = jpeg.Encode(thumbbuf, thumbimgimg, &jpeg.Options{Quality: 80})
	if err != nil {
		log.Error().Err(err).Msg("IMGTHUMBENC")
		return
	}
	thumb := thumbbuf.Bytes()
	thumbupr, err := meow.Upload(context.Background(), thumb, whatsmeow.MediaImage)
	if err != nil {
		log.Error().Err(err).Msg("IMGTHUMBUP")
		return
	}

	// Reply with image
	participant := msg.Info.MessageSource.Sender.String()
	ximsg := waE2E.ImageMessage{
		Caption:  &caption,
		Mimetype: &mime,

		URL:           &upr.URL,
		DirectPath:    &upr.DirectPath,
		MediaKey:      upr.MediaKey,
		FileEncSHA256: upr.FileEncSHA256,
		FileSHA256:    upr.FileSHA256,
		FileLength:    &upr.FileLength,

		JPEGThumbnail:       thumb,
		ThumbnailDirectPath: &thumbupr.DirectPath,
		ThumbnailEncSHA256:  thumbupr.FileEncSHA256,
		ThumbnailSHA256:     thumbupr.FileSHA256,

		ContextInfo: &waE2E.ContextInfo{
			StanzaID:      &msg.Info.ID,
			Participant:   &participant,
			QuotedMessage: msg.Message,
		},
	}

	xmsg := waE2E.Message{
		ImageMessage: &ximsg,
	}
	_, err = meow.SendMessage(context.Background(), msg.Info.Chat, &xmsg)
	if err != nil {
		log.Error().Err(err).Msg("IMG")
	} else {
		log.Debug().Msg("IMG OK")
		return
	}

	// Fallback to regular send
	rimsg := waE2E.ImageMessage{
		Caption:  &caption,
		Mimetype: &mime,

		URL:           &upr.URL,
		DirectPath:    &upr.DirectPath,
		MediaKey:      upr.MediaKey,
		FileEncSHA256: upr.FileEncSHA256,
		FileSHA256:    upr.FileSHA256,
		FileLength:    &upr.FileLength,

		JPEGThumbnail:       thumb,
		ThumbnailDirectPath: &thumbupr.DirectPath,
		ThumbnailEncSHA256:  thumbupr.FileEncSHA256,
		ThumbnailSHA256:     thumbupr.FileSHA256,
	}

	rmsg := waE2E.Message{
		ImageMessage: &rimsg,
	}
	_, err = meow.SendMessage(context.Background(), msg.Info.Chat, &rmsg)
	if err != nil {
		log.Error().Err(err).Msg("IMG FB")
	} else {
		log.Debug().Msg("IMG FB OK")
		return
	}
}

func WaVideo(msg *events.Message, video []byte, caption string, gif bool) {
	upr, err := meow.Upload(context.Background(), video, whatsmeow.MediaVideo)
	if err != nil {
		log.Error().Err(err).Msg("VID")
		WaReact(msg, "😢")
		return
	}

	mime := http.DetectContentType(video)
	participant := msg.Info.MessageSource.Sender.String()
	xvmsg := waE2E.VideoMessage{
		Caption:  &caption,
		Mimetype: &mime,

		URL:           &upr.URL,
		DirectPath:    &upr.DirectPath,
		MediaKey:      upr.MediaKey,
		FileEncSHA256: upr.FileEncSHA256,
		FileSHA256:    upr.FileSHA256,
		FileLength:    &upr.FileLength,
		GifPlayback:   &gif,

		ContextInfo: &waE2E.ContextInfo{
			StanzaID:      &msg.Info.ID,
			Participant:   &participant,
			QuotedMessage: msg.Message,
		},
	}

	xmsg := waE2E.Message{
		VideoMessage: &xvmsg,
	}
	_, err = meow.SendMessage(context.Background(), msg.Info.Chat, &xmsg)
	if err != nil {
		log.Error().Err(err).Msg("VID")
	} else {
		log.Debug().Msg("VID OK")
		return
	}

	rivmsg := waE2E.VideoMessage{
		URL:           &upr.URL,
		Mimetype:      &mime,
		FileEncSHA256: upr.FileEncSHA256,
		FileSHA256:    upr.FileSHA256,
		FileLength:    &upr.FileLength,
	}

	rmsg := waE2E.Message{
		VideoMessage: &rivmsg,
	}
	_, err = meow.SendMessage(context.Background(), msg.Info.Chat, &rmsg)
	if err != nil {
		log.Error().Err(err).Msg("VID FB")
	} else {
		log.Debug().Msg("VID FB OK")
		return
	}
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
	WaText(msg, saad)
}

func WaSaadStr(msg *events.Message, sad string) {
	WaSaad(msg, fmt.Errorf(sad))
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

func WaMsgQry(msg *events.Message) string {
	split := strings.Split(WaMsgStr(msg), " ")[1:]
	return strings.Join(split, " ")
}

func WaMsgMedia(msg *events.Message) []byte {
	if img := msg.Message.GetImageMessage(); img != nil {
		res, err := meow.Download(context.Background(), img)
		if err != nil {
			WaSaadStr(msg, "MEDIA IMG GET: "+err.Error())
			return nil
		}
		return res
	} else if video := msg.Message.GetVideoMessage(); video != nil {
		res, err := meow.Download(context.Background(), video)
		if err != nil {
			WaSaadStr(msg, "MEDIA VID GET: "+err.Error())
			return nil
		}
		return res
	} else if audio := msg.Message.GetAudioMessage(); audio != nil {
		res, err := meow.Download(context.Background(), audio)
		if err != nil {
			WaSaadStr(msg, "MEDIA AUD GET: "+err.Error())
			return nil
		}
		return res
	} else if document := msg.Message.GetDocumentMessage(); document != nil {
		res, err := meow.Download(context.Background(), document)
		if err != nil {
			WaSaadStr(msg, "MEDIA DOC GET: "+err.Error())
			return nil
		}
		return res
	}

	return nil
}

func WaMsgMediaQuoted(msg *events.Message) []byte {
	if msg.Message.GetExtendedTextMessage() == nil {
		return nil
	}
	quoted := msg.Message.GetExtendedTextMessage().GetContextInfo().GetQuotedMessage()
	if quoted == nil {
		return nil
	}

	if img := quoted.GetImageMessage(); img != nil {
		res, err := meow.Download(context.Background(), img)
		if err != nil {
			WaSaadStr(msg, "MEDIA QIMG GET: "+err.Error())
			return nil
		}
		return res
	} else if video := quoted.GetVideoMessage(); video != nil {
		res, err := meow.Download(context.Background(), video)
		if err != nil {
			WaSaadStr(msg, "MEDIA QVID GET: "+err.Error())
			return nil
		}
		return res
	} else if audio := quoted.GetAudioMessage(); audio != nil {
		res, err := meow.Download(context.Background(), audio)
		if err != nil {
			WaSaadStr(msg, "MEDIA QAUD GET: "+err.Error())
			return nil
		}
		return res
	} else if document := quoted.GetDocumentMessage(); document != nil {
		res, err := meow.Download(context.Background(), document)
		if err != nil {
			WaSaadStr(msg, "MEDIA QDOC GET: "+err.Error())
			return nil
		}
		return res
	}

	return nil
}

func cmdHandler(msg *events.Message) {
	cmd := strings.Split(WaMsgStr(msg), " ")[0]

	if ok := SdCmdChk(msg, cmd); ok {
		return
	} else if ok := HugCmdChk(msg, cmd); ok {
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
		if ENV_DEV_MODE == "1" && !strings.Contains(WaMsgUser(v), "234") {
			log.Info().Str("user", WaMsgUser(v)).Msg("DEV MODE: Skipping message from user")
			return
		}

		msgstr := WaMsgStr(v)
		log.Debug().Str("msg", msgstr).Msg("Received a message!")

		if len(WaMsgStr(v)) > 0 {
			Gacur(v)
			OutsCheck(v)

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
