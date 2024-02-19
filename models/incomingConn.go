package models

import (
	"go_im_ftt/mqtt/messages"
	"io"
	"log"
	"net"
	"strings"
)

// 客户端实例
type incomingConn struct {
	svr      *Server       // 指向创建该连接的服务器的指针
	conn     net.Conn      // 表示与客户端建立的网络连接
	jobs     chan job      // 用于传递处理任务的通道
	clientid string        // 客户端标识符，唯一的标识某个客户端的字符串
	Done     chan struct{} // 表示连接处理完成的通道，用于通知其他地方连接已经结束
}

func (c *incomingConn) add() *incomingConn {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	existing, ok := clients[c.clientid]
	if ok {
		// 连接已经存在
		return existing
	}

	clients[c.clientid] = c
	return nil
}
func (c *incomingConn) submitSync(m messages.Message) receipt {
	j := job{m: m, r: make(receipt)}
	c.jobs <- j
	return j.r
}

type job struct {
	m messages.Message
	r receipt
}

func (c *incomingConn) del() {
	clientsMu.Lock()
	delete(clients, c.clientid)
	clientsMu.Unlock()
	return
}
func (c *incomingConn) submit(m messages.Message) {
	j := job{m: m}
	select {
	case c.jobs <- j:
	default:
		log.Print(c, ": failed to submit message")
	}
	return
}
func (c *incomingConn) start() {
	go c.reader()
	go c.writer()
}
func (c *incomingConn) reader() {
	defer func() {
		err := c.conn.Close()
		if err != nil {
			log.Println("close conn err:", c.clientid)
		}
		c.svr.stats.clientDisconnect()
		//关闭客户端的消息通道
		close(c.jobs)
	}()
	for {
		//从conn中拿出一条消息
		m, err := messages.DecodeOneMessage(c.conn, nil)
		if err != nil {
			if err == io.EOF {
				return
			}
			if strings.HasSuffix(err.Error(), "use of closed network connection") {
				return
			}
			log.Print("reader: ", err)
			return
		}
		//消息数+1
		c.svr.stats.messageRecv()

		if c.svr.Dump {
			log.Printf("dump  in: %T", m)
		}

		switch m := m.(type) {
		case *messages.Connect:
			//返回码0
			rc := messages.RetCodeAccepted

			if m.ProtocolName != "MQTT" ||
				m.ProtocolLevel != 4 {
				log.Print("reader: reject connection from ", m.ProtocolName, " version ", m.ProtocolLevel)
				rc = messages.RetCodeUnacceptableProtocolVersion
			}

			if len(m.Payload.ClientId) < 1 || len(m.Payload.ClientId) > 23 {
				rc = messages.RetCodeIdentifierRejected
			}
			c.clientid = m.Payload.ClientId

			// 连接已经存在
			if existing := c.add(); existing != nil {
				disconnect := messages.GetDisconnect()
				r := existing.submitSync(disconnect)
				r.wait()
				c.add()
			}

			//构造确认连接请求
			connack := messages.GetConNack()
			connack.ReturnCode = rc
			//提交到客户端的消息通道
			c.submit(connack)

			if rc != messages.RetCodeAccepted {
				log.Printf("Connection refused for %v: %v", c.conn.RemoteAddr(), ConnectionErrors[rc])
				return
			}

			//判断CleanSession的值
			clean := 0
			if m.ConnectFlags&0x02 > 0 {
				clean = 1
			}

			log.Printf("New client connected from %v as %v (c%v, k%v).", c.conn.RemoteAddr(), c.clientid, clean, m.KeepAlive)

		case *messages.Publish:
			//只接受qos0
			if m.Header.QoS != messages.QosAtMostOnce {
				log.Printf("reader: no support for QoS %v yet", m.Header.QoS)
				return
			}
			if m.PacketIdentifier == 0 {
				log.Printf("reader: invalid MessageId in PUBLISH.")
				return
			}
			if isWildcard(m.TopicName) {
				log.Print("reader: ignoring PUBLISH with wildcard topic ", m.TopicName)
			} else {
				c.svr.subs.submit(c, m)
			}
			back := messages.GetPubBack()
			back.PacketIdentifier = m.PacketIdentifier
			c.submit(back)

		case *messages.PingReq:
			c.submit(messages.GetPingResp())

		case *messages.SubScribe:
			if m.Header.QoS != messages.QosAtLeastOnce {
				// protocol error, disconnect
				return
			}
			if m.PacketIdentifier == 0 {
				log.Printf("reader: invalid MessageId in SUBSCRIBE.")
				return
			}
			suback := messages.GetSubAck()
			suback.PacketIdentifier = m.PacketIdentifier
			suback.QoSCode = make([]messages.QoS, len(m.TopicList))
			for i, tq := range m.TopicList {
				c.svr.subs.add(tq.Topic, c)
				suback.QoSCode[i] = messages.QosAtMostOnce
			}
			c.submit(suback)

			//如果订阅的主题有保留消息，则发送
			for _, tq := range m.TopicList {
				c.svr.subs.sendRetain(tq.Topic, c)
			}

		case *messages.UnSubScribe:
			if m.Header.QoS != messages.QosAtMostOnce && m.PacketIdentifier == 0 {
				log.Printf("reader: invalid MessageId in UNSUBSCRIBE.")
				return
			}
			for _, t := range m.TopicList {
				c.svr.subs.unsub(t, c)
			}
			unAck := messages.GetUnSubAck()
			unAck.PacketIdentifier = m.PacketIdentifier
			c.submit(unAck)

		case *messages.Disconnect:
			return

		default:
			log.Printf("reader: unknown msg type %T", m)
			return
		}
	}
}

func (c *incomingConn) writer() {

	defer func() {
		c.conn.Close()
		c.del()
		c.svr.subs.unsubAll(c)
	}()

	for job := range c.jobs {
		if c.svr.Dump {
			log.Printf("dump out: %T", job.m)
		}

		err := job.m.Encode(c.conn)
		if job.r != nil {
			close(job.r)
		}
		if err != nil {
			oe, isoe := err.(*net.OpError)
			if isoe && oe.Err.Error() == "use of closed network connection" {
				return
			}
			if err.Error() == "use of closed network connection" {
				return
			}

			log.Print("writer: ", err)
			return
		}
		c.svr.stats.messageSend()
		if _, ok := job.m.(*messages.Disconnect); ok {
			log.Print("writer: sent disconnect message")
			return
		}
	}
}
