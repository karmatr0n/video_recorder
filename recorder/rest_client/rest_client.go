package rest_client

import (
  "github.com/jmcvetta/napping"
  "net/http"
  "net/url"
  "errors"
)

type Configuration struct {
  Url string
  Token string
}

type Params struct {
  CameraIp  string
  Path      string
  Thumbfile string
  BeginsAt  string
  EndsAt    string
  Filename  string
}

type WebService struct {
  url string
  token string
  s     *napping.Session
  video *Video
}

type Video struct {
  Id int
  Url string
  Status int
  Error string
}

func Start(config *Configuration) (*WebService, error) {
  if config.Url != "" && config.Token != "" {
    s := napping.Session{}
    h := http.Header{}
    h.Add("X-AUTH-TOKEN", config.Token)
    s.Header = &h
    // s.Log = true
    ws := WebService{url: config.Url, token: config.Token, s: &s}
    return &ws, nil
  } else {
    return nil, errors.New("Specify the url and the token")
  }
}

func (ws *WebService)  Post(p *Params) (status int, err error) {
  payload := url.Values{}
  payload.Set("ip", p.CameraIp)
  payload.Set("video_filename", p.Filename)
  payload.Set("video_path", p.Path)
  payload.Set("video_thumbnail", p.Thumbfile)
  payload.Set("video_start", p.BeginsAt)
  payload.Set("video_end", p.EndsAt)

  v := Video{}

  resp, err := ws.s.Post(ws.url, &payload, &v, nil)
  if err != nil {
    return -1, err
  }

  ws.video = &v

  return resp.Status(), nil
}

func (ws *WebService) VideoUrl() string {
  return ws.video.Url
}

func (ws *WebService) VideoId() int {
  return ws.video.Id
}

