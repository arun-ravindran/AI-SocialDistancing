# Uses ZED SDK to capture person keypoints. Writes to InfluxDB

import pyzed.sl as sl
import numpy as np
import math
import time
from influxdb import InfluxDBClient
from datetime import datetime

influxDBHost = '172.17.0.3'
dbname = "socdist"

def main():
	# Connect to influx dB
	client = InfluxDBClient(host=influxDBHost, port=8086)
	client.create_database(dbname)
	client.switch_database(dbname)



	# Create a Camera object
	zed = sl.Camera()

	# Create a InitParameters object and set configuration parameters
	init_params = sl.InitParameters()
	init_params.camera_resolution = sl.RESOLUTION.HD720  # Use HD720 video mode
	#init_params.camera_resolution = sl.RESOLUTION.HD1080  # Use HD1080 video mode
	init_params.camera_fps = 15  # Set fps at 15
	init_params.depth_mode = sl.DEPTH_MODE.PERFORMANCE
	init_params.coordinate_units = sl.UNIT.METER


	# Open the camera
	err = zed.open(init_params)
	if err != sl.ERROR_CODE.SUCCESS:
		exit(1)

	mat = sl.Mat()

	runtime_parameters = sl.RuntimeParameters()
	key = ''

	zed.enable_positional_tracking()

	obj_param = sl.ObjectDetectionParameters()
	obj_param.enable_tracking = True
	obj_param.detection_model = sl.DETECTION_MODEL.HUMAN_BODY_FAST
	obj_param.enable_mask_output = True
	zed.enable_object_detection(obj_param)

	objects = sl.Objects()
	obj_runtime_param = sl.ObjectDetectionRuntimeParameters()
	obj_runtime_param.detection_confidence_threshold = 40

	numFrames = 0
	totalTime = 0
	while numFrames < 1000: 
		persons = []
		bones = []
		#startTime = time()
		# Grab an image, a RuntimeParameters object must be given to grab()
		if zed.grab(runtime_parameters) == sl.ERROR_CODE.SUCCESS:
			numFrames += 1
			# A new image is available if grab() returns SUCCESS
			zed.retrieve_image(mat, sl.VIEW.LEFT)
			zed.retrieve_objects(objects, obj_runtime_param)
			obj_array = objects.object_list
			for i in range(len(obj_array)) :
				obj_data = obj_array[i]
				keypoint = obj_data.keypoint_2d
				persons.append(obj_data.position) 
				                
				for bone in sl.BODY_BONES:
					kp1 = keypoint[bone[0].value]
					kp2 = keypoint[bone[1].value]
					bones.extend([int(kp1[0]), int(kp1[1]), int(kp2[0]), int(kp2[1])])
					            
		writeToInfluxDB(client, persons, bones)		

		#endTime = time()
		#totalTime += endTime - startTime
		#numFrames += 1
		#time.sleep(1)

	# Close the camera
	zed.close()

	# Print fps
	#if totalTime != 0: 
		#print("Average FPS: {:.2f}fps".format(numFrames / totalTime))


def writeToInfluxDB(client, persons, bones):
	pstr = "".join(str(p[0])+","+str(p[1])+","+str(p[2]) for p in persons)
	bstr = "".join(str(b)+"," for b in bones)
	pstr = pstr[:-1]
	bstr = bstr[:-1]

	json_body = [ 
    	{   
        	"measurement": "socdist",
        	"tags" : { 
            	"source-id" : "cam1"
        	},
        	"time": str(datetime.now().time()),
        	"fields": {
            	"data1" : pstr,
            	"data2": bstr
        	},
    	}   
	]

	# Write to db
	client.write_points(json_body)

if __name__ == "__main__":
    main()
