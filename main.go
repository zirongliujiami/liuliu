package main

import (
	"crypto/md5"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"io"
	"net/http"
	"path/filepath"
	"text/template"
	"xorm.io/xorm"
)

var db *xorm.Engine

type Users struct {
	Id       int64
	Username string
	Password string
	Avatar   string
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UpdateUsernameRequest struct {
	NewUsername string `json:"newUsername"`
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

func registerHandler(c *gin.Context) {
	var newUser Users
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的注册请求体"})
		return
	}

	hashedPassword := hashPassword(newUser.Password)
	newUser.Password = hashedPassword

	_, err := db.Insert(&newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "注册成功，请登录"})
}

func loginHandler(c *gin.Context) {
	session := sessions.Default(c)
	var loginReq LoginRequest
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误: " + err.Error()})
		return
	}

	encryptedPassword := hashPassword(loginReq.Password)
	var user Users
	has, err := db.Where("username =? AND password=?", loginReq.Username, encryptedPassword).Get(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if has {
		session.Set("username", user.Username)
		session.Set("userID", user.Id)
		session.Save()
		c.JSON(http.StatusOK, gin.H{"message": "登录成功"})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
	}
}

func personalHandler(c *gin.Context) {
	session := sessions.Default(c)
	username := session.Get("username")
	if username == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	tmpl, err := template.ParseFiles("personal.tmpl")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "模板文件加载失败: " + err.Error()})
		return
	}
	tmpl.Execute(c.Writer, nil)
}

func getUserInfoHandler(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("userID")
	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	var user Users
	_, err := db.ID(userID).Get(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"username": user.Username, "avatar": user.Avatar})
}

func updateUsernameHandler(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("userID")
	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	var req UpdateUsernameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求体"})
		return
	}

	newUsername := req.NewUsername
	_, err := db.ID(userID).Update(&Users{Username: newUsername})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户名失败"})
		return
	}

	session.Set("username", newUsername)
	session.Save()
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "用户名更新成功"})
}

func uploadAvatarHandler(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("userID")
	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件上传失败"})
		return
	}

	filename := fmt.Sprintf("%d_%s", userID, file.Filename)
	filePath := filepath.Join("uploads", filename)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
		return
	}

	avatarURL := "/" + filePath
	_, err = db.ID(userID).Update(&Users{Avatar: avatarURL})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新头像失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "avatar": avatarURL})
}

func main() {
	initDB()
	r := gin.Default()
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))
	r.Static("/uploads", "./uploads")
	r.GET("/login", func(c *gin.Context) {
		tmpl, err := template.ParseFiles("login.tmpl")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "模板文件加载失败: " + err.Error()})
			return
		}
		tmpl.Execute(c.Writer, nil)
	})
	r.POST("/register", registerHandler)
	r.POST("/login", loginHandler)
	r.GET("/personal", personalHandler)
	r.GET("/getUserInfo", getUserInfoHandler)
	r.POST("/updateUsername", updateUsernameHandler)
	r.POST("/uploadAvatar", uploadAvatarHandler)

	r.Run(":8080")
}
