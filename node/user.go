package node

import (
	"github.com/Yuzuki616/V2bX/api/panel"
	"strconv"
)

func compareUserList(old, new []panel.UserInfo) (deleted, added []panel.UserInfo) {
	tmp := map[string]struct{}{}
	tmp2 := map[string]struct{}{}
	for i := range old {
		tmp[old[i].Uuid+strconv.Itoa(old[i].SpeedLimit)] = struct{}{}
	}
	l := len(tmp)
	for i := range new {
		e := new[i].Uuid + strconv.Itoa(new[i].SpeedLimit)
		tmp[e] = struct{}{}
		tmp2[e] = struct{}{}
		if l != len(tmp) {
			added = append(added, new[i])
			l++
		}
	}
	tmp = nil
	l = len(tmp2)
	for i := range old {
		tmp2[old[i].Uuid+strconv.Itoa(old[i].SpeedLimit)] = struct{}{}
		if l != len(tmp2) {
			deleted = append(deleted, old[i])
			l++
		}
	}
	return deleted, added
}
