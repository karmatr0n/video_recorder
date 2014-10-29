package recorder

import (
  "fmt"
  "os"
  "path/filepath"
  "strings"
  "os/exec"
  "./db_client"
  "./ftp_client"
  "./rest_client"
)

type Configuration struct {
  Worker   WorkerConf
  Db       db_client.Configuration
  Ftp_src  ftp_client.Configuration
  Ftp_dest ftp_client.Configuration
  Api      rest_client.Configuration
}

type Recorder struct {
  db *db_client.Session
  ftp *ftp_client.Connection
  api *rest_client.WebService
  worker_conf *WorkerConf
}

type WorkerConf struct {
  IpAddress string
  Mencoder string
  Ffmpeg string
  DestDir string
}

func Start(config *db_client.Configuration, worker_conf *WorkerConf) (*Recorder, error) {
  db := new(db_client.Session)
  db, err := db_client.Start(config, worker_conf.IpAddress)
  if err != nil {
    return nil, err
  }
  r := Recorder{db: db, worker_conf: worker_conf}
  return &r, nil
}

func (r *Recorder) AssignImages() (err error) {
  err = r.db.AssignImages()
  if err != nil {
    return err
  }
  return nil
}

func (r *Recorder) StartFTP(config *ftp_client.Configuration) (err error ){
  r.ftp = new(ftp_client.Connection)
  r.ftp, err = ftp_client.Start(config)
  if err != nil {
    return err
  }
  return nil
}

func (r *Recorder) DownloadImages() (err error) {
  err = r.ftp.Download(r.db.ImageList())
  if err != nil {
    return err
  }
  return nil
}

func (r *Recorder) UploadFiles() (err error) {
  files := []string{r.DestFile(), r.ThumbFile()}
  err = r.ftp.Upload(files)
  if err != nil {
    return err
  }
  return nil
}

func (r *Recorder) MakeDestDir() (err error) {
  dir := filepath.Dir(r.DestDir())
  if _, err = os.Stat(dir); os.IsNotExist(err) {
    // os.MkDirAll function doesn't work with long paths
    // err = os.MkdirAll(dir, 0755)
    exec_command := exec.Command("mkdir", "-p", r.DestDir())
    err = exec_command.Run()
    if err != nil {
      return err
    }
  }
  return nil
}

func (r *Recorder) BuildVideo() (err error) {
  if _, err := os.Stat(r.DestFile()); os.IsNotExist(err) {
    err = r.MakeDestDir()
    if err != nil {
      return err
    }

    var cmd_images string
    cmd_images = fmt.Sprintf("mf://%s",strings.Join(r.db.ImageList(), ","))
    exec_command := exec.Command(r.worker_conf.Mencoder, cmd_images,
    "-mf", "w=800:h=600:fps=1:type=jpg",
    "-ovc", "lavc",
    "-lavcopts", "vcodec=mpeg4:mbd=2:trell",
    "-oac", "copy",
    "-o", r.DestFile())
    err = exec_command.Run()
    if err != nil {
      return err
    }
  }
  return nil
}

func (r *Recorder) BuildThumbnail() (err error) {
  if _, err := os.Stat(r.ThumbFile()); os.IsNotExist(err) {
    exec_command := exec.Command(r.worker_conf.Ffmpeg, "-i", r.DestFile(),
    "-t", "0.001",
    "-ss", "0",
    "-vframes", "1",
    "-f", "mjpeg",
    "-s", "320x240",
    r.ThumbFile())
    err = exec_command.Run()
    if err != nil {
      return err
    }
  }
  return nil
}

func (r *Recorder) StartRestClient(conf *rest_client.Configuration) (err error) {
  r.api = new(rest_client.WebService)
  r.api, err = rest_client.Start(conf)
  if err != nil {
    return err
  }
  return nil
}

func (r *Recorder) RegisterVideo() (err error) {
  params := rest_client.Params{
    CameraIp: r.db.CameraIpAddress(),
    Filename: r.VideoFileName(),
    Path: r.DestFile(),
    Thumbfile: r.ThumbFile(),
    BeginsAt: r.db.VideoBeginsAt(),
    EndsAt: r.db.VideoEndsAt(),
  }
  _, err = r.api.Post(&params)

  if err != nil {
    return err
  }

  return nil
}

func (r *Recorder) DestDir() string {
  dirs := []string{r.worker_conf.DestDir, r.db.CameraIpAddress(), r.db.DirBeginsAt()}
  return strings.Join(dirs, "/")
}

func (r *Recorder) VideoFileName() string {
  file_name := fmt.Sprintf("%02d.avi", r.db.VideoBeginsAtMinute())
  return file_name
}

func (r *Recorder) DestFile() string {
  path := []string{r.DestDir(), r.VideoFileName()}
  return strings.Join(path, "/")
}

func (r *Recorder) ThumbFile() string {
  file := fmt.Sprintf("%02d.jpg", r.db.VideoBeginsAtMinute())
  path := []string{r.DestDir(), file}
  return strings.Join(path, "/")
}

func (r *Recorder) RemoveFiles() error {
  files := r.db.ImageList()
  files = append(files, r.DestFile())
  files = append(files, r.ThumbFile())
  for _, f := range files {
     if _, err := os.Stat(f); os.IsNotExist(err) {
       continue
     }
     err := os.Remove(f)
     if err != nil {
       return err
     }
  }
  return nil
}

func (r *Recorder) PrintInfo()  {
  fmt.Println("=====================================================================")
  fmt.Printf("Data source:\n\n")
  fmt.Printf("Camera ip address: %s, Camera Id: %d\n", r.db.CameraIpAddress(),
  r.db.CameraId())
  fmt.Printf("Worker ip address: %s, Worker Id: %d\n", r.db.WorkerIpAddress(),
  r.db.WorkerId())
  fmt.Printf("Video begins at: %s\n", r.db.VideoBeginsAt())
  fmt.Printf("Video ends at: %s\n", r.db.VideoEndsAt())
  fmt.Printf("Video record ID: %d\n\n", r.db.VideoId())
  fmt.Printf("Expected output: \n\n")
  fmt.Printf("Video file: %s\n", r.DestFile())
  fmt.Printf("Thumbnail file: %s\n", r.ThumbFile())
  fmt.Println("=====================================================================")
  fmt.Println("")
}

func (r *Recorder) PrintVideoInfo() {
  r.db.VideoWasUploaded()
  fmt.Printf("\n\n")
  fmt.Printf("Video location: %s\n", r.api.VideoUrl())
  fmt.Printf("Video ID: %d\n", r.api.VideoId())
}

