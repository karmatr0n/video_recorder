package db_client

import (
  "database/sql"
  _ "github.com/lib/pq"
  "time"
  "fmt"
  "errors"
)

type Configuration struct {
  Host     string
  User     string
  Password string
  Database string
}

type Session struct {
  db *sql.DB
  worker *Worker
  camera *Camera
  video_duration *VideoDuration
  video *Video
  images []Image
}

type Worker struct {
  Id int
  IpAddress string
}

type Camera struct {
  Id        int
  IpAddress string
}

type Video struct {
  Id int
  WorkerId int
  CameraId int
  BeginsAt time.Time
  EndsAt time.Time
  Uploaded bool
}

type VideoDuration struct {
  BeginsAt time.Time
  EndsAt time.Time
}

type Image struct {
  Id int64
  UploadedAt time.Time
  CameraId int64
  Path string
}

func Start(config *Configuration, worker_ip string) (*Session, error) {
  conn := fmt.Sprintf("user=%s password=%s host=%s dbname=%s sslmode=disable",
  config.User, config.Password, config.Host, config.Database)

  db, err := sql.Open("postgres", conn)
  if err != nil {
    return nil, err
  }

  w := Worker{IpAddress: worker_ip}
  s := Session{db: db, worker: &w}

  return &s, nil
}

func (s *Session) GetId() ( err error) {
  err = s.db.QueryRow("SELECT * FROM find_worker_id($1)", s.worker.IpAddress).
  Scan(&s.worker.Id)
  if  err != nil {
    return err
  }
  return nil
}

func (s *Session) VideoByCameraAndWorker() (err error) {
  s.video = new(Video)
  query := "SELECT id, worker_id, camera_id, begins_at, ends_at, uploaded " +
  "FROM videos WHERE camera_id = $1  AND worker_id = $2 " +
  "AND uploaded IS FALSE " +
  "ORDER BY begins_at ASC, created_at ASC LIMIT 1"

  err = s.db.QueryRow(query, s.camera.Id, s.worker.Id).
  Scan(&s.video.Id, &s.video.WorkerId, &s.video.CameraId,
  &s.video.BeginsAt, &s.video.EndsAt, &s.video.Uploaded)

  switch {
  case err == sql.ErrNoRows:
    err = s.AssignVideo()
    if err != nil {
      return err
    } else {
      return nil
    }
  case err != nil:
    return err
  default:
    return nil
  }
}

func (s *Session) Camera() (err error) {
  s.camera = new(Camera)
  query := "SELECT id, ip_address FROM cameras WHERE has_worker IS FALSE " +
  "ORDER BY latest_upload_at ASC LIMIT 1"

  err = s.db.QueryRow(query).Scan(&s.camera.Id, &s.camera.IpAddress)
  if err != nil {
    return err
  }

  return nil
}

func (s *Session) BlockCamera() error {
  _, err := s.db.Exec("UPDATE cameras SET has_worker = 't' WHERE id = $1", s.camera.Id)
  if err != nil {
    return err
  }
  return nil
}

func (s *Session) ReleaseCamera() error {
  _, err := s.db.Exec("UPDATE cameras SET has_worker = 'f' WHERE id = $1", s.camera.Id)
  if err != nil {
    return err
  }
  return nil
}

func (s *Session) VideoDurationByCamera() (err error) {
  s.video_duration = new(VideoDuration)

  query := "SELECT begins_at, ends_at FROM video_date_range_by_camera_id($1)"

  err = s.db.QueryRow(query, s.camera.Id).
  Scan(&s.video_duration.BeginsAt, &s.video_duration.EndsAt)

  if err != nil {
    return err
  }
  return nil
}

