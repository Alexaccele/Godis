### http测试结果

与Redis相差3.*倍
```type is http
   server is localhost
   port is 9090
   total 100000 requests
   data size is 1000
   we have 1 connections
   operation is set
   keyspacelen is 100000
   pipeline length is 1
   0 records get
   0 records miss
   100000 records set
   7.170418 seconds total
   99% requests < 1 ms
   99% requests < 3 ms
   100% requests < 4 ms
   69 usec average for each request
   throughput is 13.946189 MB/s
   rps is 13946.189318
   type is http
   server is localhost
   port is 9090
   total 100000 requests
   data size is 1000
   we have 1 connections
   operation is get
   keyspacelen is 100000
   pipeline length is 1
   63288 records get
   36712 records miss
   0 records set
   7.748722 seconds total
   99% requests < 1 ms
   99% requests < 2 ms
   99% requests < 3 ms
   99% requests < 4 ms
   99% requests < 5 ms
   100% requests < 6 ms
   75 usec average for each request
   throughput is 8.167541 MB/s
   rps is 12905.354191
   ====== SET ======
     100000 requests completed in 2.37 seconds
     1 parallel clients
     1000 bytes payload
     keep alive: 1
   
   100.00% <= 1 milliseconds
   100.00% <= 1 milliseconds
   42176.30 requests per second
   
   ====== GET ======
     100000 requests completed in 2.24 seconds
     1 parallel clients
     1000 bytes payload
     keep alive: 1
   
   100.00% <= 0 milliseconds
   44662.79 requests per second
```

### tcp测试结果
相比于http协议，性能提升了3倍
此时与Redis相差很小，但仍有较小的差距
```
 type is tcp
 server is localhost
 port is 2333
 total 100000 requests
 data size is 1000
 we have 1 connections
 operation is set
 keyspacelen is 100000
 pipeline length is 1
 0 records get
 0 records miss
 100000 records set
 2.506590 seconds total
 99% requests < 1 ms
 99% requests < 2 ms
 100% requests < 5 ms
 23 usec average for each request
 throughput is 39.894845 MB/s
 rps is 39894.844769
 
 
 type is tcp
 server is localhost
 port is 2333
 total 100000 requests
 data size is 1000
 we have 1 connections
 operation is get
 keyspacelen is 100000
 pipeline length is 1
 63643 records get
 36357 records miss
 0 records set
 2.840474 seconds total
 99% requests < 1 ms
 99% requests < 2 ms
 99% requests < 3 ms
 99% requests < 4 ms
 100% requests < 5 ms
 26 usec average for each request
 throughput is 22.405766 MB/s
 rps is 35205.389161
 
 
 Redis：
 ====== SET ======
   100000 requests completed in 2.29 seconds
   1 parallel clients
   1000 bytes payload
   keep alive: 1
 
 100.00% <= 0 milliseconds
 43649.06 requests per second
 
 ====== GET ======
   100000 requests completed in 2.25 seconds
   1 parallel clients
   1000 bytes payload
   keep alive: 1
 
 100.00% <= 1 milliseconds
 44424.70 requests per second
```
## 增加并发数量
当并发量为100时

```$xslt
type is tcp
server is localhost
port is 2333
total 100000 requests
data size is 1000
we have 100 connections
operation is set
keyspacelen is 100000
pipeline length is 1
0 records get
0 records miss
100000 records set
0.692045 seconds total
87% requests < 1 ms
97% requests < 2 ms
98% requests < 3 ms
99% requests < 4 ms
99% requests < 5 ms
99% requests < 6 ms
99% requests < 7 ms
99% requests < 8 ms
99% requests < 9 ms
99% requests < 10 ms
99% requests < 11 ms
99% requests < 12 ms
99% requests < 14 ms
99% requests < 18 ms
99% requests < 19 ms
99% requests < 20 ms
99% requests < 21 ms
99% requests < 23 ms
100% requests < 26 ms
658 usec average for each request
throughput is 144.499232 MB/s
rps is 144499.232340

type is tcp
server is localhost
port is 2333
total 100000 requests
data size is 1000
we have 100 connections
operation is get
keyspacelen is 100000
pipeline length is 1
62908 records get
37092 records miss
0 records set
0.671856 seconds total
86% requests < 1 ms
97% requests < 2 ms
99% requests < 3 ms
99% requests < 4 ms
99% requests < 5 ms
99% requests < 6 ms
99% requests < 7 ms
99% requests < 8 ms
99% requests < 9 ms
99% requests < 10 ms
99% requests < 11 ms
99% requests < 12 ms
99% requests < 13 ms
100% requests < 17 ms
650 usec average for each request
throughput is 93.633104 MB/s
rps is 148841.330892


====== SET ======
  100000 requests completed in 1.14 seconds
  100 parallel clients
  1000 bytes payload
  keep alive: 1

89.11% <= 1 milliseconds
97.68% <= 2 milliseconds
99.94% <= 3 milliseconds
100.00% <= 3 milliseconds
87719.30 requests per second

====== GET ======
  100000 requests completed in 1.01 seconds
  100 parallel clients
  1000 bytes payload
  keep alive: 1

92.93% <= 1 milliseconds
99.92% <= 2 milliseconds
100.00% <= 2 milliseconds
99304.87 requests per second
```
可以看到此时的结果，Godis的rps已经达到了14W，而Redis为8/9W
## 备份情况记录
备份6万条数据耗时180ms性能很差，待优化
