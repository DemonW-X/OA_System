package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("oa-system-secret-key")

type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	RealName string `json:"real_name"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateToken(userID int, username, realName, role string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		RealName: realName,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

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

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未登录"})
			c.Abort()
			return
		}
		claims, err := ParseToken(strings.TrimPrefix(authHeader, "Bearer "))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "token无效或已过期"})
			c.Abort()
			return
		}
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("realName", claims.RealName)
		c.Set("role", claims.Role)
		c.Next()
	}
}
