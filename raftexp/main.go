package main

import (
	"go.etcd.io/etcd/raft"
	"go.etcd.io/etcd/raft/raftpb"
	"log"
	"time"
)

type raftNode struct {
	node         raft.Node
	appliedIndex uint64
	confState     raftpb.ConfState
}

func (r *raftNode) pushCommitEntries(e []raftpb.Entry) {
	if len(e) == 0 {
		return
	}
	firstIdx := e[0].Index
	if firstIdx > r.appliedIndex+1 {
		log.Fatalf("first index of committed entry[%d] should <= progress.appliedIndex[%d]+1", firstIdx, r.appliedIndex)
	}
	var pe []raftpb.Entry
	if r.appliedIndex+1-firstIdx < uint64(len(e)) {
		pe = e[r.appliedIndex+1-firstIdx : 1]
	}
	for _, ee := range pe {
		switch ee.Type {
		case raftpb.EntryNormal:
			log.Printf("pushCommitEntries, normal Entry")
		case raftpb.EntryConfChange:
			var cc raftpb.ConfChange
			cc.Unmarshal(ee.Data)
			r.confState = *r.node.ApplyConfChange(cc)
			switch cc.Type {
			case raftpb.ConfChangeAddNode:
				log.Printf("conf change add node")
			case raftpb.ConfChangeRemoveNode:
				log.Printf("conf change remove node")
			}
			log.Printf("pushCommitEntries conf change entry")
		}
		r.appliedIndex = ee.Index
	}
}

func main() {
	node := raftNode{
		node: nil,
	}
	storage := raft.NewMemoryStorage()
	c := &raft.Config{
		ID:              0x01,
		ElectionTick:    10,
		HeartbeatTick:   1,
		Storage:         storage,
		MaxSizePerMsg:   4096,
		MaxInflightMsgs: 256,
	}
	peers := []raft.Peer{{ID: 0x01}}
	node.node = raft.StartNode(c, peers)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			log.Printf("Tick...")
			node.node.Tick()
		case rd := <-node.node.Ready():
			log.Printf("HardState = %v\n", rd.HardState)
			log.Printf("Entries = %v\n", rd.Entries)
			log.Printf("SnapShot = %v\n", rd.Snapshot)
			log.Printf("Message = %v\n", rd.Messages)
			log.Printf("CommitEntries, len = %d, value =%v\n", len(rd.CommittedEntries), rd.CommittedEntries)
			node.pushCommitEntries(rd.CommittedEntries)
			node.node.Advance()
		}
	}
}