func (s *Session) AssignVideo() (err error) {
  var err_str string

  err = s.BlockCamera()
  if err != nil {
    err_str = fmt.Sprintf("Blocking the camera_id: %d, SQL error: %s",
    s.camera.Id, err)
    return errors.New(err_str)
  }

  err = s.VideoDurationByCamera()
  if err != nil {
    err_str = fmt.Sprintf("I can't get the video duration for camera_id: %d, SQL error: %s", s.camera.Id, err)
    return errors.New(err_str)
  }

  query := "INSERT INTO videos (camera_id, worker_id, begins_at, ends_at) " +
  "VALUES ($1, $2, $3, $4)"
  _, err = s.db.Exec(query, s.camera.Id, s.worker.Id, s.video_duration.BeginsAt,
  s.video_duration.EndsAt)
  if err != nil {
    err_str = fmt.Sprintf("Can't insert a new record into the videos table:\n"+
    "(camera_id, worker_id, begins_at, ends_at) VALUES (%d, %d, '%s', '%s'),\n"+
    "SQL error: %s", s.camera.Id, s.worker.Id, &s.video_duration.BeginsAt,
    &s.video_duration.EndsAt, err)
    return errors.New(err_str)
  }

  err = s.VideoByCameraAndWorker()
  if err != nil {
    err_str = fmt.Sprintf("Can't get a new video assigned to the camera_id: %d "+
    "and the worker_id: %d, SQL error:%s", s.camera.Id, s.worker.Id, err)
    return errors.New(err_str)
  }

  err = s.ReleaseCamera()
  if err != nil {
    err_str = fmt.Sprintf("I can't release the camera_id: %d, SQL error: %s",
    s.camera.Id, err)
    return errors.New(err_str)
  }

  return  nil
}

func (s *Session) AssignImages() (err error) {
  var err_str string

  err = s.Camera()
  if err != nil {
    err_str := fmt.Sprintf("There aren't cameras without assigned workes. "+
    "SQL error: %s", err)
    return errors.New(err_str)
  }

  err = s.GetId()
  if err != nil {
    err_str = fmt.Sprintf("I can't get my worker_id: %, SQL error: %s",
    s.worker.Id, err)
    return errors.New(err_str)
  }

  err = s.VideoByCameraAndWorker()
  if err != nil {
    err_str = fmt.Sprintf("I can't get an assigned video. SQL error: %s", err)
    return errors.New(err_str)
  } else {
    query := "SELECT id, camera_id, path, uploaded_at FROM images " +
    "WHERE camera_id = ($1) AND uploaded_at >= ($2)  AND uploaded_at <= ($3) " +
    "AND assigned IS TRUE " +
    "ORDER BY uploaded_at ASC"
    rows, err := s.db.Query(query, s.video.CameraId, s.video.BeginsAt, s.video.EndsAt)
    if err != nil {
      err_str = fmt.Sprintf("Not images available for camera_id: %d, begins_at: %s, "+
      "ends_at: %s.", s.video.CameraId, s.video.BeginsAt, s.video.EndsAt)
      return errors.New(err_str)
    }
    defer rows.Close()

    s.images = make([]Image, 0)
    for rows.Next() {
      var img =  new(Image)
      err = rows.Scan(&img.Id, &img.CameraId, &img.Path, &img.UploadedAt)
      if err != nil {
        return err
      }
      s.images = append(s.images, *img)
    }

    return nil
  }
}

func (s *Session) VideoWasUploaded() error {
  _, err := s.db.Exec("UPDATE videos SET uploaded = 't' WHERE id = $1", s.video.Id)
  if err != nil {
    return err
  }
  return nil
}

func (s *Session) CameraIpAddress() string {
  return s.camera.IpAddress
}

func (s *Session) CameraId() int {
  return s.camera.Id
}

func (s *Session) WorkerIpAddress() string {
  return s.worker.IpAddress
}

func (s *Session) WorkerId() int {
  return s.worker.Id
}

func (s *Session) VideoBeginsAt() string {
  return s.video.BeginsAt.String()
}

func (s *Session) VideoEndsAt() string {
  return s.video.EndsAt.String()
}

func (s *Session) VideoId() int {
  return s.video.Id
}

func (s *Session) VideoBeginsAtMinute() int {
  return s.video.BeginsAt.Minute()
}

func (s *Session) DirBeginsAt() (path string) {
  path = fmt.Sprintf("%d/%d/%d/%d",
  s.video.BeginsAt.Year(),
  s.video.BeginsAt.Month(),
  s.video.BeginsAt.Day(),
  s.video.BeginsAt.Hour(),
  )
  return path
}

func (s *Session) ImageList() (images []string) {
  images = make([]string, 0)
  for _, image := range s.images {
    images = append(images, image.Path)
  }
  return images
}
