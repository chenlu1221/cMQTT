package models

import (
	"sync/atomic"
	"time"
)

// 服务器状态信息
type stats struct {
	recv       int64 // 表示接收的消息数量
	sent       int64 // 表示发送的消息数量
	clients    int64 // 表示当前连接的客户端数量
	clientsMax int64 // 表示连接的客户端数量的最大值
	lastmsgs   int64 // 表示最近消息的数量或相关统计
}

func (s *stats) messageRecv()      { atomic.AddInt64(&s.recv, 1) }
func (s *stats) messageSend()      { atomic.AddInt64(&s.sent, 1) }
func (s *stats) clientConnect()    { atomic.AddInt64(&s.clients, 1) }
func (s *stats) clientDisconnect() { atomic.AddInt64(&s.clients, -1) }

func (s *stats) publish(sub *subscriptions, interval time.Duration) {
	clients := atomic.LoadInt64(&s.clients)
	clientsMax := atomic.LoadInt64(&s.clientsMax)
	if clients > clientsMax {
		clientsMax = clients
		atomic.StoreInt64(&s.clientsMax, clientsMax)
	}
	sub.submit(nil, statsMessage("$SYS/broker/clients/active", clients))
	sub.submit(nil, statsMessage("$SYS/broker/clients/maximum", clientsMax))
	sub.submit(nil, statsMessage("$SYS/broker/messages/received",
		atomic.LoadInt64(&s.recv)))
	sub.submit(nil, statsMessage("$SYS/broker/messages/sent",
		atomic.LoadInt64(&s.sent)))

	msgs := atomic.LoadInt64(&s.recv) + atomic.LoadInt64(&s.sent)
	msgpersec := (msgs - s.lastmsgs) / int64(interval/time.Second)
	// no need for atomic because we are the only reader/writer of it
	s.lastmsgs = msgs

	sub.submit(nil, statsMessage("$SYS/broker/messages/per-sec", msgpersec))
}
