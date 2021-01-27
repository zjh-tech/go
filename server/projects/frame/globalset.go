package frame

const (
	DEFAULT_SERVER_TYPE  uint32 = 0
	REGISTRY_SERVER_TYPE uint32 = 1
	LOGIN_SERVER_TYPE    uint32 = 2
	GATEWAY_SERVER_TYPE  uint32 = 3
	CENTER_SERVER_TYPE   uint32 = 4
	MATCH_SERVER_TYPE    uint32 = 5
	HALL_SERVER_TYPE     uint32 = 6
	BATTLE_SERVER_TYPE   uint32 = 7
	DB_SERVER_TYPE       uint32 = 8

	ROBOT_CLIENT_TYPE uint32 = 1000
)

const (
	TS_RANK_RETISTER_SERVER_TYPE uint32 = 1
	TS_RANK_BALANCE_SERVER_TYPE  uint32 = 2
	TS_RANK_GATEWAY_SERVER_TYPE  uint32 = 3
	TS_RANK_RANK_SERVER_TYPE     uint32 = 4

	TS_RANK_ROBOT_TYPE uint32 = 1001
)

type TS_RANK_SERVICE_TYPE uint32

//var ServerNameArrays = [...]string{
//	"nil",
//	"registryserver",
//	"loginserver",
//	"gateserver",
//}
//
//func PrintServerState(serverType uint32, id uint64, connFlag bool) {
//	var State string
//	if connFlag {
//		State = " is Connected."
//	} else {
//		State = " is Terminate."
//	}
//	elog.InfoAf("%v (%v) %v", ServerNameArrays[serverType], id, State)
//}

const (
	MSG_SUCCESS uint32 = 0
	MSG_FAIL    uint32 = 1
)

const (
	METER_LOOP_COUNT = 20
	NET_LOOP_COUNT   = 100
	TIMER_LOOP_COUNT = 60000
	HTTP_LOOP_COUNT  = 100
	DB_LOOP_COUNT    = 100
)
