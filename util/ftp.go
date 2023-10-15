package util

import (
	"github.com/jlaffaye/ftp"
	"time"
	"videohub/model"
	"io"
	"strings"
	"fmt"
	"path/filepath"
	"bytes"
	"io/ioutil"
)

func ConnectToFtp(server model.FtpServer) (*ftp.ServerConn, error) {
	conn, err := ftp.Dial(server.Address+":"+server.Port, ftp.DialWithTimeout(15*time.Second))
	if err != nil {
		return nil, err
	}

	err = conn.Login(server.Username, server.Password)
	if err != nil {
		return nil, err
	}

	
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func FetchVideoFromFtp(server model.FtpServer, videoPath string) (io.Reader, error) {
	conn, err := ConnectToFtp(server)
	if err != nil {
		return nil, err
	}
	defer conn.Quit()

	reader, err := conn.Retr(videoPath)
	if err != nil {
		return nil, err
	}
	return reader, nil
}

func FetchVideoFromFtpToBuffer(server model.FtpServer, videoPath string) (*bytes.Buffer, error) {
	conn, err := ConnectToFtp(server)
	if err != nil {
		return nil, err
	}
	defer conn.Quit()

	reader, err := conn.Retr(videoPath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	buffer := bytes.NewBuffer(data)
	return buffer, nil
}

func EnsureFtpDirectories(conn *ftp.ServerConn, path string) error {
    parts := strings.Split(path, "/")
    currentPath := ""
    
    for _, part := range parts {
        if part != "" {
            currentPath = filepath.Join(currentPath, part)
            _, err := conn.List(currentPath)
            if err != nil {
                err = conn.MakeDir(currentPath)
                if err != nil {
                    return fmt.Errorf("Failed to create FTP directory %s: %v", currentPath, err)
                }
            }
        }
    }
    return nil
}

func FetchVideoSizeFromFtp(server model.FtpServer, path string) (int64, error) {
    conn, err := ftp.Dial(fmt.Sprintf("%s:%s", server.Address, server.Port), ftp.DialWithTimeout(5*time.Second))
    if err != nil {
        return 0, err
    }
    defer conn.Quit()

    err = conn.Login(server.Username, server.Password)
    if err != nil {
        return 0, err
    }

    size, err := conn.FileSize(path)
    if err != nil {
        return 0, err
    }

    return size, nil
}