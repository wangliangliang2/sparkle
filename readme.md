### Aim to:  simple code , esay reading.

#### Support Protocol:

-  RTMP
-  AMF
-  HLS
-  HTTP-FLV

#### Support Containner:

-  FLV
-  TS

#### Support Encoding:

-  H264
-  AAC

#### Useage:

```shell
ffmpeg -re -i llxw.mp4   -vcodec copy -acodec copy -f flv rtmp://localhost:1935/live/movie
ffplay rtmp://localhost:1935/live/movie 			// rtmp
ffplay http://localhost:2222/live/movie 			//flv
ffplay http://localhost:3333/live/movie.m3u8 // hls
```

