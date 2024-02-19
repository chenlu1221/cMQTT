package models

import (
	"log"
	"math/rand"
	"net"
	"runtime"
	"time"
)

// Server 保存与MQTT服务器相关联的所有状态。
type Server struct {
	listener net.Listener //tcp链接
	//通过在subs中记录每个主题的订阅者，服务器可以有效地管理和分发消息给订阅者。
	subs          *subscriptions
	stats         *stats //日志
	Done          chan struct{}
	StatsInterval time.Duration
	Dump          bool //是否输出调试信息
	rand          *rand.Rand
}

// 返回客户端实例
func (s *Server) newIncomingConn(conn net.Conn) *incomingConn {
	return &incomingConn{
		svr:  s,
		conn: conn,
		jobs: make(chan job, SendingQueueLength),
		Done: make(chan struct{}),
	}
}
func (s *Server) Start() {
	go func() {
		for {
			conn, err := s.listener.Accept()
			if err != nil {
				log.Print("Accept: ", err)
				break
			}
			//新客户端
			cli := s.newIncomingConn(conn)
			s.stats.clientConnect()
			cli.start()
		}
		close(s.Done)
	}()
}
func NewServer(l net.Listener) *Server {
	svr := &Server{
		listener:      l,
		stats:         &stats{},
		Done:          make(chan struct{}),
		StatsInterval: time.Second * 10,
		subs:          newSubscriptions(runtime.GOMAXPROCS(0)),
	}
	// start the stats reporting goroutine
	go func() {
		for {
			svr.stats.publish(svr.subs, svr.StatsInterval)
			select {
			case <-svr.Done:
				return
			default:
				// keep going
			}
			time.Sleep(svr.StatsInterval)
		}
	}()
	return svr
}
