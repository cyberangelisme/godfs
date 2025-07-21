package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sjqzhang/go-fastdfs/middleware"
)

var mux *http.ServeMux

func (c *Server) initRouter() {
	groupRoute := ""
	if Config().SupportGroupManage && Config().Group != "" {
		groupRoute = "/" + Config().Group
	}
	uploadPage := "upload.html"
	if groupRoute == "" {
		mux.HandleFunc(fmt.Sprintf("%s", "/"), c.Download)
		mux.HandleFunc(fmt.Sprintf("/%s", uploadPage), c.Index)
	} else {
		mux.HandleFunc(fmt.Sprintf("%s", "/"), c.Download)
		mux.HandleFunc(fmt.Sprintf("%s", groupRoute), c.Download)
		mux.HandleFunc(fmt.Sprintf("%s/%s", groupRoute, uploadPage), c.Index)
	}
	mux.HandleFunc(fmt.Sprintf("%s/check_files_exist", groupRoute), c.CheckFilesExist)
	mux.HandleFunc(fmt.Sprintf("%s/check_file_exist", groupRoute), c.CheckFileExist) // 获取文件是否存在，请求中包含MD5
	mux.HandleFunc(fmt.Sprintf("%s/upload", groupRoute), c.Upload)
	mux.HandleFunc(fmt.Sprintf("%s/delete", groupRoute), c.RemoveFile)
	mux.HandleFunc(fmt.Sprintf("%s/get_file_info", groupRoute), c.GetFileInfo) // 获取文件信息，利用path/MD5 转化为MD5到leveldb中查询
	mux.HandleFunc(fmt.Sprintf("%s/sync", groupRoute), c.Sync)                 // 手动同步文件，低峰时操作
	mux.HandleFunc(fmt.Sprintf("%s/stat", groupRoute), c.Stat)                 // 获取系统每日统计信息，包括文件size、数量等
	mux.HandleFunc(fmt.Sprintf("%s/repair_stat", groupRoute), c.RepairStatWeb) // 修复统计数据
	mux.HandleFunc(fmt.Sprintf("%s/status", groupRoute), c.Status)             // 查看文件系统状态
	mux.HandleFunc(fmt.Sprintf("%s/repair", groupRoute), c.Repair)             // 修复同步本地缺失文件，
	mux.HandleFunc(fmt.Sprintf("%s/report", groupRoute), c.Report)             // 请求返回report html
	mux.HandleFunc(fmt.Sprintf("%s/backup", groupRoute), c.BackUp)             // 备份元数据
	mux.HandleFunc(fmt.Sprintf("%s/search", groupRoute), c.Search)             // 搜索文件，返回匹配的fileInfos
	mux.HandleFunc(fmt.Sprintf("%s/list_dir", groupRoute), c.ListDir)          // 获取文件列表
	mux.HandleFunc(fmt.Sprintf("%s/remove_empty_dir", groupRoute), c.RemoveEmptyDir)
	mux.HandleFunc(fmt.Sprintf("%s/repair_fileinfo", groupRoute), c.RepairFileInfo) // 从文件目录中修复元数据，需要开启搬迁功能，修改cfg.json配置文件中的 enable_migrate 设为true
	mux.HandleFunc(fmt.Sprintf("%s/reload", groupRoute), c.Reload)                  //  重新加载配置文件
	mux.HandleFunc(fmt.Sprintf("%s/syncfile_info", groupRoute), c.SyncFileInfo)
	mux.HandleFunc(fmt.Sprintf("%s/get_md5s_by_date", groupRoute), c.GetMd5sForWeb)
	mux.HandleFunc(fmt.Sprintf("%s/receive_md5s", groupRoute), c.ReceiveMd5s)
	mux.HandleFunc(fmt.Sprintf("%s/gen_google_secret", groupRoute), c.GenGoogleSecret)
	mux.HandleFunc(fmt.Sprintf("%s/gen_google_code", groupRoute), c.GenGoogleCode)
	mux.Handle(fmt.Sprintf("%s/static/", groupRoute), http.StripPrefix(fmt.Sprintf("%s/static/", groupRoute), http.FileServer(http.Dir("./static"))))
	mux.HandleFunc("/"+Config().Group+"/", c.Download)
}
func (c *Server) initRouterByGin(r *gin.Engine) {
	groupRoute := ""
	if Config().SupportGroupManage && Config().Group != "" {
		groupRoute = "/" + Config().Group
	}

	uploadPage := "upload.html"

	// 全局中间件
	r.Use(middleware.LoggerToFile())

	// 不需要 JWT 认证的路由
	if groupRoute == "" {
		r.GET("/", Adapt(c.Download))
		r.GET("/"+uploadPage, Adapt(c.Index))
	} else {
		r.GET("/", Adapt(c.Download))
		r.GET(groupRoute, Adapt(c.Download))
		r.GET(groupRoute+"/"+uploadPage, Adapt(c.Index))
	}
	// 需要 JWT 认证的路由
	authGroup := r.Group(groupRoute)
	authGroup.Use(middleware.JWTAuth()) // 添加 JWT 中间件
	{
		authGroup.POST("/check_files_exist", Adapt(c.CheckFilesExist))
		authGroup.POST("/check_file_exist", Adapt(c.CheckFileExist))
		authGroup.POST("/upload", Adapt(c.Upload))
		authGroup.POST("/delete", Adapt(c.RemoveFile))
		authGroup.POST("/get_file_info", Adapt(c.GetFileInfo))
		authGroup.POST("/sync", Adapt(c.Sync))
		authGroup.GET("/stat", Adapt(c.Stat))
		authGroup.GET("/repair_stat", Adapt(c.RepairStatWeb))
		authGroup.GET("/status", Adapt(c.Status))
		authGroup.POST("/repair", Adapt(c.Repair))
		authGroup.GET("/report", Adapt(c.Report))
		authGroup.GET("/backup", Adapt(c.BackUp))
		authGroup.POST("/search", Adapt(c.Search))
		authGroup.POST("/list_dir", Adapt(c.ListDir))
		authGroup.POST("/remove_empty_dir", Adapt(c.RemoveEmptyDir))
		authGroup.POST("/repair_fileinfo", Adapt(c.RepairFileInfo))
		authGroup.GET("/reload", Adapt(c.Reload))
		authGroup.POST("/syncfile_info", Adapt(c.SyncFileInfo))
		authGroup.GET("/get_md5s_by_date", Adapt(c.GetMd5sForWeb))
		authGroup.POST("/receive_md5s", Adapt(c.ReceiveMd5s))
		authGroup.GET("/gen_google_secret", Adapt(c.GenGoogleSecret))
		authGroup.GET("/gen_google_code", Adapt(c.GenGoogleCode))
	}

	// 静态文件服务
	staticHandler := http.StripPrefix(groupRoute+"/static/", http.FileServer(http.Dir("./static")))
	r.GET(groupRoute+"/static/*filepath", Adapt(staticHandler.ServeHTTP))
	// 兜底下载
	r.Any("/"+Config().Group+"/*path", Adapt(c.Download))
}

// net/http 原始代码适配到 gin 中
func Adapt(fn http.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		fn(c.Writer, c.Request)
	}
}
