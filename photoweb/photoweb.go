/*
1.http.Request.FormFile(string) (multipart.File,*multipart.FileHeader,error): 上传文件.
type File interface {
    io.Reader
    io.ReaderAt
    io.Seeker
    io.Closer
}
type FileHeader struct {
    Filename    string
    Header      textproto.MIMEHeader
    Size        int64
    content     []byte
    tmpfile     string
}
2.os.Create(string): 创建(临时)文件。
3.io.Copy(dst Writer, src Reader)(written int64, err error): 将src的内容拷贝到dst中，并返回拷贝的字节数。
4.http.Redirect(ResponseWriter, *Request, url string, code int)：重定向，根据url答复请求。即根据条件跳转页面。
5.http.Request.FormValue(key string)string: 查询并返回表单中第一个名字为key的内容。
6.http.ResponseWriter.Header().Set(key, value string): 将value设置为key的内容，即key：value。
7.http.ServeFile(ResponseWriter, *Request, name string): 将路径下的文件从磁盘中读取并作为服务端的返回信息输出给客户端。
8.os.Stat(path string)(FileInfo, error): 获取该路径下文件的信息。
type FileInfo interface {
    Name() string
    Size() int64
    Mode() FileMode
    ModTime() time.Time
    IsDir() bool
	Sys() interface{}
9.os.IsExist(error)bool：文件是否已存在。
10.ioutil.ReadDir(dirname string) ([]os.FileInfo, error)：读取目录并返回排好序的文件以及子目录名。
11.template.ParseFiles(filenames ...string) (*Template, error): 读取并解析HTML文件，生成HTML模板，返回一个*template.Template值。
注：（1）返回的模板==第一个文件的名称和内容；
    （2）当同时解析多个在不同目录中具有相同的filename的文件时，显示结果为最后一个文件。
12.template.Template.Execute(wr io.Writer, data interface{}) error：Execute根据模板语法来执行模板的渲染，并将渲染结果作为HTTP的返回数据输出到网页。
13.path.Ext(path string) string: 获取路径文件的扩展名（例：.exe、.html）。
14.template.Must(t *Template, err error) *Template：检测这个模板，如果有错，直接崩溃。你可以找到问题并解决，而不需要再判断是否有错误，再进行提示操作等。
15.ioutil.TempFile(dir, pattern string) (f *os.File, err error):TempFile在目录dir中创建一个新的临时文件pattern，打开该文件进行读写，然后返回生成的os.File。
16.http.ServeMux：默认路由表结构体。
*/
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
		/*1.0版本
		io.WriteString(w, "<form method=\"POST\" action=\"/upload\" enctype=\"multipart/form-data\">"+
		"Choose an image to upload: <input name=\"image\" type=\"file\" />"+
		"<input type=\"submit\" value=\"Upload\" />"+
		"</form>")
		*/
		/*2.0版本
		t, err := template.ParseFiles(TemplateDIR + "/upload.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		t.Execute(w, nil)
		*/
		/*3.0版本
		if err := renderHTML(w, TemplateDIR+"/upload.html", nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		*/
		return
	}
	//上传文件，并接收文件
	if r.Method == "POST" {
		f, h, err := r.FormFile("image")
		filename := filepath.Base(h.Filename)
		/*if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}*/
		check(err)
		defer f.Close()

		/*t, err := os.Create(UploadDIR + "/" + filename)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		*/
		t, err := ioutil.TempFile(UploadDIR, filename)
		check(err)
		fmt.Println(filename + "已上传！")
		defer t.Close()

		/*if _, err := io.Copy(t, f); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		*/
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
	/*if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	*/
	check(err)
	locals := make(map[string]interface{})
	images := []string{}
	for _, fileInfo := range fileInfoArr {
		images = append(images, fileInfo.Name())
	}
	locals["images"] = images

	/*1.0版本
	var listHtml string
	for _, fileInfo := range fileInfoArr {
		imgid := fileInfo.Name()
		listHtml += "<li><a href=\"/views?id="+imgid+"\">imgid</a></li>"
	}
	io.WriteString(w, "<ol>"+listHtml+"</ol>")
	*/
	/*2.0版本
	t, err := template.ParseFiles(TemplateDIR + "/list.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(w, nil)
	*/
	/*3.0版本
	if err := renderHTML(w, TemplateDIR+"/list.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	*/
	renderHTML(w, TemplateDIR+"/list.html", locals)
}

//模板渲染函数
func renderHTML(w http.ResponseWriter, tmpl string, locals map[string]interface{}) (err error) {
	/*t, err := template.ParseFiles(tmpl)
	if err != nil {
		return
	}
	err = t.Execute(w, locals)
	*/
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
