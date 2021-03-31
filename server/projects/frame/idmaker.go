package frame

import (
	"errors"
	"projects/go-engine/elog"
	"sync"
	"time"
)

const (
	//t := time.Date(2015, 1, 1, 00, 00, 00, 00, time.Local).UnixNano() / 1e6;//获取时间戳 毫秒
	//41位的时间截，可以使用69年，年T = (1L << 41) / (1000L * 60 * 60 * 24 * 365) = 69<br>
	epoch          int64 = 1420041600000
	server_id_bits int64 = 12
	//0-4095
	max_server_id int64 = -1 ^ (-1 << server_id_bits)
	sequence_bits int64 = 10
	//0-1023
	sequence_mask int64 = -1 ^ (-1 << sequence_bits)
	// 数据标识id向左移10位
	server_id_shift int64 = sequence_bits
	// 时间截向左移22位
	timestamp_left_shift int64 = sequence_bits + server_id_bits
)

type IdMaker struct {
	mutex          sync.Mutex // 添加互斥锁 确保并发安全
	last_timestamp int64      // 上次生成ID的时间截
	server_id      int64
	sequence       int64 // 毫秒内序列(0~4095)
}

//(0-4095)
func NewIdMaker(server_id int64) (*IdMaker, error) {
	if server_id < 0 || server_id > max_server_id {
		return nil, errors.New("Server ID excess of quantity")
	}
	return &IdMaker{
		last_timestamp: 0,
		server_id:      server_id,
		sequence:       0,
	}, nil
}

func (m *IdMaker) next_id() (int64, error) {
	now := time.Now().UnixNano() / 1e6
	if now < m.last_timestamp {
		return 0, errors.New("Clock moved backwards")
	}
	if m.last_timestamp == now {
		m.sequence = (m.sequence + 1) & sequence_mask
		if m.sequence == 0 {
			// 阻塞到下一个毫秒，直到获得新的时间戳
			for now <= m.last_timestamp {
				now = time.Now().UnixNano() / 1e6
			}
		}
	} else {
		m.sequence = 0
	}

	m.last_timestamp = now
	//1(不用) + 41(41位的时间截，可以使用69年) + 12(4095) + 10(1023) = 64位
	ID := int64((now-epoch)<<timestamp_left_shift | m.server_id<<server_id_shift | m.sequence)
	return ID, nil
}

func (m *IdMaker) NextId() int64 {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	id, err := m.next_id()
	if err != nil {
		elog.Errorf("Error=%v", err)
		GServer.Quit()
	}

	return id
}

func (m *IdMaker) NextIds(num int) []int64 {
	ids := make([]int64, num)
	m.mutex.Lock()
	defer m.mutex.Unlock()
	var err error
	for i := 0; i < num; i++ {
		ids[i], err = m.next_id()
		if err != nil {
			elog.Errorf("Error=%v", err)
			GServer.Quit()
		}
	}
	return ids
}

var GIdMaker *IdMaker
