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

var db *sql.DB

// 连接数据库
func initDB() {
	var err error
	db, err = sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/user_db") // 替换为你的MySQL用户、密码和数据库名
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("连接数据库成功!")
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		id := r.FormValue("id")
		username := r.FormValue("username")
		password := r.FormValue("password")
		hashedPassword := hashPassword(password)

		//插入数据
		stmt, err := db.Prepare("INSERT INTO users(id,username, password) VALUES(?, ?,?)")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(id, username, hashedPassword)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "注册成功!")
	} else {
		t, err := template.ParseFiles("register.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		t.Execute(w, nil)
	}
}

func main() {
	initDB()
	http.HandleFunc("/register", registerHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
func hashPassword(password string) string {
	hasher := md5.New()
	io.WriteString(hasher, password)
	return fmt.Sprintf("%x", hasher.Sum(nil))
}
