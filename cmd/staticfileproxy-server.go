package main

import (
	"encoding/json"
	"fmt"
	m "github.com/boyosoft/staticfileproxy"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type ConfigStruct struct {
	GinMode           string `json:"gin_mode"`
	Host              string `json:"host"`
	Port              int    `json:"port"`
	APIPrefix         string `json:"apiprefix"`
	StaticFilesFolder string `json:"static_files"`
	RedisHost         string `json:"redis_host"`
}

var (
	Config ConfigStruct
	logger = log.New(os.Stdout, "[APP API]:", log.LstdFlags)
)

func parseJsonFile(path string, v interface{}) {
	file, err := os.Open(path)
	if err != nil {
		logger.Fatal("配置文件读取失败:", err)
	}
	defer file.Close()
	dec := json.NewDecoder(file)
	err = dec.Decode(v)
	if err != nil {
		logger.Fatal("配置文件解析失败:", err)
	}
}

func getDefaultCode(path string) (code template.HTML) {
	if path != "" {
		content, err := ioutil.ReadFile(path)
		if err != nil {
			logger.Fatal("文件 " + path + " 没有找到")
		}
		code = template.HTML(string(content))
	}
	return
}

func loadStaticFiles(root string) (staticFiles []string, err error) {
	staticFiles = make([]string, 0, 1000)
	// Read content into docs field.
	extMap := map[string]bool{
		".rar":    true,
		".zip":    true,
		".tar.gz": true,
		".tar":    true}
	fn := func(p string, info os.FileInfo, err error) error {
		if _, found := extMap[filepath.Ext(p)]; !found {
			return nil
		}
		f, err := os.Open(p)
		if err != nil {
			return err
		}
		defer f.Close()
		staticFiles = append(staticFiles, p)
		return nil
	}
	err = filepath.Walk(root, fn)
	if err != nil {
		return
	}
	sort.Strings(staticFiles)
	return
}

func main() {

	parseJsonFile("config/config.json", &Config)

	gin.SetMode(Config.GinMode)

	r := gin.Default()

	// r.GET()
	type FOO struct {
		APIPrefix string
		Filename  string
	}
	assetNames := func() (rlt []string) {
		rawAssetNames := m.AssetNames()
		rlt = make([]string, 0, len(rawAssetNames))
		for _, filename := range rawAssetNames {
			if strings.HasPrefix(filename, "W2016") || strings.HasSuffix(filename, "png") || strings.HasSuffix(filename, ".ico") || strings.HasSuffix(filename, ".rptconfig") || strings.HasSuffix(filename, ".rptdesign") {
				rlt = append(rlt, filename)
			}
		}
		sort.Strings(rlt)
		return
	}()
	foos := make([]*FOO, 0, len(assetNames))
	for _, filename := range assetNames {
		foo := &FOO{APIPrefix: Config.APIPrefix,
			Filename: filename}
		foos = append(foos, foo)
	}

	indextmpl := template.New("index.tmpl")
	indextmpl.Parse(string(m.MustAsset("index.tmpl")))
	//
	r.HTMLRender = &render.HTMLProduction{Template: indextmpl}
	r.SetHTMLTemplate(indextmpl)

	/////////////////////////////////////////
	// r.LoadHTMLGlob("data/*.tmpl")
	// r.StaticFile(fmt.Sprintf("%s/", Config.APIPrefix), )
	// r.Static("/css", "templates/css")
	// r.Static("/js", "templates/js")
	// r.Static("/fonts", "templates/fonts")
	getRelativePath := func(staticFilePath string) string {
		return "downloads" + "/" + filepath.Base(staticFilePath)
	}
	staticFiles := func() (staticFiles []string) {

		if staticFilePaths, err := loadStaticFiles(Config.StaticFilesFolder); err == nil {
			staticFiles = make([]string, 0, len(staticFilePaths))
			for _, staticFilePath := range staticFilePaths {
				log.Println(staticFilePath)
				r.StaticFile(getRelativePath(staticFilePath), staticFilePath)
				staticFiles = append(staticFiles, filepath.Base(staticFilePath))
			}

		}
		return
	}()
	/////////////////////////////////////////
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.GET("/", func(c *gin.Context) {
		type BS struct {
			BindataFiles []*FOO
			StaticFiles  []string
		}
		bs := &BS{foos,
			staticFiles}
		c.HTML(200, "index.tmpl", bs)
	})
	r.GET(fmt.Sprintf("%s/:filename", Config.APIPrefix), func(c *gin.Context) {
		filename := c.Param("filename")
		data := m.MustAsset(filename)
		if strings.HasSuffix(filename, ".xml") {
			c.Data(200, "text/xml", data)
		} else if strings.HasSuffix(filename, ".png") {
			c.Data(200, "image/png", data)
		} else if strings.HasSuffix(filename, ".rptdesign") || strings.HasSuffix(filename, ".rptconfig") {
			c.Data(200, "application-xdownload", data)
			c.Header("Content-Disposition", fmt.Sprintf("attachment;filename=%s", filename))
		}
	})
	r.Run(fmt.Sprintf("%s:%d", Config.Host, Config.Port))
}
