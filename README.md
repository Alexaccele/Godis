## Godis

用Go语言实现的一个缓存系统

### 功能
* 提供http方式实现缓存功能
* 提供tcp方式实现缓存功能
* 提供数据持久化功能
* 提供数据淘汰功能
* 支持缓存过期时间设置，以秒位单位
* 提供配置文件进行部分参数配置

### 运行
`go run main.go`将读取默认配置文件config.toml，可通过`-conf`参数指定配置文件路径。

服务将同时启动http与tcp两种服务方式。

### http基本接口
1. `get /status`获取缓存数量与占用内存大小
2. `get /cache/key`获取键为key的缓存数据
3. `put /cache/key -body:value`设置键为key,缓存数据为value
4. `del /cache/key`删除键为key的缓存数据

### tcp方式
参考`client`文件夹下的客户端测试程序，具体实现见`cache-benchmark/cacheClient/tcp.go`。