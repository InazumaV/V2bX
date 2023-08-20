package panel

import (
	"fmt"
	log "github.com/sirupsen/logrus"

	"github.com/goccy/go-json"
)

type OnlineUser struct {
	UID int
	IP  string
}

type UserInfo struct {
	Id         int    `json:"id"`
	Uuid       string `json:"uuid"`
	SpeedLimit int    `json:"speed_limit"`
}

type UserListBody struct {
	//Msg  string `json:"msg"`
	Users []UserInfo `json:"users"`
}

// GetUserList will pull user form sspanel
func (c *Client) GetUserList() (UserList []UserInfo, err error) {
	const path = "/api/v1/server/UniProxy/user"
	r, err := c.client.R().
		SetHeader("If-None-Match", c.userEtag).
		Get(path)
	err = c.checkResponse(r, path, err)
	if err != nil {
		return nil, err
	}
	err = c.checkResponse(r, path, err)
	if r.StatusCode() == 304 {
		return nil, nil
	}
	var userList *UserListBody
	err = json.Unmarshal(r.Body(), &userList)
	if err != nil {
		return nil, fmt.Errorf("unmarshal userlist error: %s", err)
	}
	c.userEtag = r.Header().Get("ETag")
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
	r, err := c.client.R().
		SetBody(data).
		ForceContentType("application/json").
		Post(path)
	err = c.checkResponse(r, path, err)
	if err != nil {
		return err
	}
	log.Println(r.String())
	return nil
}
