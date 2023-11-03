package service

import (
	"chat-demo/conf"
	"chat-demo/model/ws"
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SendSortMsg struct {
	Content  string `json:"content"`
	Read     uint   `json:"read"`
	CreateAt int64  `json:"creat_at"`
}

// expire为过期时间
func InsertMsg(database, id string, content string, read uint, expire int64) error {
	// 插入到mongoDB中
	// 创建集合
	collection := conf.MongoDBClient.Database(database).Collection(id) //没有这个id集合会自动创建
	comment := ws.Traniner{
		Content:   content,
		StartTime: time.Now().Unix(),
		EndTime:   time.Now().Unix() + expire,
		Read:      read,
	}
	_, err := collection.InsertOne(context.TODO(), comment)
	return err
}
func FindMany(database, sendID, id string, time int64, pageSize int) (results []ws.Result, err error) {
	var resultMe []ws.Traniner  //id
	var resultYou []ws.Traniner //sendID
	sendIDCollection := conf.MongoDBClient.Database(database).Collection(sendID)
	idCollection := conf.MongoDBClient.Database(database).Collection(id)

	// options.Find().SetSort() 过滤器，对信息排序 时间排序，限制大小
	// bson.D是一种有序的键值对,每一对键值对都需要用大括号来连接，bson.D{{key,value},{key,value}...}
	sendIDTimeCursor, err := sendIDCollection.Find(context.TODO(),
		options.Find().SetSort(bson.D{{"startTime", -1}}),
		options.Find().SetLimit(int64(pageSize)))

	idTimeCursor, err := idCollection.Find(context.TODO(),
		options.Find().SetSort(bson.D{{"startTime", -1}}),
		options.Find().SetLimit(int64(pageSize)))
	// 如果不知道该使用什么context，可以通过context.TODO() 产生context
	err = sendIDTimeCursor.All(context.TODO(), &resultYou)
	err = idTimeCursor.All(context.TODO(), &resultMe)
	results, _ = AppendAndSort(resultMe, resultYou)
	return
}

func AppendAndSort(resultMe, resultYou []ws.Traniner) (results []ws.Result, err error) {
	for _, r := range resultMe {
		sendSort := SendSortMsg{ //构造返回的msg
			Content:  r.Content,
			Read:     r.Read,
			CreateAt: r.StartTime,
		}
		result := ws.Result{ //构造返回的所有内容，包括传送者
			StartTime: r.StartTime,
			Msg:       fmt.Sprintf("%v", sendSort),
			From:      "me",
		}
		results = append(results, result)
	}

	for _, r := range resultYou {
		sendSort := SendSortMsg{ //构造返回的msg
			Content:  r.Content,
			Read:     r.Read,
			CreateAt: r.StartTime,
		}
		result := ws.Result{ //构造返回的所有内容，包括构造者
			StartTime: r.StartTime,
			Msg:       fmt.Sprintf("%v", sendSort),
			From:      "you",
		}
		results = append(results, result)
	}
	return
}

func FirsFindtMsg(database string, sendId string, id string) (results []ws.Result, err error) {
	// 首次查询(把对方发来的所有未读都取出来)
	var resultsMe []ws.Traniner
	var resultsYou []ws.Traniner
	sendIdCollection := conf.MongoDBClient.Database(database).Collection(sendId)
	idCollection := conf.MongoDBClient.Database(database).Collection(sendId)
	filter := bson.M{"read": bson.M{
		"&all": []uint{0},
	}}
	sendIdCursor, err := sendIdCollection.Find(context.TODO(), filter, options.Find().SetSort(bson.D{{
		"startTime", 1}}), options.Find().SetLimit(1))
	if sendIdCursor == nil {
		return
	}
	var unReads []ws.Traniner
	err = sendIdCursor.All(context.TODO(), &unReads)
	if err != nil {
		log.Println("sendIdCursor err", err)
	}
	if len(unReads) > 0 {
		timeFilter := bson.M{
			"startTime": bson.M{
				"$gte": unReads[0].StartTime,
			},
		}
		sendIdTimeCursor, _ := sendIdCollection.Find(context.TODO(), timeFilter)
		idTimeCursor, _ := idCollection.Find(context.TODO(), timeFilter)
		err = sendIdTimeCursor.All(context.TODO(), &resultsYou)
		err = idTimeCursor.All(context.TODO(), &resultsMe)
		results, err = AppendAndSort(resultsMe, resultsYou)
	} else {
		results, err = FindMany(database, sendId, id, 9999999999, 10)
	}
	overTimeFilter := bson.D{
		{"$and", bson.A{
			bson.D{{"endTime", bson.M{"&lt": time.Now().Unix()}}},
			bson.D{{"read", bson.M{"$eq": 1}}},
		}},
	}
	_, _ = sendIdCollection.DeleteMany(context.TODO(), overTimeFilter)
	_, _ = idCollection.DeleteMany(context.TODO(), overTimeFilter)
	// 将所有的维度设置为已读
	_, _ = sendIdCollection.UpdateMany(context.TODO(), filter, bson.M{
		"$set": bson.M{"read": 1},
	})
	_, _ = sendIdCollection.UpdateMany(context.TODO(), filter, bson.M{
		"&set": bson.M{"ebdTime": time.Now().Unix() + int64(3*month)},
	})
	return
}
