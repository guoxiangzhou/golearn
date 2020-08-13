package main

import (
	"go.etcd.io/etcd/raft"
	"go.etcd.io/etcd/raft/raftpb"
	"sync"
	"time"
)

type node struct {
	raft.Node
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

	rn := raft.StartNode(c, []raft.Peer{{ID: 0x01}})
	n := &node{
		Node:    rn,
		storage: st,
	}

	stopc := make(chan struct{})

	go func() {
		ticker := time.Tick(100 * time.Millisecond)
		for {
			select {
			case <-ticker:
				n.Tick()
			case rd := <-n.Ready():
				if !raft.IsEmptyHardState(rd.HardState) {
					n.mu.Lock()
					n.state = rd.HardState
					n.mu.Unlock()
					n.storage.SetHardState(n.state)
				}
				n.storage.Append(rd.Entries)
				n.Advance()
			case <-stopc:
				n.Stop()
				return
			}
		}
	}()
	for i := 0; i < 10; i++ {
		if i == 9 {
			stopc <- struct{}{}
		}
		time.Sleep(time.Second)
	}
}
