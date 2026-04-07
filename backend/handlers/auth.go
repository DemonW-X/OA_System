package handlers

import (
	"net/http"
	"oa-system/database"
	"oa-system/dto"
	"oa-system/middleware"
	"oa-system/models"
	"regexp"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// 密码正则：长度8位以上，且同时包含字母和数字
// 思路：匹配"含字母部分"和"含数字部分"各自独立校验，再加长度
var (
	pwdHasLetter = regexp.MustCompile(`^[^a-zA-Z]*[a-zA-Z]`)
	pwdHasDigit  = regexp.MustCompile(`^[^0-9]*[0-9]`)
	pwdMinLen    = regexp.MustCompile(`^.{8,}$`)
)

func isValidPassword(pwd string) bool {
	return pwdMinLen.MatchString(pwd) &&
		pwdHasLetter.MatchString(pwd) &&
		pwdHasDigit.MatchString(pwd)
}

func Login(c *gin.Context) {
	var req dto.AuthLoginRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}
	var user models.User
	if err := database.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1, "msg": "用户名或密码错误"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1, "msg": "用户名或密码错误"})
		return
	}
	token, err := middleware.GenerateToken(user.ID, user.Username, user.RealName, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "token生成失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"token":     token,
			"username":  user.Username,
			"real_name": user.RealName,
			"role":      user.Role,
		},
	})
}

func Logout(c *gin.Context) {
	userID := c.GetInt("userID")
	middleware.DeleteToken(userID)
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "已登出"})
}

func GetProfile(c *gin.Context) {
	userID := c.GetInt("userID")
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "用户不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": user})
}

func UpdateProfile(c *gin.Context) {
	userID := c.GetInt("userID")
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "用户不存在"})
		return
	}
	var req dto.AuthUpdateProfileRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	var count int64
	database.DB.Model(&models.User{}).Where("username = ? AND id != ?", req.Username, userID).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "用户名已被占用"})
		return
	}
	user.Username = req.Username
	user.RealName = req.RealName
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": user, "msg": "更新成功"})
}

func ChangePassword(c *gin.Context) {
	userID := c.GetInt("userID")
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "用户不存在"})
		return
	}
	var req dto.AuthChangePasswordRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	// 验证原密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "原密码不正确"})
		return
	}
	// 新密码不能与原密码相同
	if req.OldPassword == req.NewPassword {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "新密码不能与原密码相同"})
		return
	}
	// 校验新密码规则：不少于8位，必须包含英文和数字
	if !isValidPassword(req.NewPassword) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "新密码不少于8位，且必须包含英文字母和数字"})
		return
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "密码加密失败"})
		return
	}
	if err := database.DB.Model(&user).Update("password", string(hashed)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "修改失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "密码修改成功"})
}

func InitAdmin() {
	var count int64
	database.DB.Model(&models.User{}).Count(&count)
	if count > 0 {
		return
	}
	hashed, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	database.DB.Create(&models.User{
		Username: "admin",
		Password: string(hashed),
		RealName: "管理员",
		Role:     "admin",
	})
}
