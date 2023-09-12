package events

import (
	"github.com/withqb/xproto/types"
)

type Event interface {
	isEvent()
}

type PeerAdded struct {
	Port   types.SwitchPortID
	PeerID string
}

// Tag PeerAdded as an Event
func (e PeerAdded) isEvent() {}

type PeerRemoved struct {
	Port   types.SwitchPortID
	PeerID string
}

// Tag PeerRemoved as an Event
func (e PeerRemoved) isEvent() {}

type TreeParentUpdate struct {
	PeerID string
}

// Tag TreeParentUpdate as an Event
func (e TreeParentUpdate) isEvent() {}

type SnakeDescUpdate struct {
	PeerID string
}

// Tag SnakeDescUpdate as an Event
func (e SnakeDescUpdate) isEvent() {}

type TreeRootAnnUpdate struct {
	Root     string // Root Public Key
	Sequence uint64
	Time     uint64 // Unix Time
	Coords   []uint64
}

// Tag TreeRootAnnUpdate as an Event
func (e TreeRootAnnUpdate) isEvent() {}

type SnakeEntryAdded struct {
	EntryID string
	PeerID  string
}

// Tag SnakeEntryAdded as an Event
func (e SnakeEntryAdded) isEvent() {}

type SnakeEntryRemoved struct {
	EntryID string
}

// Tag SnakeEntryRemoved as an Event
func (e SnakeEntryRemoved) isEvent() {}

type BroadcastReceived struct {
	PeerID string
	Time   uint64
}

// Tag BroadcastReceived as an Event
func (e BroadcastReceived) isEvent() {}

type PeerBandwidthUsage struct {
	Protocol struct {
		Rx uint64
		Tx uint64
	}
	Overlay struct {
		Rx uint64
		Tx uint64
	}
}

type BandwidthReport struct {
	CaptureTime uint64 // Unix Time
	Peers       map[string]PeerBandwidthUsage
}

// Tag BandwidthReport as an Event
func (e BandwidthReport) isEvent() {}
