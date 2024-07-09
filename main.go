package main

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

var ab *sql.DB

func initAB() {
	var err error
	ab, err = sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/user_db") // 替换为你的MySQL用户、密码和数据库名
	if err != nil {
		log.Fatal(err)
	}

	if err = ab.Ping(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("连接数据库成功!")
}
func encryptPassword(password string) string {
	hasher := md5.New()
	io.WriteString(hasher, password)
	return fmt.Sprintf("%x", hasher.Sum(nil))
}
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")
		encryptedPassword := encryptPassword(password)
		// 验证用户名和密码
		var dbPassword string
		err := ab.QueryRow("SELECT password FROM users WHERE username = ?", username).Scan(&dbPassword)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "账号或密码不正确", http.StatusUnauthorized)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if dbPassword == encryptedPassword {
			fmt.Fprintf(w, "登录成功!")
		} else {
			http.Error(w, "账号或密码不正确!", http.StatusUnauthorized)
		}
	} else {
		// 加载页面
		t, err := template.ParseFiles("login.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		t.Execute(w, nil)
	}
}

// 监听端口
func main() {
	initAB()
	http.HandleFunc("/login", loginHandler)
	log.Fatal(http.ListenAndServe(":9300", nil))
}
