package main

import (
  "fmt"
  "encoding/json"
  "io/ioutil"
  "os"
  "time"
  "flag"
  "./recorder"
)

var defaultConf string

func init() {
  const (
    conf_path = "/etc/video_recorder.json"
    usage = "The configuration file is missed"
  )
  flag.StringVar(&defaultConf, "config", conf_path, usage)
  flag.StringVar(&defaultConf, "c", conf_path, usage+" (shorthand)")
}

func main() {

  flag.Parse()
  filename := defaultConf
  filebyte, err := ioutil.ReadFile(filename)
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }

  var conf recorder.Configuration
  err = json.Unmarshal(filebyte, &conf)
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }

  r, err := recorder.Start(&conf.Db, &conf.Worker)
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }

  err = r.AssignImages()
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }

  r.PrintInfo()

  err = r.StartFTP(&conf.Ftp_src)
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }

  fmt.Printf("Downloading images starts at: %s \n", time.Now().String())
  err = r.DownloadImages()
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  } else {
    fmt.Printf("Downloading images ends at:   %s\n", time.Now().String())
  }

  fmt.Printf("Video recording starts at:    %s\n", time.Now().String())
  err = r.BuildVideo()
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }

  fmt.Printf("Video recording ends at:      %s\n", time.Now().String())
  err = r.BuildThumbnail()
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }

  err = r.StartFTP(&conf.Ftp_dest)
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }

  fmt.Printf("Video uploading starts at:    %s\n", time.Now().String())
  err = r.UploadFiles()
  if err != nil {
    fmt.Println(err)
    os.Exit(1)

  } else {
    fmt.Printf("Video uploading ends at:      %s\n", time.Now().String())
  }

  err = r.StartRestClient(&conf.Api)
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
  err = r.RegisterVideo()
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  } else {
    r.PrintVideoInfo()
  }

  fmt.Println("")
  fmt.Printf("Removing generated files:         %s\n", time.Now().String())
  err = r.RemoveFiles()
  if err != nil {
    fmt.Println(err)
    os.Exit(1)

  } else {
    fmt.Printf("Removing files ends at:           %s\n", time.Now().String())
  }

  os.Exit(0)
}
