package epub

import (
	"context"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nexptr/omnigram-server/epub/schema"
	"github.com/nexptr/omnigram-server/log"
	"github.com/nexptr/omnigram-server/utils"
)

func coverImageHandle(c *gin.Context) {

	coverPath := strings.TrimPrefix(c.Param(`book_cover_path`), `/`)

	ext := filepath.Ext(c.Param(`book_cover_path`))

	if ext != ".png" && ext != ".jpeg" && ext != ".jpg" {
		log.E(`图片路径ID为空：`, ext)
		c.JSON(200, utils.ErrReqArgs)
		return
	}

	log.I(`获取图片内容`, coverPath)

	obj, err := kv.GetObject(context.TODO(), coverPath[0:2], coverPath)

	if err != nil {
		log.E(`获取图片内容失败`, err.Error())
		c.JSON(http.StatusNotFound, utils.ErrNoFound)
		return
	}

	c.Data(200, "image/"+ext, obj.Data)

}

func BookDetail(c *gin.Context) {
	panic(`TODO`)
}

// BookUpload 上传文件
func bookUploadHandle(c *gin.Context) {
	//处理上传文件并存储到数据库

	file, err := c.FormFile("file")
	if err != nil {
		log.E(`上传文件失败：`, err)
		c.JSON(200, utils.ErrReqArgs)
		return
	}

	log.I(`上传文件成功：`, file.Filename)

	uploadfile := filepath.Join(uploadPath, file.Filename)

	//存储文件到上传目录
	if err := c.SaveUploadedFile(file, uploadfile); err != nil {
		log.E(`上传文件失败：`, err)
		c.JSON(http.StatusOK, utils.ErrSaveFile)
		return
	}

	//尝试解析文件
	book := &schema.Book{Path: uploadfile}

	if err := book.GetMetadataFromFile(); err != nil {
		log.E(`解析文件失败：`, err)
		c.JSON(http.StatusOK, utils.ErrParseEpubFile.WithMessage(err.Error()))
		return
	}

	if err := book.Save(context.Background(), store.DB, kv); err != nil {
		log.E(`录入文档失败`, err)
		c.JSON(http.StatusOK, utils.ErrSaveFile)
		return
	}

	c.JSON(http.StatusOK, utils.SUCCESS)
}

// BookDownload 下载图书
// /books/:book_id/download
func bookDownloadHandle(c *gin.Context) {
	dentifier := c.Param(`book_id`)

	if dentifier == "" {
		log.E(`图书ID为空`)
		c.JSON(200, utils.ErrReqArgs)
		return
	}

	book, err := schema.FirstBookByIdentifier(store.DB, dentifier)

	if err != nil {
		log.E(`获取图书失败：`, err)
		c.JSON(200, utils.ErrNoFound)
		return
	}

	//读取书籍文件路径到io

	c.Header(`Content-Type`, `application/octet-stream`)
	c.Header("Content-Disposition", "attachment; filename="+book.Title+".epub")
	c.Header("Content-Transfer-Encoding", "binary")
	c.File(book.Path)

}

func RecentBook(c *gin.Context) {
	req := &struct {
		Recent int `json:"recent" binding:"required,gte=0"`
	}{
		Recent: 12,
	}

	if err := c.ShouldBind(req); err != nil {
		log.I(`用户登录参数异常`, err)
		c.JSON(200, utils.ErrReqArgs.WithMessage(err.Error()))
		return
	}
	recentBooks, err := schema.RecentBooks(store.DB, req.Recent)

	if err != nil {
		log.I(`用户登录参数异常`, err)
		c.JSON(200, utils.ErrInnerServer.WithMessage(err.Error()))
		return
	}

	c.JSON(200, utils.SUCCESS.WithData(recentBooks))

}

// SearchBook 模糊搜索
func SearchBook(c *gin.Context) {

	req := &utils.Query{}

	if err := c.ShouldBind(req); err != nil {
		log.I(`用户登录参数异常`, err)
		c.JSON(200, utils.ErrReqArgs.WithMessage(err.Error()))
		return
	}

	//TODO过滤特殊字符

	recentBooks, err := schema.SearchBooks(store.DB, req)

	if err != nil {
		log.I(`用户登录参数异常`, err)
		c.JSON(200, utils.ErrInnerServer.WithMessage(err.Error()))
		return
	}

	c.JSON(200, utils.SUCCESS.WithData(recentBooks))

}

// Index 返回 首页 随机ID书籍和最近添加到书籍集。
func Index(c *gin.Context) {

	req := &struct {
		Random int `json:"random" binding:"required,gte=0,lt=30"`
		Recent int `json:"recent" binding:"required,gte=0,lt=30"`
	}{
		Random: 10,
		Recent: 12,
	}

	if err := c.ShouldBind(req); err != nil {
		log.I(`用户登录参数异常`, err)
		c.JSON(200, utils.ErrReqArgs.WithMessage(err.Error()))
		return
	}

	randBooks, err := schema.RandomBooks(store.DB, req.Random)

	if err != nil {
		log.I(`用户登录参数异常`, err)
		c.JSON(200, utils.ErrInnerServer.WithMessage(err.Error()))
		return
	}

	recentBooks, err := schema.RecentBooks(store.DB, req.Recent)

	if err != nil {
		log.I(`用户登录参数异常`, err)
		c.JSON(200, utils.ErrInnerServer.WithMessage(err.Error()))
		return
	}

	data := map[string]interface{}{
		"random": randBooks.Books,
		"recent": recentBooks.Books,
	}

	c.JSON(200, utils.SUCCESS.WithData(data))

}

// UserInfo GET /api/user/info
// 获取用户信息
func GetBookStats(c *gin.Context) {
	log.D(`获取书籍概览信息`)

	stats, err := schema.GetBookStats(store.DB)

	if err != nil {
		log.E(`获取数据信息失败`)
		c.JSON(404, utils.ErrNoFound)
	}

	c.JSON(200, utils.SUCCESS.WithData(stats))

}
