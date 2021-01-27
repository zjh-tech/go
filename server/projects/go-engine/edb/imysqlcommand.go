package edb

type IMysqlCommand interface {
	//协程执行mysql操作
	OnExecuteSql(IMysqlConn)
	//主协程处理返回结果
	OnExecuted()
}
