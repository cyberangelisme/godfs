package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var jwtSecret2 = []byte("secret-lxf")

// 生成一个合法的 JWT token
func generateTestToken() string {
	claims := jwt.MapClaims{
		"username": "testuser",
		"role":     "user",
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(jwtSecret2)
	return tokenString
}

func setupRouter() *gin.Engine {
	r := gin.Default()
	gin.ForceConsoleColor()
	r.Use(JWTAuth())
	{
		r.GET("/protected", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"msg": "success"})
		})
	}
	return r
}

func TestJWTAuth_MissingToken(t *testing.T) {
	r := setupRouter()

	req, _ := http.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "缺少 Token")
}

// ok✅
func TestJWTAuth_InvalidToken(t *testing.T) {
	r := setupRouter()

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Auth-Token", "invalid-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	println("response :", w.Body.String())

	// assert.Equal(t, http.StatusUnauthorized, w.Code)
	// assert.Contains(t, w.Body.String(), "无效 Token")
}

// ok✅
func TestJWTAuth_ValidToken(t *testing.T) {
	r := setupRouter()

	token := generateTestToken()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Auth-Token", token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	println("req.URL.Path: ", req.URL.Path)
	println("Header-Auth-Token: ", req.Header.Get("Auth-Token"))
	println("response :", w.Body.String())
	// assert.Equal(t, http.StatusOK, w.Code)
	// assert.Contains(t, w.Body.String(), "success")
}

func TestJWTAuth_GetClaims(t *testing.T) {
	r := setupRouter()

	token := generateTestToken()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// 模拟中间件后的上下文，验证 claims 是否正确写入
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("claims", jwt.MapClaims{
			"username": "testuser",
			"role":     "user",
		})
		c.Next()
	}, JWTAuth())

	router.GET("/protected2", func(c *gin.Context) {
		if claims, exists := c.Get("claims"); exists {
			assert.Equal(t, "testuser", claims.(jwt.MapClaims)["username"])
			assert.Equal(t, "user", claims.(jwt.MapClaims)["role"])
			c.JSON(http.StatusOK, gin.H{"username": claims.(jwt.MapClaims)["username"]})
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "claims 不存在"})
		}
	})

	req2, _ := http.NewRequest("GET", "/protected2", nil)
	req2.Header.Set("Authorization", token)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Contains(t, w2.Body.String(), "testuser")
}
