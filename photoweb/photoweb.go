package main

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime/debug"
)

const (
	//UploadDIR 上传文件保存路径
	UploadDIR = "../src/photoweb/uploads"
	//TemplateDIR HTML模板文件保存路径
	TemplateDIR = "../src/photoweb/views"
	//ListDir 显示错误代码
	ListDir = 0x0001
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	//HTML上传表单,提供文件上传
	if r.Method == "GET" {
		renderHTML(w, TemplateDIR+"/upload.html", nil)
		return
	}
	//上传文件，并接收文件
	if r.Method == "POST" {
		f, h, err := r.FormFile("image")
		filename := filepath.Base(h.Filename)
		check(err)
		defer f.Close()

		t, err := ioutil.TempFile(UploadDIR, filename)
		check(err)
		fmt.Println(filename + "已上传！")
		defer t.Close()

		_, err = io.Copy(t, f)
		check(err)
		http.Redirect(w, r, "/view?id="+filename, http.StatusFound)
	}
}

//客户端显示图片
func viewHandler(w http.ResponseWriter, r *http.Request) {
	imageID := r.FormValue("id")           //接收id=
	imagePath := UploadDIR + "/" + imageID //组成图形文件存放路径
	if exists := isExists(imagePath); !exists {
		fmt.Println("图片不存在！")
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "image")
	http.ServeFile(w, r, imagePath)
}

//检查文件是否已存在
func isExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return os.IsExist(err)
}

//网页上列出文件
func listHandler(w http.ResponseWriter, r *http.Request) {
	fileInfoArr, err := ioutil.ReadDir(UploadDIR)

	check(err)
	locals := make(map[string]interface{})
	images := []string{}
	for _, fileInfo := range fileInfoArr {
		images = append(images, fileInfo.Name())
	}
	locals["images"] = images

	renderHTML(w, TemplateDIR+"/list.html", locals)
}

//模板渲染函数
func renderHTML(w http.ResponseWriter, tmpl string, locals map[string]interface{}) (err error) {
	err = templates[tmpl].Execute(w, locals)
	check(err)
	return
}

//模板缓存
var templates = make(map[string]*template.Template)

func init() {
	fileInfoArr, err := ioutil.ReadDir(TemplateDIR)
	if err != nil {
		panic(err)
	}

	var templateName, templatePath string
	for _, fileInfo := range fileInfoArr {
		templateName = fileInfo.Name()
		if ext := path.Ext(templateName); ext != ".html" {
			continue
		}
		templatePath = TemplateDIR + "/" + templateName
		log.Println("Loading template:", templatePath)
		t := template.Must(template.ParseFiles(templatePath))
		templates[templatePath] = t
	}
}

//捕捉http.Error()50x系列的服务端内部错误
func check(err error) {
	if err != nil {
		panic(err)
	}
}

//用闭包函数避免程序运行时出错崩溃
func safeHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if e, ok := recover().(error); ok {
				http.Error(w, e.Error(), http.StatusInternalServerError)
				log.Printf("\nWARN: panic in %v - %v", fn, e)
				log.Println(string(debug.Stack()))
			}
		}()
		fn(w, r)
	}
}

//动态请求和静态资源分离
func staticDirHandler(mux *http.ServeMux, prefix string, staticDir string, flags int) {
	mux.HandleFunc(prefix, func(w http.ResponseWriter, r *http.Request) {
		file := staticDir + r.URL.Path[len(prefix)-1:]
		if (flags & ListDir) == 0 {
			if exists := isExists(file); !exists {
				http.NotFound(w, r)
				return
			}
		}
		http.ServeFile(w, r, file)
	})
}

func main() {
	mux := http.NewServeMux()
	staticDirHandler(mux, "/assets/", "../src/photoweb/public", 0)
	http.HandleFunc("/list", safeHandler(listHandler))
	http.HandleFunc("/view", safeHandler(viewHandler))
	http.HandleFunc("/upload", safeHandler(uploadHandler))
	err := http.ListenAndServe(":8080", nil)
	fmt.Println("aa")
	if err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
	}
}
