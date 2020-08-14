package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/raft"
	"go.etcd.io/etcd/raft/raftpb"
	"log"
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
				for _, ent := range rd.CommittedEntries {
					if ent.Type == raftpb.EntryNormal && len(ent.Data) > 0 {
						msg := string(ent.Data)
						log.Printf("commit msg : %s\n", msg)
					}
				}
				raftNode.Advance()
			case <-stopCh:
				raftNode.Stop()
			}
		}
	}()
	for i := 0; i < 10; i++ {
		if i == 9 {
			stopCh <- struct{}{}
		} else {
			raftNode.Propose(context.TODO(), []byte(fmt.Sprintf("value %d", i)))
		}
		time.Sleep(100 * time.Millisecond)
	}
}
