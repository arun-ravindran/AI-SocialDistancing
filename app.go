// Read from InfluxDB, determine social distancing violations, and write associated data (timestamp, person position, bones keypoints) to MongoDB
package main

import (
    "log"
	"fmt"
	"context"
	//"reflect"
	"time"
	//"strings"
    "github.com/influxdata/influxdb1-client/v2"

    //"go.mongodb.org/mongo-driver/bson"
    //"go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"

)

const dbname="socdist"
const measurement="socdist"
const influxdbEndpoint="http://172.17.0.3:8086"
const mongodbEndpoint="mongodb://172.17.0.2:27017"
const personField="data1"
const boneField="data2"
const mongodbName= "socdist"
const mongodbCol="kpCollection"

type ViolationRecord struct{
	Timestamp string	`bson:"timestamp,omitempty"`
	PersonCoords string `bson:"personcoords,omitempty"`
	BoneKeypoints string `bson:"bonekeypoints,omitempty"`
}

func main() {
	// Connect to InfluxDB
	influxdbClient, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: influxdbEndpoint,
	})
	if err != nil {
		log.Fatal("From app server, InfluxDB connect ", err)
	}
	defer influxdbClient.Close()

	// Connect to MongoDB
    ctx, _ := context.WithTimeout(context.Background(), 100*time.Second)
    client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongodbEndpoint)) //Container IP
    if err != nil {
        log.Fatal("From app server, MongoDB connect ", err)
    }
    defer client.Disconnect(ctx)

    database := client.Database(mongodbName)
    kpCollection := database.Collection(mongodbCol)

	// From InfluxDB

	// Find the first and last timestamps 
	tbegin := findEarliestTime(influxdbClient)

	// This will be an infinite loop
	for {
		tend := findLatestTime(influxdbClient)

		//t1, _ := time.Parse(time.RFC3339, tbegin)
		//t2, _ := time.Parse(time.RFC3339, tend)

		// Find the person coordinates, and the latest timestamp
		persons, ts := rangeQueryDB(influxdbClient, personField, tbegin, tend)

		// Check social distancing violations from person coordinates in a frame
		// Returns all person coordinates, and timestamps of frames with violations
		pViolations, tsViolations := findViolations(persons, ts)

		// For timestamps associated with violations, query bone keypoints
		// write person coordinates, bone keypoints, and timestamps to MongoDB
		var vrecMultiple []interface{}
		for i, tv := range tsViolations {
			bones := pointQueryDB(influxdbClient, boneField, tv)
			vrec := ViolationRecord {
				Timestamp: tv,
				PersonCoords: pViolations[i],
				BoneKeypoints: bones[0],
			}
			vrecMultiple = append(vrecMultiple, vrec)
		}
		// Write to MongoDB
		if len(vrecMultiple) != 0 {
			_, err = kpCollection.InsertMany(ctx, vrecMultiple)
			if err != nil {
				log.Fatal("From app server, MongoDB write ", err)
			}
			vrecMultiple = nil
		}
		tbegin = tend // For the next read from InfluxDB
	}
}

func findViolations(persons []string, ts []string)([]string, []string) {
	// TO DO: Determine distance between each pair of persons, and return the corresponding persons, timestamp
	return persons, ts
}


//Returns field values, and time stamps from InfluxDB from tbegin to tend
func rangeQueryDB(c client.Client, field, tbegin, tend string) ([]string, []string) {
	var res []string
	var ts []string
	var qstr string
	if (tbegin != tend) {
		qstr = fmt.Sprintf("SELECT %s from %s WHERE time > '%s' AND time <= '%s'", field, measurement, tbegin, tend)
		q := client.NewQuery(qstr, dbname, "")
		if response, err := c.Query(q); err == nil && response.Error() == nil {
			val := response.Results[0].Series[0].Values
			for _, v := range val {
				ts = append(ts, v[0].(string))
				res = append(res, v[1].(string))
			}
		}
	}
	return res, ts
}

//Returns field values, and time stamps from InfluxDB for a single timestamp
func pointQueryDB(c client.Client, field, ts string) ([]string) {
	var res []string
	var qstr string
	qstr = fmt.Sprintf("SELECT %s from %s WHERE time = '%s'", field, measurement, ts)
	q := client.NewQuery(qstr, dbname, "")
	if response, err := c.Query(q); err == nil && response.Error() == nil {
		val := response.Results[0].Series[0].Values
		for _, v := range val {
			res = append(res, v[1].(string))
		}
	}
	return res
}


func findEarliestTime(c client.Client) string {
	var tstr string
	qstr := fmt.Sprintf("SELECT * FROM %s LIMIT 1", measurement)
	q := client.NewQuery(qstr, dbname, "")
	if response, err := c.Query(q); err == nil && response.Error() == nil {
		tstr = response.Results[0].Series[0].Values[0][0].(string)
	}
	return tstr
}

func findLatestTime(c client.Client) string {
	var tstr string
	qstr := fmt.Sprintf("SELECT * FROM %s ORDER BY time DESC LIMIT 1", measurement)
	q := client.NewQuery(qstr, dbname, "")
	if response, err := c.Query(q); err == nil && response.Error() == nil {
		tstr = response.Results[0].Series[0].Values[0][0].(string)
	}
	return tstr
}



