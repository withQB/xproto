package simulator

type APIEventMessageID int
type APICommandMessageID int
type APIUpdateID int
type APICommandID int
type APINodeType int

const (
	UnknownEventMsg APIEventMessageID = iota
	SimInitialState
	SimStateUpdate
)

const (
	UnknownCommandMsg APICommandMessageID = iota
	SimPlaySequence
)

const (
	UnknownUpdate APIUpdateID = iota
	SimNodeAdded
	SimNodeRemoved
	SimPeerAdded
	SimPeerRemoved
	SimTreeParentUpdated
	SimSnakeAscUpdated
	SimSnakeDescUpdated
	SimTreeRootAnnUpdated
	SimSnakeEntryAdded
	SimSnakeEntryRemoved
	SimPingStateUpdated
	SimNetworkStatsUpdated
	SimBroadcastReceived
	SimBandwidthReport
)

const (
	UnknownCommand APICommandID = iota
	SimDebug
	SimPlay
	SimPause
	SimDelay
	SimAddNode
	SimRemoveNode
	SimAddPeer
	SimRemovePeer
	SimConfigureAdversaryDefaults
	SimConfigureAdversaryPeer
	SimStartPings
	SimStopPings
)

const (
	UnknownType APINodeType = iota
	DefaultNode
	GeneralAdversaryNode
)

type InitialNodeState struct {
	PublicKey          string
	NodeType           APINodeType
	RootState          RootState
	Peers              []PeerInfo
	TreeParent         string
	SnakeAsc           string
	SnakeAscPath       string
	SnakeDesc          string
	SnakeDescPath      string
	SnakeEntries       []SnakeRouteEntry
	BroadcastsReceived []BroadcastEntry
	BandwidthReports   []BandwidthSnapshot
}

type RootState struct {
	Root        string
	AnnSequence uint64
	AnnTime     uint64
	Coords      []uint64
}

type PeerInfo struct {
	ID   string
	Port int
}

type SnakeRouteEntry struct {
	EntryID string
	PeerID  string
}

type BroadcastEntry struct {
	PeerID string
	Time   uint64
}

type SimEventMsg struct {
	UpdateID APIUpdateID
	Event    SimEvent
}

type InitialStateMsg struct {
	MsgID               APIEventMessageID
	Nodes               map[string]InitialNodeState
	End                 bool
	BWReportingInterval int
}

type StateUpdateMsg struct {
	MsgID APIEventMessageID
	Event SimEventMsg
}

type SimCommandSequenceMsg struct {
	MsgID  APICommandMessageID
	Events []SimCommandMsg
}

type SimCommandMsg struct {
	MsgID APICommandID
	Event interface{}
}
