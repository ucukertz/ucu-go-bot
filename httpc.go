package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/samber/lo"
)

var HttpcBase = resty.New().
	OnBeforeRequest(HttpcMidwareBeforeReq).
	OnSuccess(HttpcMidwareSuccess).
	OnError(HttpcMidwareErr).
	EnableTrace()

type HttpcCtxKey string

func HttpcMidwareBeforeReq(c *resty.Client, r *resty.Request) error {
	id := lo.RandomString(6, lo.LowerCaseLettersCharset)
	fullUrl := c.BaseURL + r.URL
	ctx := context.WithValue(r.Context(), HttpcCtxKey("ID"), id)
	ctx = context.WithValue(ctx, HttpcCtxKey("URL"), fullUrl)
	r.SetContext(ctx)
	log.Debug().Str("ID", id).Str("URL", fullUrl).Msg("HTTPC bgn")
	return nil
}

func HttpcMidwareSuccess(c *resty.Client, r *resty.Response) {
	ti := r.Request.TraceInfo()
	took := fmt.Sprintf("%s", ti.TotalTime.Round(time.Millisecond))
	id := r.Request.Context().Value(HttpcCtxKey("ID")).(string)
	ResMime := http.DetectContentType(r.Body())
	if strings.Contains(ResMime, "image") {
		log.Trace().Str("ID", id).Int("ResCode", r.StatusCode()).Str("ResBody", "[[[IMAGE RESPONSE]]]").Str("Took", took).Msg("HTTPC end")
		return
	} else if strings.Contains(ResMime, "octet") {
		log.Trace().Str("ID", id).Int("ResCode", r.StatusCode()).Str("ResBody", "[[[BINARY RESPONSE]]]").Str("Took", took).Msg("HTTPC end")
		return
	} else if strings.Contains(r.Request.URL, "sdapi") {
		log.Trace().Str("ID", id).Int("ResCode", r.StatusCode()).Str("ResBody", "[[[SDAPI RESPONSE]]]").Str("Took", took).Msg("HTTPC end")
		return
	}
	log.Trace().Str("ID", id).Int("ResCode", r.StatusCode()).Bytes("ResBody", r.Body()).Str("Took", took).Msg("HTTPC end")
}

func HttpcMidwareErr(r *resty.Request, err error) {
	id := r.Context().Value(HttpcCtxKey("ID")).(string)
	url := r.Context().Value(HttpcCtxKey("URL")).(string)
	if v, ok := r.Body.(string); ok {
		log.Error().Err(err).Str("ID", id).Str("URL", url).Str("ReqBody", v).Msg("HTTPC err")
	} else if v, ok := r.Body.([]byte); ok {
		log.Error().Err(err).Str("ID", id).Str("URL", url).Hex("ReqBody", v).Msg("HTTPC err")
	} else {
		log.Error().Err(err).Str("ID", id).Str("URL", url).Strs("ReqContent", r.Header["Content-Type"]).Msg("HTTPC err")
	}

	if v, ok := err.(*resty.ResponseError); ok {
		log.Error().Err(err).Str("ID", id).Bytes("ResBody", v.Response.Body()).Int("ResCode", v.Response.StatusCode()).Msg("HTTPC err")
	}
}
