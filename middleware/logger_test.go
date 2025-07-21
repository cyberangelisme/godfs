package middleware

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sjqzhang/go-fastdfs/pkg/logger"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func initViper() {
	configName := "config"
	configPath := "../config"
	viper.SetConfigName(configName) // 设置配置文件名（不带后缀）
	viper.SetConfigType("yaml")     // 设置配置文件类型（如 yaml/json/toml）
	viper.AddConfigPath(configPath) // 添加配置文件所在目录
	viper.AddConfigPath(".")        // 可以添加多个路径

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}

	fmt.Println("使用配置文件:", viper.ConfigFileUsed())
	logger.Init()
}

func TestLoggerInit(t *testing.T) {
	// 设置默认配置
	viper.Set("settings.log.level", "debug")
	viper.Set("settings.log.path", "./test.log")
	viper.Set("settings.log.maxsize", 10)
	viper.Set("settings.log.maxBackups", 5)
	viper.Set("settings.log.maxAge", 7)
	viper.Set("settings.log.localtime", true)
	viper.Set("settings.log.compress", false)
	viper.Set("settings.log.consoleStdout", true)
	viper.Set("settings.log.fileStdout", true)

	// 调用 Init
	logger.Init()

	// 测试日志是否能正常输出
	logger.Info("This is a test log message")
}

// ok👌
func TestLoggerToFile(t *testing.T) {
	initViper()

	// 设置为测试模式，避免输出日志到控制台（如果 logger 有设置）
	//gin.SetMode(gin.TestMode)

	// 创建一个模拟的 Gin 路由
	r := gin.New()
	r.Use(LoggerToFile())

	// 注册一个测试路由
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	// 创建一个测试请求
	req, err := http.NewRequest("GET", "/test", nil)
	assert.NoError(t, err)

	// 设置客户端 IP（模拟真实 IP）
	req.RemoteAddr = "192.168.1.1:12345"

	// 记录响应
	w := httptest.NewRecorder()

	// 执行请求，不用先run再测试了。
	r.ServeHTTP(w, req)

	// 断言响应状态码是否为 200
	assert.Equal(t, http.StatusOK, w.Code)
}
