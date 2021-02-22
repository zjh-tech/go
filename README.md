# 环境:  go1.14.2版本

# 库:
## log:
*同步,异步API
## timer: 
*时间路定时器
*调时间: 时间回滚,前进
*Delay的API
*执行时间太长导致的误差越来越大
## tcp:
*粘包，断包
*主动被动断开,主动断开时支持发送数据
*本地字节序和网络字节序转换
*监听地址端口重用
*加解密,压缩,coder通用接口
## mysql:
*同步,异步API
*支持分库分表,事务
*自动拼接Sql语句工具
*liquibase:分库分表sql的管理
*DBRecordSet支持基础类型的转换
## redis: 
*同步API    
## http： 
*httpserver,httpclient
# 框架:    
## 服务发现: 支持tcp/http二种方式
## 唯一ID生成
## 策划配置加载和热更        
## 信号    
## protobuf3.0以上版本使用
## tcp:        
*心跳,验证
*白黑名单,验证超时待开发
*csclientsession: 用于Unity与Server之间
*ssclientsession: 用于SdkServer与Server之间
*ssserversession: 用于Server与Server之间
*Package: 包头 + 包体,包体: attach + other
*Coder: 自定义Package
*TimerMeter: 监控逻辑模块时间
*LogicServer: Server与Server接口
# 工具:
*打表工具: excel转protobuf 自动化代码生成,支持热更

# QPS
*测试环境:虚拟机,8核CPU...
##  mysql: select 6500 insert 4100 
##  log:   12万
##  redis: 1.2万
##  tcp:   38万
