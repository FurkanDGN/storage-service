package util

import (
	"fmt"
	"github.com/jlaffaye/ftp"
	"io"
	"path/filepath"
	"strings"
	"time"
	"videohub/model"
)

var cacheMap = make(map[string]CacheItem)

type CacheItem struct {
	Size      int64
	ExpiresAt time.Time
}

func ConnectToFtp(server model.FtpServer) (*ftp.ServerConn, error) {
	conn, err := ftp.Dial(server.Address+":"+server.Port, ftp.DialWithTimeout(15*time.Second))
	if err != nil {
		return nil, err
	}

	err = conn.Login(server.Username, server.Password)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func FetchVideoReaderAndConnFromFtp(server model.FtpServer, videoPath string) (*ftp.ServerConn, io.ReadCloser, error) {
	conn, err := ConnectToFtp(server)
	if err != nil {
		return nil, nil, err
	}

	reader, err := conn.Retr(videoPath)
	if err != nil {
		conn.Quit()
		return nil, nil, err
	}

	return conn, reader, nil
}

func FetchPartialVideoFromFtpToReader(server model.FtpServer, videoPath string, start, end int64) (*ftp.ServerConn, io.Reader, error) {
	conn, err := ConnectToFtp(server)
	if err != nil {
		return nil, nil, err
	}

	reader, err := conn.RetrFrom(videoPath, uint64(start))
	if err != nil {
		conn.Quit()
		return nil, nil, err
	}

	limitedReader := io.LimitReader(reader, end-start+1)
	return conn, limitedReader, nil
}

func FetchVideoSizeFromFtp(server model.FtpServer, path string) (int64, error) {
	item, exists := cacheMap[path]
	if exists && time.Now().Before(item.ExpiresAt) {
		return item.Size, nil
	}

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

	cacheMap[path] = CacheItem{
		Size:      size,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 2),
	}

	return size, nil
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
					return fmt.Errorf("failed to create FTP directory %s: %v", currentPath, err)
				}
			}
		}
	}
	return nil
}
