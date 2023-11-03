package ws

//插入mongoDB的数据类型
// BSON（Binary JSON）是对JSON的一种二进制编码数据格式，
// bson对象是键值对对象，bson是JSON的二进制格式
// 主要用于MongoDB中。
type Traniner struct {
	Content   string `bson:"content"`   //内容
	StartTime int64  `bson:"startTime"` //创建时间
	EndTime   int64  `bson:"endTime"`   //过期时间
	Read      uint   `bson:"read"`      //已读
}

type Result struct {
	StartTime int64
	Msg       string
	Content   interface{}
	From      string
}
