package main

import (
	"go.etcd.io/etcd/raft"
	"go.etcd.io/etcd/raft/raftpb"
	"log"
	"sync"
	"time"
)

type node struct {
	raft.Node
	id      uint64
	storage *raft.MemoryStorage

	mu    sync.Mutex
	state raftpb.HardState
}

func main() {
	st := raft.NewMemoryStorage()
	c := &raft.Config{
		ID:              0x01,
		ElectionTick:    10,
		HeartbeatTick:   1,
		Storage:         st,
		MaxSizePerMsg:   4096,
		MaxInflightMsgs: 256,
	}
	peers := []raft.Peer{{ID: 0x01}}
	rn := raft.StartNode(c, peers)
	n := node{
		Node:    rn,
		id:      0x01,
		storage: st,
	}

	ticker := time.Tick(100 * time.Millisecond)
	for {
		select {
		case <-ticker:
			n.Tick()
			log.Println("=== Tick ===")
		case rd := <-n.Ready():
			if !raft.IsEmptyHardState(rd.HardState) {
				n.mu.Lock()
				n.state = rd.HardState
				n.mu.Unlock()
				n.storage.SetHardState(n.state)
			}
			n.storage.Append(rd.Entries)
			n.Advance()
			log.Printf("=== len(rd.Entries) = %d ===\n",len(rd.Entries))
		}
	}

}
