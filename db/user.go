package db

import (
	mydb "filestoreServer/db/mysql"
	"fmt"
)

// UserSignup 通过用户名和密码完成user表的注册操作
func UserSignup(userName string, password string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"insert ignore into tbl_user(`user_name`,`user_pwd`)values (?,?)")
	if err != nil {
		fmt.Println("Failed to insert,err:" + err.Error())
		return false
	}
	defer stmt.Close()
	ret, err := stmt.Exec(userName, password)
	if err != nil {
		fmt.Println("Failed to insert,err:" + err.Error())
		return false
	}
	if rowsAffected, err := ret.RowsAffected(); err == nil && rowsAffected > 0 {
		return true
	}
	return false
}

// UserSignin 判断密码是否一致
func UserSignin(userName string, password string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"select * from tbl_user where user_name=? and user_pwd =?  limit 1")
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	rows, err := stmt.Query(userName, password)
	if err != nil {
		fmt.Println(err.Error())
		return false
	} else if rows == nil {
		fmt.Println("username not found:" + userName)
		return false
	}
	//pRows := mydb.DBConn().ParseRows(rows)
	//if len(pRows) > 0 && string(pRows[0]["userPwd"].([]byte)) == userName {
	//	fmt.Println("username not found:" + userName)
	//	return false
	//}
	return true
}

// UpdateToken 刷新用户登陆过的token
func UpdateToken(username string, token string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"replace into tbl_user_token(`user_name`,`user_token`)values (?,?)")
	if err != nil {
		fmt.Println(err.Error())
	}
	defer stmt.Close()
	_, err = stmt.Exec(username, token)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

type User struct {
	Username     string
	Email        string
	Phone        string
	SignupAt     string
	LastActiveAt string
	Status       int
}

func GetUserInfo(username string) (User, error) {
	user := User{}
	stmt, err := mydb.DBConn().Prepare(
		"select user_name,signup_at from tbl_user where user_name=? limit 1")
	if err != nil {
		fmt.Println(err.Error())
		return user, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(username).Scan(&user.Username, &user.SignupAt)
	if err != nil {
		return user, err
	}
	return user, nil
}
