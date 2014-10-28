package ftp_client

import (
  "io/ioutil"
  "os"
  "fmt"
  "path/filepath"
  "errors"
  "strings"
  "github.com/jlaffaye/ftp"
)

type Configuration struct {
  Host string
  User string
  Password string
}

type Connection struct {
  ftp *ftp.ServerConn
}

func Start(config *Configuration) (*Connection, error) {

  f := new(ftp.ServerConn)
  f, err := ftp.Connect(config.Host + ":21")
  if err != nil {
    return nil, err
  }

  err = f.Login(config.User, config.Password)
  if err != nil {
    return nil, err
  }

  defer f.NoOp()
  c := Connection{ftp: f}

  return &c, nil
}

func (c *Connection) Download(files []string) (err error) {

  for _, f := range files {
    if _, err := os.Stat(f); os.IsNotExist(err) {
      dir := filepath.Dir(f)
      if _, err := os.Stat(dir); os.IsNotExist(err) {
        os.MkdirAll(dir, 0755)
      }

      r, err := c.ftp.Retr(f)
      if err != nil {
        err_str := fmt.Sprintf("I can't download the file: %s, FTP error: %s", f, err)
        return errors.New(err_str)
      }

      stream, err := ioutil.ReadAll(r)
      err = ioutil.WriteFile(f, stream, 0644)

      r.Close()
      if err != nil {
        return err
      }
    }
  }

  c.ftp.Quit()
  return nil
}

func (c *Connection) MkRemoteDirForFile(file string) {
  base_dir := filepath.Dir(file)
  subdirs := strings.Split(base_dir, "/")
  path := ""
  for _, dir := range subdirs {
    if dir != "" {
      path = path + "/" + dir
      _ = c.ftp.MakeDir(path)
    }
  }
}

func (c *Connection) Upload(files []string) (err error) {

  for _, f := range files {
    _ = c.ftp.Delete(f)
    c.MkRemoteDirForFile(f)
    stream, _ := os.Open(f)
    err = c.ftp.Stor(f, stream)
    if err != nil {
      err_str := fmt.Sprintf("I can't store the file: %s, FTP error: %s", f, err)
      return errors.New(err_str)
    }
  }

  c.ftp.Quit()
  return nil
}
