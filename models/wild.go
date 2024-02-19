package models

// 通配符订阅
type wild struct {
	wild []string
	c    *incomingConn
}

// 判断past是否被wild通配符订阅包含
func matches(will []string, past []string) bool {
	if will[0] == "#" {
		return true
	}
	if will[0] == "+" && len(past) == len(will) {
		return matches(will[1:], past[1:])
	}
	if will[0] != past[0] && will[0] != "+" {
		return false
	}
	return matches(will[1:], past[1:])
}

// topic是否合规
func (w wild) valid() bool {
	for i, part := range w.wild {
		//finance#
		if isWildcard(part) && len(part) != 1 {
			return false
		}
		if part == "#" && i != len(w.wild)-1 {
			return false
		}
	}
	return true
}
