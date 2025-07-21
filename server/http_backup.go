package server

// 备份接口，扩展功能
import (
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/astaxie/beego/httplib"
	log "github.com/sjqzhang/seelog"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// 创建备份文件：为指定日期目录下的元数据创建两个备份文件，分别是 meta.data 和一个日志文件
/*
创建备份文件：生成 meta.data 和日志文件。
文件路径：/data/backup/2023-10-05/meta.data 和 /data/backup/2023-10-05/fileMd5.log
遍历数据库条目，反序列化 fileInfo 数据。
写入 meta.data 文件：
第一条记录：d41d8cd98f00b204e9800998ecf8427e {"Name":"example.txt", ...}
第二条记录（MD5 based on path + name）：md5_of_/data/files/example.txt {"Name":"example.txt", ...}
写入日志文件：
记录内容：d41d8cd98f00b204e9800998ecf8427e|1024|1633423423|/data/files/example.txt
*/
func (c *Server) BackUpMetaDataByDate(date string) {
	defer func() {
		if re := recover(); re != nil {
			buffer := debug.Stack()
			log.Error("BackUpMetaDataByDate")
			log.Error(re)
			log.Error(string(buffer))
		}
	}()
	var (
		err          error
		keyPrefix    string
		msg          string
		name         string
		fileInfo     FileInfo
		logFileName  string
		fileLog      *os.File
		fileMeta     *os.File
		metaFileName string
		fi           os.FileInfo
	)
	logFileName = DATA_DIR + "/" + date + "/" + CONST_FILE_Md5_FILE_NAME
	c.lockMap.LockKey(logFileName)
	defer c.lockMap.UnLockKey(logFileName)
	metaFileName = DATA_DIR + "/" + date + "/" + "meta.data"
	os.MkdirAll(DATA_DIR+"/"+date, 0775)
	if c.util.IsExist(logFileName) {
		os.Remove(logFileName)
	}
	if c.util.IsExist(metaFileName) {
		os.Remove(metaFileName)
	}
	fileLog, err = os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)
	if err != nil {
		log.Error(err)
		return
	}
	defer fileLog.Close()
	fileMeta, err = os.OpenFile(metaFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)
	if err != nil {
		log.Error(err)
		return
	}
	defer fileMeta.Close()
	keyPrefix = "%s_%s_"
	keyPrefix = fmt.Sprintf(keyPrefix, date, CONST_FILE_Md5_FILE_NAME)
	iter := server.logDB.NewIterator(util.BytesPrefix([]byte(keyPrefix)), nil)
	defer iter.Release()
	for iter.Next() {
		if err = json.Unmarshal(iter.Value(), &fileInfo); err != nil {
			continue
		}
		name = fileInfo.Name
		if fileInfo.ReName != "" {
			name = fileInfo.ReName
		}
		msg = fmt.Sprintf("%s\t%s\n", fileInfo.Md5, string(iter.Value()))
		if _, err = fileMeta.WriteString(msg); err != nil {
			log.Error(err)
		}
		msg = fmt.Sprintf("%s\t%s\n", c.util.MD5(fileInfo.Path+"/"+name), string(iter.Value()))
		if _, err = fileMeta.WriteString(msg); err != nil {
			log.Error(err)
		}
		msg = fmt.Sprintf("%s|%d|%d|%s\n", fileInfo.Md5, fileInfo.Size, fileInfo.TimeStamp, fileInfo.Path+"/"+name)
		if _, err = fileLog.WriteString(msg); err != nil {
			log.Error(err)
		}
	}
	if fi, err = fileLog.Stat(); err != nil {
		log.Error(err)
	} else if fi.Size() == 0 {
		fileLog.Close()
		os.Remove(logFileName)
	}
	if fi, err = fileMeta.Stat(); err != nil {
		log.Error(err)
	} else if fi.Size() == 0 {
		fileMeta.Close()
		os.Remove(metaFileName)
	}
}

// 提供 HTTP 接口，对外暴露接口，支持本地和集群调用。
func (c *Server) BackUp(w http.ResponseWriter, r *http.Request) {
	var (
		err    error
		date   string
		result JsonResult
		inner  string
		url    string
	)
	result.Status = "ok"
	r.ParseForm()
	date = r.FormValue("date")
	inner = r.FormValue("inner")
	if date == "" {
		date = c.util.GetToDay()
	}
	if c.IsPeer(r) {
		// 集群内部调用
		if inner != "1" {
			for _, peer := range Config().Peers {
				backUp := func(peer string, date string) {
					url = fmt.Sprintf("%s%s", peer, c.getRequestURI("backup"))
					req := httplib.Post(url)
					req.Param("date", date)
					req.Param("inner", "1")
					req.SetTimeout(time.Second*5, time.Second*600)
					if _, err = req.String(); err != nil {
						log.Error(err)
					}
				}
				go backUp(peer, date)
			}
		}
		// 本地元数据备份
		go c.BackUpMetaDataByDate(date)
		result.Message = "back job start..."
		w.Write([]byte(c.util.JsonEncodePretty(result)))
	} else {
		result.Message = c.GetClusterNotPermitMessage(r)
		w.Write([]byte(c.util.JsonEncodePretty(result)))
	}
}
