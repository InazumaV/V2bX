package panel

import (
	"github.com/Yuzuki616/V2bX/conf"
	"github.com/go-resty/resty/v2"
	"log"
	"strconv"
	"sync"
	"time"
)

// Panel is the interface for different panel's api.

type ClientInfo struct {
	APIHost  string
	NodeID   int
	Key      string
	NodeType string
}

type Client struct {
	client   *resty.Client
	APIHost  string
	NodeID   int
	Key      string
	NodeType string
	//EnableSS2022     bool
	EnableVless     bool
	EnableXTLS      bool
	SpeedLimit      float64
	DeviceLimit     int
	LocalRuleList   []DetectRule
	RemoteRuleCache *[]Rule
	access          sync.Mutex
	NodeInfoRspMd5  [16]byte
	NodeRuleRspMd5  [16]byte
}

func New(apiConfig *conf.ApiConfig) Panel {
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
	return &Client{
		client:   client,
		NodeID:   apiConfig.NodeID,
		Key:      apiConfig.Key,
		APIHost:  apiConfig.APIHost,
		NodeType: apiConfig.NodeType,
		//EnableSS2022:  apiConfig.EnableSS2022,
		EnableVless:   apiConfig.EnableVless,
		EnableXTLS:    apiConfig.EnableXTLS,
		SpeedLimit:    apiConfig.SpeedLimit,
		DeviceLimit:   apiConfig.DeviceLimit,
		LocalRuleList: localRuleList,
	}
}
