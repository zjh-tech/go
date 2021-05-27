package frame

import (
	"errors"
	"sync"
	"time"
)

const (
	//t := time.Date(2015, 1, 1, 00, 00, 00, 00, time.Local).UnixNano() / 1e6;//获取时间戳 毫秒
	//41位的时间截，可以使用69年，年T = (1L << 41) / (1000L * 60 * 60 * 24 * 365) = 69<br>
	epoch        int64 = 1420041600000
	serverIdBits int64 = 12
	//0-4095
	maxServerId  int64 = -1 ^ (-1 << serverIdBits)
	sequenceBits int64 = 10
	//0-1023
	sequenceMask int64 = -1 ^ (-1 << sequenceBits)
	// 数据标识id向左移10位
	serverIdShift int64 = sequenceBits
	// 时间截向左移22位
	timestampLeftShift int64 = sequenceBits + serverIdBits
)

type IdMaker struct {
	mutex         sync.Mutex // 添加互斥锁 确保并发安全
	lastTimestamp int64      // 上次生成ID的时间截
	serverId      int64
	sequence      int64 // 毫秒内序列(0~4095)
}

//(0-4095)
func NewIdMaker(serverId int64) (*IdMaker, error) {
	if serverId < 0 || serverId > maxServerId {
		return nil, errors.New("Server ID excess of quantity")
	}
	return &IdMaker{
		lastTimestamp: 0,
		serverId:      serverId,
		sequence:      0,
	}, nil
}

func (m *IdMaker) nextId() (int64, error) {
	now := time.Now().UnixNano() / 1e6
	if now < m.lastTimestamp {
		return 0, errors.New("Clock moved backwards")
	}
	if m.lastTimestamp == now {
		m.sequence = (m.sequence + 1) & sequenceMask
		if m.sequence == 0 {
			// 阻塞到下一个毫秒，直到获得新的时间戳
			for now <= m.lastTimestamp {
				now = time.Now().UnixNano() / 1e6
			}
		}
	} else {
		m.sequence = 0
	}

	m.lastTimestamp = now
	//1(不用) + 41(41位的时间截，可以使用69年) + 12(4095) + 10(1023) = 64位
	ID := int64((now-epoch)<<timestampLeftShift | m.serverId<<serverIdShift | m.sequence)
	return ID, nil
}

func (m *IdMaker) NextId() (int64, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	id, err := m.nextId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (m *IdMaker) NextIds(num int) ([]int64, error) {
	ids := make([]int64, num)
	m.mutex.Lock()
	defer m.mutex.Unlock()
	var err error
	for i := 0; i < num; i++ {
		ids[i], err = m.nextId()
		if err != nil {
			return nil, err
		}
	}
	return ids, nil
}

var GIdMaker *IdMaker
