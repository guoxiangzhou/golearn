package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"go.etcd.io/etcd/raft"
	"go.etcd.io/etcd/raft/raftpb"
	"log"
	"sync"
	"time"
)

type node struct {
	raft.Node
	storage *raft.MemoryStorage

	mu      sync.Mutex
	kvstore map[string]string
}

type kv struct {
	Key   string
	Value string
}

func (n *node) applyCommits(entries []raftpb.Entry) {
	if len(entries) == 0 {
		return
	}
	firstIdx := entries[0].Index
	if firstIdx > n.Status().Applied+1 {
		log.Fatalf("first index %d should <= (appliedIndex %d) + 1\n", firstIdx, n.Status().Applied)
	}
	if n.Status().Applied-firstIdx+1 < uint64(len(entries)) {
		entries := entries[n.Status().Applied-firstIdx+1:]
		for _, entry := range entries {
			if entry.Type == raftpb.EntryNormal && len(entry.Data) > 0 {
				var datakv kv
				dec := gob.NewDecoder(bytes.NewBuffer(entry.Data))
				if err := dec.Decode(&datakv); err != nil {
					log.Fatal(err)
				}
				n.mu.Lock()
				n.kvstore[datakv.Key] = datakv.Value
				n.mu.Unlock()
			}
		}
	}
}

func (n *node) propose(key string, value string) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(kv{key, value}); err != nil {
		log.Fatal(err)
	}
	n.Propose(context.TODO(), buf.Bytes())
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
		kvstore: make(map[string]string),
	}

	stopc := make(chan struct{})

	go func() {
		ticker := time.Tick(10 * time.Millisecond)
		for {
			select {
			case <-ticker:
				n.Tick()
			case rd := <-n.Ready():
				if !raft.IsEmptyHardState(rd.HardState) {
					n.storage.SetHardState(rd.HardState)
				}
				n.storage.Append(rd.Entries)
				n.applyCommits(rd.CommittedEntries)
				n.Advance()
			case <-stopc:
				n.Stop()
				return
			}
		}
	}()
	for i := 0; i < 10; i++ {
		n.propose(fmt.Sprintf("Key%d", i), fmt.Sprintf("Value%d", i))
		if i == 9 {
			stopc <- struct{}{}
		}
		time.Sleep(100 * time.Millisecond)
	}
	for key, value := range n.kvstore {
		log.Printf("%s,%s", key, value)
	}
}
