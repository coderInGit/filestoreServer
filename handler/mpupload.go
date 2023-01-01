package handler

import (
	rPool "filestoreServer/cache/redis"
	dbplayer "filestoreServer/db"
	"filestoreServer/util"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

// MultipartUploadInfo 初始化信息
type MultipartUploadInfo struct {
	FileHash   string
	FileSize   int
	UploadId   string
	ChunkSize  int
	ChunkCount int
}

func InitialMultipartUploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesize, err := strconv.Atoi(r.Form.Get("filesize"))
	if err != nil {
		w.Write(util.NewRespMsg(-1, "params invalid", nil).JSONBytes())
		return
	}
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()
	upInfo := MultipartUploadInfo{
		FileHash:   filehash,
		FileSize:   filesize,
		UploadId:   username + fmt.Sprintf("%x", time.Now().UnixNano()),
		ChunkSize:  5 * 1024 * 1024,
		ChunkCount: int(math.Ceil(float64(filesize) / (5 * 1024 * 1024))),
	}
	rConn.Do("HSET", "MP_"+upInfo.UploadId, "chunkcount", upInfo.ChunkCount)
	rConn.Do("HSET", "MP_"+upInfo.UploadId, "filehash", upInfo.FileHash)
	rConn.Do("HSET", "MP_"+upInfo.UploadId, "filesize", upInfo.FileSize)
	w.Write(util.NewRespMsg(0, "OK", upInfo).JSONBytes())
}

// CompleteUploadHandler 通知上存合并
func CompleteUploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	upid := r.Form.Get("upload")
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesize := r.Form.Get("filesize")
	filename := r.Form.Get("filename")

	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	data, err := redis.Values(rConn.Do("HGETALL", "MP_"+upid))
	if err != nil {
		w.Write(util.NewRespMsg(-1, "complete upload failed", nil).JSONBytes())
		return
	}
	totalCount := 0
	chunkCount := 0
	for i := 0; i < len(data); i += 2 {
		k := string(data[i].([]byte))
		v := string(data[i].([]byte))
		if k == "chunkcount" {
			totalCount, _ = strconv.Atoi(v)
		} else if strings.HasPrefix(k, "chkidx_") && v == "1" {
			chunkCount++
		}
	}
	if totalCount != chunkCount {
		w.Write(util.NewRespMsg(-2, "invalid request", nil).JSONBytes())
		return
	}
	fsize, _ := strconv.ParseInt(filesize, 10, 64)
	dbplayer.OnFileUploadFinished(filehash, filename, fsize, "")
	dbplayer.OnUserFileUploadFinished(username, filehash, filename, fsize)
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}

// UploadParHandler 上存文件分块
func UploadParHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	//username := r.Form.Get("username")
	uploadID := r.Form.Get("uploadid")
	chunkIndex := r.Form.Get("index")

	rConn := rPool.RedisPool().Get()
	defer rConn.Close()
	fpath := "/data/" + uploadID + "/" + chunkIndex
	os.MkdirAll(path.Dir(fpath), 0744)
	fd, err := os.Create(path.Dir(fpath))
	if err != nil {
		w.Write(util.NewRespMsg(-1, "Upload part faild", nil).JSONBytes())
		return
	}
	buf := make([]byte, 1024*1024)
	for {
		n, err := r.Body.Read(buf)
		fd.Write(buf[:n])
		if err != nil {
			break
		}
	}
	rConn.Do("HSET", "MP_"+uploadID, "chkidx_"+chunkIndex, 1)
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}
