package chat

// senderType represents the type of sender of a message in the chat
type senderType uint

const (
	senderTypeAssistant senderType = iota
	senderTypeSystem
	senderTypeTool
	senderTypeToolCaller
	senderTypeUser
)

// Sender represents a sender of a message in the chat
type Sender interface {
	GetType() senderType
}

// senderBase is a base type for simple sender types
type senderBase struct{}

// SenderAssistant represents an assistant sender
type SenderAssistant senderBase

func (s SenderAssistant) GetType() senderType { return senderTypeAssistant }

// SenderSystem represents a system sender
type SenderSystem senderBase

func (s SenderSystem) GetType() senderType { return senderTypeSystem }

// SenderUser represents a user sender
type SenderUser senderBase

func (s SenderUser) GetType() senderType { return senderTypeUser }

// SenderTool represents a tool sender
type SenderTool senderBase

func (s SenderTool) GetType() senderType { return senderTypeTool }

// SenderToolCall represents an assistant sender's tool call
type SenderToolCaller struct {
	// Name is the name of the tool
	Name   string
	CallId string
}

func (s SenderToolCaller) GetType() senderType { return senderTypeToolCaller }
