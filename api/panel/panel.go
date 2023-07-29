package panel

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/InazumaV/V2bX/conf"
	"github.com/go-resty/resty/v2"
)

// Panel is the interface for different panel's api.

type Client struct {
	client        *resty.Client
	APIHost       string
	Token         string
	NodeType      string
	NodeId        int
	LocalRuleList []*regexp.Regexp
	nodeEtag      string
	userEtag      string
}

func New(c *conf.ApiConfig) (*Client, error) {
	client := resty.New()
	client.SetRetryCount(3)
	if c.Timeout > 0 {
		client.SetTimeout(time.Duration(c.Timeout) * time.Second)
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
	client.SetBaseURL(c.APIHost)
	// Check node type
	c.NodeType = strings.ToLower(c.NodeType)
	switch c.NodeType {
	case "v2ray", "trojan", "shadowsocks", "hysteria":
	default:
		return nil, fmt.Errorf("unsupported Node type: %s", c.NodeType)
	}
	// set params
	client.SetQueryParams(map[string]string{
		"node_type": c.NodeType,
		"node_id":   strconv.Itoa(c.NodeID),
		"token":     c.Key,
	})
	// Read local rule list
	localRuleList := readLocalRuleList(c.RuleListPath)
	return &Client{
		client:        client,
		Token:         c.Key,
		APIHost:       c.APIHost,
		NodeType:      c.NodeType,
		NodeId:        c.NodeID,
		LocalRuleList: localRuleList,
	}, nil
}

// readLocalRuleList reads the local rule list file
func readLocalRuleList(path string) (LocalRuleList []*regexp.Regexp) {
	LocalRuleList = make([]*regexp.Regexp, 0)
	if path != "" {
		// open the file
		file, err := os.Open(path)
		//handle errors while opening
		if err != nil {
			log.Printf("Error when opening file: %s", err)
			return
		}
		fileScanner := bufio.NewScanner(file)
		// read line by line
		for fileScanner.Scan() {
			LocalRuleList = append(LocalRuleList, regexp.MustCompile(fileScanner.Text()))
		}
		// handle first encountered error while reading
		if err := fileScanner.Err(); err != nil {
			log.Fatalf("Error while reading file: %s", err)
			return
		}
	}
	return
}
