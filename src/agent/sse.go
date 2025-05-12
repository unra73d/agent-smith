package agent

type SSEMessageType string

const (
	SSEMessageSessionListUpdate = "session_list_update"
	SSEMessageSessionUpdate     = "session_update"
	SSEMessageNewMessage        = "new_message"
	SSEMessageLastMessageUpdate = "last_message_update"
	SSEMessageModelListUpdate   = "model_list_update"
	SSEMessageMCPListUpdate     = "mcp_list_update"
)

type SSEMessage struct {
	Type SSEMessageType
	Data any
}

var sseCh chan *SSEMessage = make(chan *SSEMessage)

func NewSSEConnection() chan *SSEMessage {
	return sseCh
}
