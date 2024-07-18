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
func ShowRegisterForm(c *gin.Context) {
	tmpl, err := template.ParseFiles("register.tmpl") // 确保register.tmpl文件存在
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "模板文件加载失败: " + err.Error()})
		return
	}
	tmpl.Execute(c.Writer, nil)
}
func registerHandler(c *gin.Context) {
	if c.Request.Method == "POST" {
		// 绑定表单数据到结构体
		var user Users
		if err := c.ShouldBind(&user); err != nil {
			c.JSON(400, gin.H{"error": "无效的请求体"})
			return
		}

		hashedPassword := hashPassword(user.Password)
		user.Password = hashedPassword

		// 插入数据
		_, err := db.Insert(&user)
		if err != nil {
			c.JSON(500, gin.H{"error": "注册失败"})
			return
		}

		c.JSON(200, gin.H{"message": "注册成功!"})
	} else {
		c.JSON(405, gin.H{"message": "仅支持POST请求"})
	}
}

func main() {
	initDB()
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/register", ShowRegisterForm)
	router.POST("/register", registerHandler)
	router.SetTrustedProxies([]string{"127.0.0.1"})
	router.Run(":8080")
}
