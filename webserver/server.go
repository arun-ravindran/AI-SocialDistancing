package main

import (
	"bytes"
    "context"
    "log"
    "net/http"
    "encoding/base64"
    "time"
    "github.com/gorilla/websocket"

	"go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/gridfs"
    "go.mongodb.org/mongo-driver/mongo/options"

)

const (
    socketBufferSize  = 1024
	mongodbEndpoint="mongodb://172.17.0.2:27017"
	mongodbName= "socdist"
	mongodbCol="kpCollection"
	backgroundImage = "blank.jpg"

)

var upgrader = &websocket.Upgrader{
    ReadBufferSize:  socketBufferSize,
    WriteBufferSize: socketBufferSize,
}


func sceneHandler(w http.ResponseWriter, req *http.Request) {
    socket, err := upgrader.Upgrade(w, req, nil)
    if err != nil {
        log.Fatal("ServeHTTP:", err)
        return
    }

    _, err = req.Cookie("auth")
    if err != nil {
        log.Fatal("Failed to get auth cookie:", err)
        return
    }

	type ViolationRecord struct{
		Timestamp string	`bson:"timestamp,omitempty"`
		PersonCoords string `bson:"personcoords,omitempty"`
		BoneKeypoints string `bson:"bonekeypoints,omitempty"`
	}


	// Connect to MongoDB
    conn := InitiateMongoClient()
    db := conn.Database(mongodbName)

	res := make(map[string]interface{})
	var img64 []byte

	// Read a specific image from MongoDB
	fileName := backgroundImage
	bucket, _ := gridfs.NewBucket(
        db,
    )
    var buf bytes.Buffer
    _, err = bucket.DownloadToStreamByName(fileName, &buf)
    if err != nil {
        log.Fatal("From server, image read from MongoDB ", err)
    }
	img64 = buf.Bytes()
	// Send image 
	str := base64.StdEncoding.EncodeToString(img64)
	res["img64"] = str
	res["type"] = "image"
	if err = socket.WriteJSON(&res); err != nil {
		log.Println("From server: web socket write image ", err)
	}

	//Read keypoints from MongoDB
	var vr ViolationRecord

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
    keypointsCollection := db.Collection(mongodbCol)
	//err = keypointsCollection.FindOne(ctx, bson.M{}).Decode(&vr)
	opts := options.Find()
	opts.SetSort(bson.D{{"timestamp", -1}}) // Reverse sorted by timestamp
	cursor, err := keypointsCollection.Find(ctx, bson.M{}, opts) // Read all
    if err != nil {
        log.Fatal("From server, keypoint read from MongoDB", err)
    }
	defer cursor.Close(ctx)
/*
	// Get the latest record for displaying only the latest violation
	cursor.Next(ctx)
	if err = cursor.Decode(&vr); err != nil {
		log.Fatal("From server, curser decode MongoDB ", err)
	}

	res["type"] = "keypoint"
	res["ts"] = vr.Timestamp
	res["pcord"] = vr.PersonCoords
	res["bkp"] = vr.BoneKeypoints

	if err = socket.WriteJSON(&res); err != nil {
			log.Println("From server: websocket write keypoint ", err)
	}
*/

	// Get all records - reverse timestamp
    for cursor.Next(ctx) {
        if err = cursor.Decode(&vr); err != nil {
            log.Fatal(err)
        }
		res["type"] = "keypoint"
		res["ts"] = vr.Timestamp
		res["pcord"] = vr.PersonCoords
		res["bkp"] = vr.BoneKeypoints

		if err = socket.WriteJSON(&res); err != nil {
			log.Println("From server: ", err)
		}

    }

}

func InitiateMongoClient() *mongo.Client {
    var err error
    var client *mongo.Client
    opts := options.Client()
    opts.ApplyURI(mongodbEndpoint)
	opts.SetMaxPoolSize(5)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
    if client, err = mongo.Connect(ctx, opts); err != nil {
        log.Fatal("From server, MongoDB connection initiate error", err.Error())
    }
    return client
}

