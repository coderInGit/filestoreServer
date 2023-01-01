package db

import (
	mydb "filestoreServer/db/mysql"
	"fmt"
	"time"
)

type UserFile struct {
	UserName    string
	FileHash    string
	FileName    string
	FileSize    int64
	UploadAt    string
	LastUpdated string
}

// OnUserFileUploadFinished 更新用户文件表
func OnUserFileUploadFinished(userName, fileHash, fileName string, fileSize int64) bool {
	stmt, err := mydb.DBConn().Prepare(
		"insert ignore into tbl_user_file(`user_name`,`file_sha1`,`file_name`,`file_size`," +
			"`uplaod_at`)values (?,?,?,?,?)")
	if err != nil {
		return false
	}
	defer stmt.Close()
	_, err = stmt.Exec(userName, fileHash, fileName, fileSize, time.Now())
	if err != nil {
		return false
	}
	return true
}

// QueryUserFileMeta 批量获取文件信息
func QueryUserFileMeta(userName string, limit int) ([]UserFile, error) {
	stmt, err := mydb.DBConn().Prepare(
		"select file_sha1,file_name,file_size,upload_at,last_update from" +
			"tbl_user_file where user_name=? limit=?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(userName, limit)
	if err != nil {
		return nil, err
	}
	var userFiles []UserFile
	for rows.Next() {
		ufile := UserFile{}
		err = rows.Scan(&ufile.FileHash, &ufile.FileName, &ufile.FileSize, &ufile.UploadAt, &ufile.LastUpdated)
		if err != nil {
			fmt.Println(err.Error())
			break
		}
		userFiles = append(userFiles, ufile)
	}
	return userFiles, err
}
