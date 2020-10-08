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

