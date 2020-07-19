package main

import (
	"go.etcd.io/etcd/raft"
	"go.etcd.io/etcd/raft/raftpb"
	"log"
	"time"
)

func pushCommitEntries(e []raftpb.Entry) {
	if len(e) == 0{
		return
	}

}

func main() {
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
	n := raft.StartNode(c, peers)
	ticker := time.NewTicker(100 * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			log.Printf("Tick...")
			n.Tick()
		case rd := <-n.Ready():
			log.Printf("HardState = %v\n", rd.HardState)
			log.Printf("Entries = %v\n", rd.Entries)
			log.Printf("SnapShot = %v\n", rd.Snapshot)
			log.Printf("Message = %v\n", rd.Messages)
			log.Printf("CommitEntries, len = %d, value =%v\n", len(rd.CommittedEntries), rd.CommittedEntries)
			pushCommitEntries(rd.CommittedEntries)
			n.Advance()
		}
	}
}
