package models

import (
	"go_im_ftt/mqtt/messages"
	"sync"
)

const SendingQueueLength = 10000
const postQueue = 100

var clients = make(map[string]*incomingConn)

var clientsMu sync.Mutex

type receipt chan struct{}

func (r receipt) wait() {
	<-r
}

// 保留消息
type retain struct {
	m    messages.Publish // 保留消息的具体内容
	wild wild             // 与该保留消息相关的通配符信息
}

// publish 消息体
type post struct {
	c *incomingConn
	m *messages.Publish
}

func newSubscriptions(workers int) *subscriptions {
	s := &subscriptions{
		subs:    make(map[string][]*incomingConn),
		retain:  make(map[string]retain),
		posts:   make(chan post, postQueue),
		workers: workers,
	}
	for i := 0; i < s.workers; i++ {
		go s.run(i)
	}
	return s
}
