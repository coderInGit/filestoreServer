package db

import (
	"database/sql"
	mydb "filestoreServer/db/mysql"
	"fmt"
)

// OnFileUploadFinished 文件上存完成，保存meta
func OnFileUploadFinished(fileHash string, fileName string, fileSize int64,
	fileAddr string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"insert ignore into tbl_file(file_sha1,file_name,file_size,file_addr," +
			"status) values (?,?,?,?,1)")
	if err != nil {
		fmt.Println("Failed to prepare statement,err:" + err.Error())
		return false
	}
	defer stmt.Close()
	ret, err := stmt.Exec(fileHash, fileName, fileSize, fileAddr)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	if rf, err := ret.RowsAffected(); err == nil {
		if rf <= 0 {
			fmt.Printf("File with hash:%s has been uploaded before\n", fileHash)
			return false
		}
	}
	return false
}

type TableFile struct {
	FileHash string
	FileName sql.NullString
	FileSize sql.NullInt64
	FileAddr sql.NullString
}

// GetFileMeta 从mysql获取元信息
func GetFileMeta(fileHash string) (*TableFile, error) {
	stmt, err := mydb.DBConn().Prepare(
		"select file_sha1,file_addr,file_name,file_size from tbl_file where file_sha1=? and status=1 limit 1")
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()
	tfile := &TableFile{}
	err = stmt.QueryRow(fileHash).Scan(&tfile.FileHash, &tfile.FileAddr, &tfile.FileName, &tfile.FileSize)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	return tfile, nil
}
