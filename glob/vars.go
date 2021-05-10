package glob

import "time"

// Vars which may change during runtime
//

var COMM_ID_LENGTH = 16
var MSG_ID_LENGTH = 16
var NODE_ID_LENGTH = 16
var NODE_TO_NODE_RES_TIMEOUT time.Duration = 2000 // milliseconds

var F_CLIENT_LISTEN = ""
var F_NODE_LISTEN = ""
var F_NODE_CONNECT = make([]string, 0)
var F_LOG_IO_MSG = false
var F_CONN_RC_CLEANUP_INTERVAL time.Duration = 2500 // milliseconds

// Memory = keys length + data bytes length in Megabytes
var F_MAX_USED_MEMORY int64 = 0
var F_MAX_CACHED_KEYS int64 = 0
