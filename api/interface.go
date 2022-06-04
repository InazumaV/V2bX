package api

type API interface {
	GetNodeInfo() (nodeInfo *NodeInfo, err error)
	GetUserList() (userList *[]UserInfo, err error)
	ReportUserTraffic(userTraffic *[]UserTraffic) (err error)
	Describe() ClientInfo
	GetNodeRule() (ruleList *[]DetectRule, err error)
	Debug()
}
