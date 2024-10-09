package p2p

type Message struct {
	Payload any
	From    string
}

type BroadcastTo struct {
	To      []string
	Payload any
}

func NewMessage(from string, payload any) *Message {
	return &Message{
		From:    from,
		Payload: payload,
	}
}

type Handshake struct {
	Version     string
	GameVariant GameVariant
	GameStatus  GameStatus
	ListenAddr  string
}

type MessagePlayerAction struct {
	CurrentGameStatus GameStatus
	Action            PlayerAction
	Value             int
}

type MessagePreFlop struct{}

func (msg *MessagePreFlop) string() string {
	return "MSG: PRE FLOP"
}

type MessagePeerList struct {
	Peers []string
}

type MessageEncDeck struct {
	Deck [][]byte
}

type MessageReady struct{}

func (msg *MessageReady) string() string {
	return "MSG: READY"
}
