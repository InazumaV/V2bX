package v2board

import (
	"bufio"
	md52 "crypto/md5"
	"fmt"
	"github.com/Yuzuki616/V2bX/api"
	"github.com/go-resty/resty/v2"
	"github.com/goccy/go-json"
	"log"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"
)

// APIClient create an api client to the panel.
type APIClient struct {
	client           *resty.Client
	APIHost          string
	NodeID           int
	Key              string
	NodeType         string
	EnableSS2022     bool
	EnableVless      bool
	EnableXTLS       bool
	SpeedLimit       float64
	DeviceLimit      int
	LocalRuleList    []api.DetectRule
	RemoteRuleCache  *api.Rule
	access           sync.Mutex
	NodeInfoRspMd5   [16]byte
	UserListCheckNum int
}

// New create an api instance
func New(apiConfig *api.Config) *APIClient {

	client := resty.New()
	client.SetRetryCount(3)
	if apiConfig.Timeout > 0 {
		client.SetTimeout(time.Duration(apiConfig.Timeout) * time.Second)
	} else {
		client.SetTimeout(5 * time.Second)
	}
	client.OnError(func(req *resty.Request, err error) {
		if v, ok := err.(*resty.ResponseError); ok {
			// v.Response contains the last response from the server
			// v.Err contains the original error
			log.Print(v.Err)
		}
	})
	client.SetBaseURL(apiConfig.APIHost)
	// Create Key for each requests
	client.SetQueryParams(map[string]string{
		"node_id": strconv.Itoa(apiConfig.NodeID),
		"token":   apiConfig.Key,
	})
	// Read local rule list
	localRuleList := readLocalRuleList(apiConfig.RuleListPath)
	apiClient := &APIClient{
		client:        client,
		NodeID:        apiConfig.NodeID,
		Key:           apiConfig.Key,
		APIHost:       apiConfig.APIHost,
		NodeType:      apiConfig.NodeType,
		EnableSS2022:  apiConfig.EnableSS2022,
		EnableVless:   apiConfig.EnableVless,
		EnableXTLS:    apiConfig.EnableXTLS,
		SpeedLimit:    apiConfig.SpeedLimit,
		DeviceLimit:   apiConfig.DeviceLimit,
		LocalRuleList: localRuleList,
	}
	return apiClient
}

// readLocalRuleList reads the local rule list file
func readLocalRuleList(path string) (LocalRuleList []api.DetectRule) {

	LocalRuleList = make([]api.DetectRule, 0)
	if path != "" {
		// open the file
		file, err := os.Open(path)

		//handle errors while opening
		if err != nil {
			log.Printf("Error when opening file: %s", err)
			return LocalRuleList
		}

		fileScanner := bufio.NewScanner(file)

		// read line by line
		for fileScanner.Scan() {
			LocalRuleList = append(LocalRuleList, api.DetectRule{
				ID:      -1,
				Pattern: regexp.MustCompile(fileScanner.Text()),
			})
		}
		// handle first encountered error while reading
		if err := fileScanner.Err(); err != nil {
			log.Fatalf("Error while reading file: %s", err)
			return []api.DetectRule{}
		}

		file.Close()
	}

	return LocalRuleList
}

// Describe return a description of the client
func (c *APIClient) Describe() api.ClientInfo {
	return api.ClientInfo{APIHost: c.APIHost, NodeID: c.NodeID, Key: c.Key, NodeType: c.NodeType}
}

// Debug set the client debug for client
func (c *APIClient) Debug() {
	c.client.SetDebug(true)
}

func (c *APIClient) assembleURL(path string) string {
	return c.APIHost + path
}

func (c *APIClient) checkResponse(res *resty.Response, path string, err error) error {
	if err != nil {
		return fmt.Errorf("request %s failed: %s", c.assembleURL(path), err)
	}

	if res.StatusCode() > 400 {
		body := res.Body()
		return fmt.Errorf("request %s failed: %s, %s", c.assembleURL(path), string(body), err)
	}
	return nil
}

