package panel

import (
	"fmt"
	"github.com/goccy/go-json"
	"strconv"
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
	/*DeviceLimit int             `json:"device_limit"`
	SpeedLimit  uint64          `json:"speed_limit"`*/
	UID        int             `json:"id"`
	Traffic    int64           `json:"-"`
	Port       int             `json:"port"`
	Cipher     string          `json:"cipher"`
	Secret     string          `json:"secret"`
	V2rayUser  *V2RayUserInfo  `json:"v2ray_user"`
	TrojanUser *TrojanUserInfo `json:"trojan_user"`
}

func (p *UserInfo) GetUserEmail() string {
	if p.V2rayUser != nil {
		return p.V2rayUser.Email
	} else if p.TrojanUser != nil {
		return p.TrojanUser.Password
	}
	return p.Cipher
}

type UserListBody struct {
	//Msg  string `json:"msg"`
	Data []UserInfo `json:"data"`
}

// GetUserList will pull user form sspanel
func (c *Client) GetUserList() (UserList []UserInfo, err error) {
	var path string
	switch c.NodeType {
	case "V2ray":
		path = "/api/v1/server/Deepbwork/user"
	case "Trojan":
		path = "/api/v1/server/TrojanTidalab/user"
	case "Shadowsocks":
		path = "/api/v1/server/ShadowsocksTidalab/user"
	default:
		return nil, fmt.Errorf("unsupported Node type: %s", c.NodeType)
	}
	res, err := c.client.R().
		ForceContentType("application/json").
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
	return userList.Data, nil
}

type UserTraffic struct {
	UID      int   `json:"user_id"`
	Upload   int64 `json:"u"`
	Download int64 `json:"d"`
}

// ReportUserTraffic reports the user traffic
func (c *Client) ReportUserTraffic(userTraffic []UserTraffic) error {
	var path string
	switch c.NodeType {
	case "V2ray":
		path = "/api/v1/server/Deepbwork/submit"
	case "Trojan":
		path = "/api/v1/server/TrojanTidalab/submit"
	case "Shadowsocks":
		path = "/api/v1/server/ShadowsocksTidalab/submit"
	}

	data := make([]UserTraffic, len(userTraffic))
	for i, traffic := range userTraffic {
		data[i] = UserTraffic{
			UID:      traffic.UID,
			Upload:   traffic.Upload,
			Download: traffic.Download}
	}

	res, err := c.client.R().
		SetQueryParam("node_id", strconv.Itoa(c.NodeID)).
		SetBody(data).
		ForceContentType("application/json").
		Post(path)
	err = c.checkResponse(res, path, err)
	if err != nil {
		return err
	}
	return nil
}
