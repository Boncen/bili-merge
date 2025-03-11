# How to use
bili-merge 通过ffmpeg对从bilibili缓存的文件进行批量合并成mp4. 

从bilibili缓存的目录结构如下：
```
1803768206
├── c_1523511798
│   ├── 32
│   │   ├── audio.m4s
│   │   ├── index.json
│   │   └── video.m4s
│   ├── danmaku.xml
│   └── entry.json
├── c_1523512589
│   ├── 32
│   │   ├── audio.m4s
│   │   ├── index.json
│   │   └── video.m4s
│   ├── danmaku.xml
│   └── entry.json
├── c_1523514876
│   ├── 32
│   │   ├── audio.m4s
│   │   ├── index.json
│   │   └── video.m4s
│   ├── danmaku.xml
│   └── entry.json
...
```

指定输入目录参数：
```sh
bili-merge ./1803768206
```

一次传递多个目录：
```sh
bili-merge ./1803768206 ./1803768207 ./1803768208
```

```sh
bili-merge  # 无参数直接运行取当前目录下所有子目录
```

> 程序依赖ffmpeg命令，需要提前安装并配置好环境变量。