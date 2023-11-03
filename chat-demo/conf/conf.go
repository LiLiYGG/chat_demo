package conf

import (
	"chat-demo/model"
	"context"
	"fmt"
	"strings"

	logging "github.com/sirupsen/logrus" //github.com/sirupsen/logrus
	"go.mongodb.org/mongo-driver/mongo"  //MongoDB的Go驱动包
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/ini.v1"
)

// 读取配置信息
var (
	MongoDBClient *mongo.Client
	AppMode       string
	HttpPort      string
	Db            string
	DbHost        string
	DbPort        string
	DbUser        string
	DbPassWord    string
	DbName        string

	MongoDBName string
	MongoDBAddr string
	MongoDBPwd  string
	MongoDBPort string
)

func Init() {
	//从本地读取环境变量
	file, err := ini.Load("E:\\chat-demo\\conf\\conf.ini")
	if err != nil {
		fmt.Println("配置文件读取错误，请检查文件路径:", err)
	}
	LoadServer(file)
	LoadMysqlData(file)
	LoadMongoDB(file)

	//MySQL
	path := strings.Join([]string{DbUser, ":", DbPassWord, "@tcp(", DbHost, ":", DbPort, ")/", DbName, "?charset=utf8&parseTime=true"}, "")
	model.Database(path)
	//MongoDB
	MongoDB()
}

func MongoDB() {
	// 设置mongoDB客户端连接信息
	// mongo.Connect() 函数创建了一个客户端实例，
	// 并向它传递了一个Gocontext 。
	// 当向数据库发出请求时，
	// 使用context.TODO() ，
	// 创建一个服务器会接受的上下文。
	clientOptions := options.Client().ApplyURI("mongodb://" + MongoDBAddr + ":" + MongoDBPort)
	var err error
	MongoDBClient, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		logging.Info(err)
	}
	// Ping() 判断是否已经找到并连接了MongoDB数据库。
	err = MongoDBClient.Ping(context.TODO(), nil)
	if err != nil {
		logging.Info(err)
	}
	logging.Info("MongoDB Connect")
}

// service
func LoadServer(file *ini.File) {
	AppMode = file.Section("service").Key("AppMode").String()
	HttpPort = file.Section("service").Key("HttpPort").String()
}

// mysql
func LoadMysqlData(file *ini.File) {
	Db = file.Section("mysql").Key("Db").String()
	DbHost = file.Section("mysql").Key("DbHost").String()
	DbPort = file.Section("mysql").Key("DbPort").String()
	DbUser = file.Section("mysql").Key("DbUser").String()
	DbPassWord = file.Section("mysql").Key("DbPassWord").String()
	DbName = file.Section("mysql").Key("DbName").String()
}

// MongoDB
func LoadMongoDB(file *ini.File) {
	MongoDBName = file.Section("MongoDB").Key("MongoDBName").String()
	MongoDBAddr = file.Section("MongoDB").Key("MongoDBAddr").String()
	MongoDBPwd = file.Section("MongoDB").Key("MongoDBPwd").String()
	MongoDBPort = file.Section("MongoDB").Key("MongoDBPort").String()
}
