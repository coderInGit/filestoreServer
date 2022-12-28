package handler

import (
	dbplayer "filestoreServer/db"
	"filestoreServer/util"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	pawSalt = "#890"
)

// SignupHandler 处理用户注册请求
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data, err := ioutil.ReadFile("./static/view/signup.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(data)
		return
	}
	r.ParseForm()
	userName := r.Form.Get("username")
	password := r.Form.Get("password")
	if len(userName) < 3 || len(password) < 5 {
		w.Write([]byte("Invalid parameter"))
		return
	}
	enc_passwd := util.Sha1([]byte(password + pawSalt))
	suc := dbplayer.UserSignup(userName, enc_passwd)
	if suc {
		w.Write([]byte("SUCCESS"))
	} else {
		w.Write([]byte("FAILED"))
	}
}

// SignInHandler 登陆接口
func SignInHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	userName := r.Form.Get("username")
	password := r.Form.Get("password")

	encPasswd := util.Sha1([]byte(password + pawSalt))
	pwdChecked := dbplayer.UserSignin(userName, encPasswd)
	if !pwdChecked {
		w.Write([]byte("FAILED"))
		return
	}
	token := GenToken(userName)
	upRes := dbplayer.UpdateToken(userName, token)
	if !upRes {
		w.Write([]byte("FAILED"))
		return
	}
	w.Write([]byte("http://" + r.Host + "/static/view/home.html"))
}

func GenToken(username string) string {
	ts := fmt.Sprintf("%x", time.Now().Unix())
	tokenPrefix := util.MD5([]byte(username + ts + "_tokenSalt"))
	return tokenPrefix + ts[:8]
}
