package mongodb

import "time"

const (
	// time
	DbConnTimeout      = 30 * time.Second
	DbOperationTimeout = 15 * time.Second
	DbClientWait       = 3 * time.Second
	DbPingDelay        = 30 * time.Second
	DbPingTimeout      = 5 * time.Second
	// log
	Component     = "dbservice"
	ActionConnect = "connect"
	ActionClose   = "close"
	ActionColl    = "collection"
	ActionIndex   = "index"
	ActionPing    = "ping"
	DbAdminHexID  = "63d121d10cbf3c227c7ce765"
)
