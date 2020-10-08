#  Edge based social distancing
Zed2 is connected to Edge server via USB
Edge server hosts the following - 
 - AI Vision based human keypoint detector (Python)
 - InfluxDB for storing output of keypoint detector (Golang)
 - Background scene capture (Python)
 - Social distancing application that ingests data from InfluxDB (Golang)
 - MongoDB for storing output of social distancing application, and background scene (Golang)
 - Webserver (Golang)
Web front end (HTML, CSS, Bootstrap and Javascript)


To run, in separate terminals
- start up mongodb and influxdb containers. Check container IP addresses
- go run utils/image\_upload.go to load background image in mongodb
- python3 visionAI.py
- go run app.go
- go run webserver/*.go
