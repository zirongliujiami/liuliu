package main

import (
	"crypto/md5"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"io"
	"net/http"
	"text/template"
	"xorm.io/xorm"
)

type users struct {
	Id       int64
	Username string
	Password string
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var ab *xorm.Engine

func initAB() {
	var err error
	ab, err = xorm.NewEngine("mysql", "root:123456@tcp(127.0.0.1:3306)/user_db")
	if err != nil {
		panic(err)
	}

	if err = ab.Ping(); err != nil {
		panic(err)
	}
	fmt.Println("连接数据库成功!")
}
func encryptPassword(password string) string {
	hasher := md5.New()
	io.WriteString(hasher, password)
	return fmt.Sprintf("%x", hasher.Sum(nil))
}
func getLoginHandler(c *gin.Context) {
	tmpl, err := template.ParseFiles("login.tmpl") // 确保使用正确的模板解析方法
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "模板文件加载失败: " + err.Error()})
		return
	}
	tmpl.Execute(c.Writer, gin.H{})
}
func loginHandler(c *gin.Context) {
	if c.Request.Method == "POST" {
		var loginReq LoginRequest
		// 使用 ShouldBind 而不是 ShouldBindJSON 来处理表单数据
		if err := c.ShouldBind(&loginReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误: " + err.Error()})
			return
		}

		encryptedPassword := encryptPassword(loginReq.Password)
		var user users
		// 验证用户名和密码
		has, err := ab.Where("username = ? AND password=?", loginReq.Username, encryptedPassword).Get(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if !has {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "账号或密码不正确"})
			return
		}
		c.String(http.StatusOK, "登陆成功!")
	}
}

// 监听端口
func init() {
	initAB()
}
func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/login", getLoginHandler)
	router.POST("/login", loginHandler)

	router.Run(":9300")
}
