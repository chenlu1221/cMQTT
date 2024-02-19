package models

import (
	"go_im_ftt/mqtt/messages"
	"strings"
)

// 判断topic是否包含通配符
func isWildcard(topic string) bool {
	if strings.Contains(topic, "#") || strings.Contains(topic, "+") {
		return true
	}
	return false
}

// stats消息创建
func statsMessage(topic string, stat int64) *messages.Publish {
	return &messages.Publish{
		Header: messages.FixedHeader{
			MessageType: messages.MsgPublish,
			Dup:         false,
			QoS:         messages.QosAtMostOnce,
			Retain:      true,
		},
		TopicName:        topic,
		PacketIdentifier: 0,
		Payload:          newIntPayload(stat),
	}
}
