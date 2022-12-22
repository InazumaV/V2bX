package panel

import (
	"fmt"
	"github.com/goccy/go-json"
)

type OnlineUser struct {
	UID int
	IP  string
}

type V2RayUserInfo struct {
	Uuid    string `json:"uuid"`
	Email   string `json:"email"`
	AlterId int    `json:"alter_id"`
}
type TrojanUserInfo struct {
	Password string `json:"password"`
}
type UserInfo struct {
	Id         int    `json:"id"`
	Uuid       string `json:"uuid"`
	Email      string `json:"-"`
	SpeedLimit int    `json:"speed_limit"`
	Traffic    int64  `json:"-"`
}

type UserListBody struct {
	//Msg  string `json:"msg"`
	Users []UserInfo `json:"users"`
}

// GetUserList will pull user form sspanel
func (c *Client) GetUserList() (UserList []UserInfo, err error) {
	const path = "/api/v1/server/UniProxy/user"
	res, err := c.client.R().
		Get(path)
	err = c.checkResponse(res, path, err)
	if err != nil {
		return nil, err
	}
	var userList *UserListBody
	err = json.Unmarshal(res.Body(), &userList)
	if err != nil {
		return nil, fmt.Errorf("unmarshal userlist error: %s", err)
	}
	return userList.Users, nil
}

type UserTraffic struct {
	UID      int
	Upload   int64
	Download int64
}

// ReportUserTraffic reports the user traffic
func (c *Client) ReportUserTraffic(userTraffic []UserTraffic) error {
	data := make(map[int][]int64, len(userTraffic))
	for i := range userTraffic {
		data[userTraffic[i].UID] = []int64{userTraffic[i].Upload, userTraffic[i].Download}
	}
	const path = "/api/v1/server/UniProxy/push"
	res, err := c.client.R().
		SetBody(data).
		ForceContentType("application/json").
		Post(path)
	err = c.checkResponse(res, path, err)
	if err != nil {
		return err
	}
	return nil
}
