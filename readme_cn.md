### 目标：简洁、容易阅读的代码

#### 支持的传输协议

-  RTMP
-  AMF
-  HLS
-  HTTP-FLV

#### 支持的容器格式

-  FLV
-  TS

#### 支持的编码格式

-  H264
-  AAC

#### 使用方式：

```shell
ffmpeg -re -i llxw.mp4   -vcodec copy -acodec copy -f flv rtmp://localhost:1935/live/movie
ffplay rtmp://localhost:1935/live/movie 			// rtmp
ffplay http://localhost:2222/live/movie 			//flv
ffplay http://localhost:3333/live/movie.m3u8 // hls
```

