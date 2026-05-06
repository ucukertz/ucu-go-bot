package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"resty.dev/v3"
	"github.com/samber/lo"
)

func HttpcBase() *resty.Client {
	return resty.New().
		AddRequestMiddleware(HttpcMidwareBeforeReq).
		AddResponseMiddleware(HttpcMidwareAfterRes).
		SetResponseBodyUnlimitedReads(true).
		EnableTrace()
}

type HttpcCtxKey string

func HttpcMidwareBeforeReq(c *resty.Client, r *resty.Request) error {
	id := lo.RandomString(6, lo.LowerCaseLettersCharset)
	fullUrl := c.BaseURL() + r.URL
	ctx := context.WithValue(r.Context(), HttpcCtxKey("ID"), id)
	ctx = context.WithValue(ctx, HttpcCtxKey("URL"), fullUrl)
	r.SetContext(ctx)
	log.Debug().Str("ID", id).Str("URL", fullUrl).Msg("HTTPC bgn")
	return nil
}

func HttpcMidwareAfterRes(c *resty.Client, r *resty.Response) error {
	id := r.Request.Context().Value(HttpcCtxKey("ID")).(string)
	if err := r.Err; err != nil {
		url := r.Request.Context().Value(HttpcCtxKey("URL")).(string)
		if v, ok := r.Request.Body.(string); ok {
			log.Error().Err(err).Str("ID", id).Str("URL", url).Str("ReqBody", v).Msg("HTTPC err")
		} else if v, ok := r.Request.Body.([]byte); ok {
			log.Error().Err(err).Str("ID", id).Str("URL", url).Hex("ReqBody", v).Msg("HTTPC err")
		} else {
			log.Error().Err(err).Str("ID", id).Str("URL", url).Strs("ReqContent", r.Request.Header["Content-Type"]).Msg("HTTPC err")
		}

		if v, ok := err.(*resty.ResponseError); ok {
			log.Error().Err(err).Str("ID", id).Bytes("ResBody", v.Response.Bytes()).Int("ResCode", v.Response.StatusCode()).Msg("HTTPC err")
		}
		return nil
	}

	ti := r.Request.TraceInfo()
	took := fmt.Sprintf("%s", ti.TotalTime.Round(time.Millisecond))
	ResMime := http.DetectContentType(r.Bytes())
	if strings.Contains(ResMime, "image") {
		log.Trace().Str("ID", id).Int("ResCode", r.StatusCode()).Str("ResBody", "[[[IMAGE RESPONSE]]]").Str("Took", took).Msg("HTTPC end")
		return nil
	} else if strings.Contains(ResMime, "octet") {
		log.Trace().Str("ID", id).Int("ResCode", r.StatusCode()).Str("ResBody", "[[[BINARY RESPONSE]]]").Str("Took", took).Msg("HTTPC end")
		return nil
	} else if strings.Contains(r.Request.URL, "sdapi") {
		log.Trace().Str("ID", id).Int("ResCode", r.StatusCode()).Str("ResBody", "[[[SDAPI RESPONSE]]]").Str("Took", took).Msg("HTTPC end")
		return nil
	}
	log.Trace().Str("ID", id).Int("ResCode", r.StatusCode()).Bytes("ResBody", r.Bytes()).Str("Took", took).Msg("HTTPC end")
	return nil
}

