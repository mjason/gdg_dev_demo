package main

import (
	"flute"
	"flute/mongo"
	"fmt"
	"io"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
	"os"
)

var C map[string]*mgo.Collection

func main() {
	// 初始化日志
	err := flute.InitAppLogConfig("dev.log")
	defer flute.CloseLogFile()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// 初始化数据库
	err = mongo.Init("127.0.0.1")
	defer mongo.Session.Close()
	if err != nil {
		flute.Logger.Fatalln(err.Error())
	}

	// 初始化我们需要的集合
	C = make(map[string]*mgo.Collection)
	C["image"] = mongo.Session.DB("gdg").C("images")

	// 初始化路由
	flute.Resources("api/images", &ImageController{})

	// 静态文件路由
	flute.Flute.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))
	// 启动服务器
	flute.Start(":3001")
}

type ImageController struct {
	flute.Controller
}

func (self *ImageController) Index() {
	var image []*Image
	err := C["image"].Find(bson.M{}).All(&image)
	if err != nil {
		flute.Logger.Println(err.Error())
		self.RenderJson(500, map[string]string{"error": "查询时发生错误"})
		return
	}
	self.RenderJson(200, image)
}

func (self *ImageController) Create() {
	imageFile, handler, err := self.ParamsFile("image")
	if err != nil {
		flute.Logger.Println(err.Error())
		self.RenderJson(403, map[string]string{"error": "没有提交图片"})
		return
	}
	defer imageFile.Close()
	id := bson.NewObjectId()
	imagePath := "./public/image/" + id.Hex() + handler.Filename
	f, err := os.OpenFile(imagePath, os.O_WRONLY|os.O_CREATE, 0666)
	defer f.Close()
	if err != nil {
		flute.Logger.Fatalln(err.Error())
		self.RenderJson(500, map[string]string{"error": "服务发生错误"})
		return
	}
	io.Copy(f, imageFile)
	image := &Image{
		id,
		"/image/" + id.Hex() + handler.Filename,
		self.ParamsFrom("explanation"),
		self.ParamsFrom("tag"),
	}
	err = image.Save()
	if err != nil {
		flute.Logger.Fatalln(err.Error())
		self.RenderJson(500, map[string]string{"error": "服务发生错误"})
		return
	}
	self.RenderJson(200, image)
}

type Image struct {
	Id               bson.ObjectId `bson:"_id" json:"id"`
	ImagePath        string        `json:"image_path"`
	imageExplanation string        `json:"image_explanation"`
	Tag              string        `json:"tag"`
}

func (self *Image) Save() error {
	return C["image"].Insert(self)
}
