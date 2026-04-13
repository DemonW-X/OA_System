package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"oa-system/database"
	"oa-system/services"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

const tokenTTL = 24 * time.Hour

type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	RealName string `json:"real_name"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// redisTokenKey 执行相关业务逻辑
func redisTokenKey(userID int) string {
	return fmt.Sprintf("token:%d", userID)
}

// GenerateToken 生成业务数据
func GenerateToken(userID int, username, realName, role string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		RealName: realName,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}
	// 存入 Redis
	database.RDB.Set(context.Background(), redisTokenKey(userID), tokenStr, tokenTTL)
	return tokenStr, nil
}

// ParseToken 解析输入数据
func ParseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrTokenInvalidClaims
}

// DeleteToken 删除数据
func DeleteToken(userID int) {
	database.RDB.Del(context.Background(), redisTokenKey(userID))
}

// JWTAuth 执行相关业务逻辑
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未登录"})
			c.Abort()
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := ParseToken(tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "token无效或已过期"})
			c.Abort()
			return
		}
		// 校验 Redis 中的 Token 是否一致
		cached, err := database.RDB.Get(context.Background(), redisTokenKey(claims.UserID)).Result()
		if err != nil || cached != tokenStr {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "token已失效，请重新登录"})
			c.Abort()
			return
		}
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("realName", claims.RealName)
		c.Set("role", claims.Role)
		// 每次鉴权请求（含页面刷新后的请求）都触发流程缓存预热。
		services.WarmActiveOrchidWorkflowCacheIfDue(0)
		c.Next()
	}
}
