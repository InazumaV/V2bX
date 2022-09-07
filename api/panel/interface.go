package panel

type Panel interface {
	GetNodeInfo() (nodeInfo *NodeInfo, err error)
	GetUserList() (userList []UserInfo, err error)
	ReportUserTraffic(userTraffic []UserTraffic) (err error)
	Describe() ClientInfo
	GetNodeRule() (ruleList []DetectRule, protocolList []string, err error)
	Debug()
}
