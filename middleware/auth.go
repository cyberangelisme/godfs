// v2.0 把jwt 改成中间件形式
package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v4"
)

var jwtSecret = []byte("secret-lxf")

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Auth-Token")
		if authHeader == "" {
			c.AbortWithStatusJSON(401, gin.H{"code": 401, "msg": "Missing token"})
			return
		}
		// // Bearer <token> 去除Bearer字符
		// parts := strings.SplitN(authHeader, " ", 2)
		// if len(parts) != 2 || parts[0] != "Bearer" {
		// 	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		// 		"code": http.StatusUnauthorized,
		// 		"msg":  "Token 格式错误",
		// 	})
		// 	return
		// }

		// tokenString := parts[1]

		tokenString := authHeader

		// 解析并验证 JWT
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// 验证签名方法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecret, nil
		})
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "Invalid token",
			})
		}

		// 验证 token 是否有效
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// 可选：检查过期时间
			if exp, ok := claims["exp"].(float64); ok {
				if time.Unix(int64(exp), 0).Before(time.Now()) {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
						"code": http.StatusUnauthorized,
						"msg":  "Token 已过期",
					})
					return
				}
			}

			// 将用户信息写入上下文，供后续处理使用
			c.Set("claims", claims)
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": http.StatusUnauthorized,
				"msg":  "Token 解析失败",
			})
			return
		}
		c.Next()
	}
}
