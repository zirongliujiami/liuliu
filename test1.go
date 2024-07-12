package main

import (
	"crypto/md5"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"text/template"
	"xorm.io/xorm"

	_ "github.com/go-sql-driver/mysql"
)

var db *xorm.Engine

type Users struct {
	Id       int64
	Username string
	Password string
}

// 连接数据库
func initDB() {
	var err error
	db, err = xorm.NewEngine("mysql", "root:123456@tcp(127.0.0.1:3306)/user_db")
	if err != nil {
		panic(err)
	}

	if err = db.Ping(); err != nil {
		panic(err)
	}
	fmt.Println("连接数据库成功!")
}
func hashPassword(password string) string {
	hasher := md5.New()
	io.WriteString(hasher, password)
	return fmt.Sprintf("%x", hasher.Sum(nil))
}
func LoginHandler(c *gin.Context) {
	tmpl, err := template.ParseFiles("register.tmpl") // 确保使用正确的模板解析方法
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "模板文件加载失败: " + err.Error()})
		return
	}
	tmpl.Execute(c.Writer, gin.H{})
}
func registerHandler(c *gin.Context) {
	if c.Request.Method == "POST" {
		username := c.PostForm("username")
		password := c.PostForm("password")
		hashedPassword := hashPassword(password)

		//插入数据
		newUser := &Users{
			Username: username,
			Password: hashedPassword,
		}
		_, err := db.Insert(newUser)
		if err != nil {
			c.JSON(500, gin.H{"error": "注册失败"})
			return
		}
		c.JSON(200, gin.H{"message": "注册成功!"})
	} else {
		c.JSON(405, gin.H{"message": "支持POST请求"})
	}
}

func main() {
	initDB()
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/register", LoginHandler)
	router.POST("/register", registerHandler)
	router.SetTrustedProxies([]string{"127.0.0.1"})
	router.Run(":8080")
}
