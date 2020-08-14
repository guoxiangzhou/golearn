package main

import (
	"go.etcd.io/etcd/raft"
	"time"
)

func main() {
	storage := raft.NewMemoryStorage()
	config := &raft.Config{
		ID:              0x01,
		ElectionTick:    10,
		HeartbeatTick:   1,
		Storage:         storage,
		MaxSizePerMsg:   4096,
		MaxInflightMsgs: 256,
	}
	raftNode := raft.StartNode(config, []raft.Peer{{ID: 0x01}})
	stopCh := make(chan struct{})
	go func() {
		ticker := time.Tick(10 * time.Millisecond)
		for {
			select {
			case <-ticker:
				raftNode.Tick()
			case rd := <-raftNode.Ready():
				if !raft.IsEmptyHardState(rd.HardState) {
					storage.SetHardState(rd.HardState)
				}
				storage.Append(rd.Entries)
				raftNode.Advance()
			case <-stopCh:
				raftNode.Stop()
			}
		}
	}()
	for i := 0; i < 10; i++ {
		if i == 9 {
			stopCh <- struct{}{}
		}
		time.Sleep(100 * time.Millisecond)
	}
}
