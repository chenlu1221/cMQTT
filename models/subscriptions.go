package models

import (
	"fmt"
	"go_im_ftt/mqtt/messages"
	"log"
	"strings"
	"sync"
)

// 管理订阅信息
type subscriptions struct {
	workers   int                        // 表示处理订阅的工作池的大小
	posts     chan post                  // 用于传递订阅消息的通道
	mu        sync.Mutex                 // 用于保护以下字段的互斥锁
	subs      map[string][]*incomingConn // 主题和对应的订阅者列表
	wildcards []wild                     // 记录使用通配符的订阅者信息
	retain    map[string]retain          // 记录topic的保留消息
	stats     *stats                     // 指向统计信息的指针
}

// topic的所有订阅者
func (s *subscriptions) subscribers(topic string) []*incomingConn {
	s.mu.Lock()
	defer s.mu.Unlock()

	res := s.subs[topic]

	parts := strings.Split(topic, "/")
	for _, w := range s.wildcards {
		if matches(w.wild, parts) {
			res = append(res, w.c)
		}
	}

	return res
}

// add订阅
func (s *subscriptions) add(topic string, c *incomingConn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	//通配符主题
	if isWildcard(topic) {
		w := wild{wild: strings.Split(topic, "/"), c: c}
		if w.valid() {
			s.wildcards = append(s.wildcards, w)
		}
	} else {
		s.subs[topic] = append(s.subs[topic], c)
	}
}

// 发送保留消息
func (s *subscriptions) sendRetain(topic string, c *incomingConn) {
	s.mu.Lock()
	var tlist []string
	if isWildcard(topic) {
		//遍历topic
		for key, _ := range s.subs {
			if matches(strings.Split(topic, "/"), strings.Split(key, "/")) {
				tlist = append(tlist, key)
			}
		}
	} else {
		tlist = []string{topic}
	}
	for _, t := range tlist {
		if r, ok := s.retain[t]; ok {
			c.submit(&r.m)
		}
	}
	s.mu.Unlock()
}

// 取消订阅
func (s *subscriptions) unsub(topic string, c *incomingConn) {
	s.mu.Lock()
	if subs, ok := s.subs[topic]; ok {
		nils := 0

		for i, sub := range subs {
			if sub == c {
				subs[i] = nil
				nils++
			}
			if sub == nil {
				nils++
			}
		}

		if nils == len(subs) {
			delete(s.subs, topic)
		}
	}
	s.mu.Unlock()
}
func (s *subscriptions) unsubAll(c *incomingConn) {
	s.mu.Lock()
	for _, v := range s.subs {
		for i := range v {
			if v[i] == c {
				v[i] = nil
			}
		}
	}

	// 通配符
	var wildNew []wild
	for i := 0; i < len(s.wildcards); i++ {
		if s.wildcards[i].c != c {
			wildNew = append(wildNew, s.wildcards[i])
		}
	}
	s.wildcards = wildNew

	s.mu.Unlock()
}

// 订阅消息的通道传入消息
func (s *subscriptions) submit(c *incomingConn, m *messages.Publish) {
	s.posts <- post{c: c, m: m}
}

// 分发消息给订阅者
func (s *subscriptions) run(id int) {
	tag := fmt.Sprintf("worker %d ", id)
	log.Print(tag, "started")
	for post := range s.posts {
		//保留消息
		isRetain := post.m.Header.Retain
		post.m.Header.Retain = false

		//处理删除保留消息的情况
		if isRetain && post.m.Payload.Size() == 0 {
			s.mu.Lock()
			delete(s.retain, post.m.TopicName)
			s.mu.Unlock()
			return
		}

		// 找到post的所有订阅者
		subscribers := s.subscribers(post.m.TopicName)
		for _, c := range subscribers {
			if c == post.c {
				continue
			}
			if c != nil {
				c.submit(post.m)
			}
		}
		//保留保留消息
		if isRetain {
			s.mu.Lock()
			msg := *post.m
			msg.Header.Retain = true
			s.retain[post.m.TopicName] = retain{m: msg}
			s.mu.Unlock()
		}
	}
}
