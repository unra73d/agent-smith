package agent

type SSEMessageType string

const (
	SSEMessageSessionUpdate      = "session_update"
	SSEMessageNewMessage         = "new_message"
	SSEMessageLastMessageUpdate  = "last_message_update"
	SSEMessageProviderListUpdate = "provider_list_update"
	SSEMessageMCPListUpdate      = "mcp_list_update"
	SSEMessageRoleListUpdate     = "role_list_update"
)

type SSEMessage struct {
	Type SSEMessageType
	Data any
}

var sseCh chan *SSEMessage = make(chan *SSEMessage)

func NewSSEConnection() chan *SSEMessage {
	return sseCh
}
