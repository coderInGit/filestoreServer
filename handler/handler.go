package handler

import (
	"encoding/json"
	dbplayer "filestoreServer/db"
	"filestoreServer/meta"
	"filestoreServer/store/oss"
	"filestoreServer/util"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		data, err := os.ReadFile("./static/view/index.html")
		if err != nil {
			io.WriteString(w, "internal server error")
			return
		}
		io.WriteString(w, string(data))
	} else if r.Method == "POST" {
		r.ParseForm()
		file, head, err := r.FormFile("file")
		if err != nil {
			fmt.Printf("Failed to get data,err:%s\n", err.Error())
		}
		defer file.Close()
		fileMeta := meta.FileMeta{
			FileName: head.Filename,
			Location: "/tmp/" + head.Filename,
			UploadAt: time.Now().Format("2006-01-02 15:04:05"),
		}

		newFile, err := os.Create(fileMeta.Location)

		if err != nil {
			fmt.Printf("Failed to create file,err:%s\n", err.Error())
			return
		}
		_, err = io.Copy(newFile, file)
		if err != nil {
			fmt.Printf("Failed to save data into file,err:%s\n", err.Error())
			return
		}
		newFile.Seek(0, 0)
		fileMeta.FileSha1 = util.FileSha1(newFile)

		//同时将文件写入ceph存储
		newFile.Seek(0, 0)
		//data, _ := io.ReadAll(newFile)
		//buket := ceph.GetCephBucket("userfile")
		//cephPath := "/ceph/" + fileMeta.FileSha1
		//buket.Put(cephPath, data, "octet-stream", s3.PublicRead)
		//fileMeta.Location = cephPath

		//写入oss存储
		ossPath := "oss/" + fileMeta.FileSha1
		err = oss.Buket().PutObject(ossPath, newFile)
		if err != nil {
			fmt.Printf(err.Error())
			return
		}
		fileMeta.Location = ossPath

		//meta.UpdateFileMeta(fileMeta)
		meta.UpdateFileMetaDB(fileMeta)

		userName := r.Form.Get("username")
		fmt.Println(userName)
		suc := dbplayer.OnUserFileUploadFinished(userName, fileMeta.FileSha1, fileMeta.FileName, fileMeta.FileSize)
		if suc {
			http.Redirect(w, r, "/static/view/home.html", http.StatusFound)
		} else {
			w.Write([]byte("Upload Failed"))
		}
		//http.Redirect(w, r, "upload/suc", http.StatusFound)
	}
}

// UploadSucHandler 上传已完成
func UploadSucHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Upload finished")
}

// GetFileMetaHandler 获取文件信息
func GetFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fileHash := r.Form["filehash"][0]
	//fMeta := meta.GetFileMeta(fileHash)
	fMeta, err := meta.GetFileMetaDB(fileHash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(fMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write(data)
}

// DownloadHandler 下载文件
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fsha1 := r.Form.Get("filehash")
	fm := meta.GetFileMeta(fsha1)
	f, err := os.Open(fm.Location)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octect-stream")
	w.Header().Set("Content-Descrption", "attachment;filename=\""+fm.FileName+"\"")
	w.Write(data)

}

// FileMetaUpdateHandler 更新元信息接口(重命名)
func FileMetaUpdateHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	opType := r.Form.Get("op")
	fileSha1 := r.Form.Get("filehash")
	newFileName := r.Form.Get("filename")
	if opType != "0" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	curFileMeta := meta.GetFileMeta(fileSha1)
	curFileMeta.FileName = newFileName
	meta.UpdateFileMeta(curFileMeta)
	data, err := json.Marshal(curFileMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// FileDeleteHandler 删除文件及元信息
func FileDeleteHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fileSha1 := r.Form.Get("filehash")

	fMeta := meta.GetFileMeta(fileSha1)
	os.RemoveAll(fMeta.Location)

	meta.RemoveFileMeta(fileSha1)
	w.WriteHeader(http.StatusOK)
}

// FileQueryHandler 查询批量的文件元信息
func FileQueryHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	limitCnt, _ := strconv.Atoi(r.Form.Get("limit"))
	username := r.Form.Get("username")
	userFiles, err := dbplayer.QueryUserFileMeta(username, limitCnt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(userFiles)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func TryFastUploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	userName := r.Form.Get("username")
	fileHash := r.Form.Get("filehash")
	fileName := r.Form.Get("filename")
	fileSize, _ := strconv.ParseInt(r.Form.Get("filesize"), 10, 64)

	fileMeta, err := meta.GetFileMetaDB(fileHash)
	if err != nil {
		fmt.Printf(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if fileMeta == nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "秒传失败，请访问普通上传接口",
		}
		w.Write(resp.JSONBytes())
		return
	}
	suc := dbplayer.OnUserFileUploadFinished(userName, fileHash, fileName, fileSize)
	if suc {
		resp := util.RespMsg{
			Code: 0,
			Msg:  "秒传失败",
		}
		w.Write(resp.JSONBytes())
		return
	} else {
		resp := util.RespMsg{
			Code: -2,
			Msg:  "秒传失败，请稍后重试",
		}
		w.Write(resp.JSONBytes())
		return
	}
}