// GetNodeInfo will pull NodeInfo Config from sspanel
func (c *APIClient) GetNodeInfo() (nodeInfo *api.NodeInfo, err error) {
	var path string
	var res *resty.Response
	switch c.NodeType {
	case "V2ray":
		path = "/api/v1/server/Deepbwork/config"
		res, err = c.client.R().
			SetQueryParam("local_port", "1").
			ForceContentType("application/json").
			Get(path)
	case "Trojan":
		path = "/api/v1/server/TrojanTidalab/config"
	case "Shadowsocks":
		if nodeInfo, err = c.ParseSSNodeResponse(); err == nil {
			return nodeInfo, nil
		} else {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported Node type: %s", c.NodeType)
	}
	md := md52.Sum(res.Body())
	if c.NodeInfoRspMd5 != [16]byte{} {
		if c.NodeInfoRspMd5 == md {
			return nil, nil
		}
	}
	c.NodeInfoRspMd5 = md
	res, err = c.client.R().
		SetQueryParam("local_port", "1").
		ForceContentType("application/json").
		Get(path)
	err = c.checkResponse(res, path, err)
	if err != nil {
		return nil, err
	}
	c.access.Lock()
	defer c.access.Unlock()
	switch c.NodeType {
	case "V2ray":
		nodeInfo, err = c.ParseV2rayNodeResponse(res.Body())
	case "Trojan":
		nodeInfo, err = c.ParseTrojanNodeResponse(res.Body())
	}
	return nodeInfo, nil
}

// GetUserList will pull user form sspanel
func (c *APIClient) GetUserList() (UserList *[]api.UserInfo, err error) {
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
	var userList *api.UserListBody
	err = json.Unmarshal(res.Body(), &userList)
	if err != nil {
		return nil, fmt.Errorf("unmarshal userlist error: %s", err)
	}
	checkNum := userList.Data[len(userList.Data)-1].UID +
		userList.Data[len(userList.Data)/2-1].UID +
		userList.Data[0].UID
	if c.UserListCheckNum != 0 {
		if c.UserListCheckNum == checkNum {
			return nil, nil
		}
	}
	c.UserListCheckNum = userList.Data[len(userList.Data)-1].UID
	return &userList.Data, nil
}

// ReportUserTraffic reports the user traffic
func (c *APIClient) ReportUserTraffic(userTraffic *[]api.UserTraffic) error {
	var path string
	switch c.NodeType {
	case "V2ray":
		path = "/api/v1/server/Deepbwork/submit"
	case "Trojan":
		path = "/api/v1/server/TrojanTidalab/submit"
	case "Shadowsocks":
		path = "/api/v1/server/ShadowsocksTidalab/submit"
	}

	data := make([]UserTraffic, len(*userTraffic))
	for i, traffic := range *userTraffic {
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

// GetNodeRule implements the API interface
func (c *APIClient) GetNodeRule() (*[]api.DetectRule, error) {
	ruleList := c.LocalRuleList
	if c.NodeType != "V2ray" {
		return &ruleList, nil
	}

	// V2board only support the rule for v2ray
	// fix: reuse config response
	c.access.Lock()
	defer c.access.Unlock()
	for i, rule := range c.RemoteRuleCache.Domain {
		ruleListItem := api.DetectRule{
			ID:      i,
			Pattern: regexp.MustCompile(rule),
		}
		ruleList = append(ruleList, ruleListItem)
	}
	return &ruleList, nil
}

// ParseTrojanNodeResponse parse the response for the given nodeinfor format
func (c *APIClient) ParseTrojanNodeResponse(body []byte) (*api.NodeInfo, error) {
	node := &api.NodeInfo{Trojan: &api.TrojanConfig{}}
	err := json.Unmarshal(body, node.Trojan)
	if err != nil {
		return nil, fmt.Errorf("unmarshal nodeinfo error: %s", err)
	}
	node.NodeId = c.NodeID
	node.NodeType = c.NodeType
	return node, nil
}

// ParseSSNodeResponse parse the response for the given nodeinfor format
func (c *APIClient) ParseSSNodeResponse() (*api.NodeInfo, error) {
	var port int
	var method string
	userInfo, err := c.GetUserList()
	if err != nil {
		return nil, err
	}
	if len(*userInfo) > 0 {
		port = (*userInfo)[0].Port
		method = (*userInfo)[0].Cipher
	}

	if err != nil {
		return nil, err
	}
	node := &api.NodeInfo{
		EnableSS2022: c.EnableSS2022,
		NodeType:     c.NodeType,
		NodeId:       c.NodeID,
		SS: &api.SSConfig{
			Port:              port,
			TransportProtocol: "tcp",
			CypherMethod:      method,
		},
	}
	return node, nil
}

// ParseV2rayNodeResponse parse the response for the given nodeinfor format
func (c *APIClient) ParseV2rayNodeResponse(body []byte) (*api.NodeInfo, error) {
	node := &api.NodeInfo{V2ray: &api.V2rayConfig{}}
	err := json.Unmarshal(body, node.V2ray)
	if err != nil {
		return nil, fmt.Errorf("unmarshal nodeinfo error: %s", err)
	}
	node.NodeType = c.NodeType
	node.NodeId = c.NodeID
	c.RemoteRuleCache = &node.V2ray.Routing.Rules[0]
	node.V2ray.Routing = nil
	if c.EnableXTLS {
		node.TLSType = "xtls"
	} else {
		node.TLSType = "tls"
	}
	node.EnableVless = c.EnableVless
	node.EnableTls = node.V2ray.Inbounds[0].StreamSetting.Security == "tls"
	return node, nil
}
