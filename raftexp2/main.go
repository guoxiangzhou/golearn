package main

import (
	"context"
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

	mu    sync.Mutex
	state raftpb.HardState

	mbox chan raftpb.Message
}

func buildNode(id uint64, peers []raft.Peer) *node {
	st := raft.NewMemoryStorage()
	c := &raft.Config{
		ID:              id,
		ElectionTick:    10,
		HeartbeatTick:   1,
		Storage:         st,
		MaxSizePerMsg:   4096,
		MaxInflightMsgs: 256,
	}
	rn := raft.StartNode(c, peers)
	n := &node{
		Node:    rn,
		storage: st,
	}
	return n
}

func sendMessages(src *node, nodes []*node, msgs []raftpb.Message) {
	if len(msgs) == 0 {
		return
	}

	//var buffer bytes.Buffer
	//for _, m := range msgs {
	//	buffer.WriteString(fmt.Sprintf(" %v %d->%d;", m.Type, m.From, m.To))
	//}
	//log.Printf("src:%d, msg : %s\n", src.Status().ID, buffer.String())

	for _, m := range msgs {
		b, err := m.Marshal()
		if err != nil {
			log.Fatal(err)
		}
		var cm raftpb.Message
		err = cm.Unmarshal(b)
		if err != nil {
			log.Fatal(err)
		}
		toIdx := m.To - 1
		go func() {
			nodes[toIdx].mbox <- cm
		}()
	}
}

func applyCommits(n *node, entries []raftpb.Entry) {
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
			if entry.Type == raftpb.EntryConfChange {
				var cc raftpb.ConfChange
				cc.Unmarshal(entry.Data)
				n.ApplyConfChange(cc)
			} else if entry.Type == raftpb.EntryNormal && n.Status().ID == 1 {
				if len(entry.Data) != 0 {
					msg := string(entry.Data)
					log.Printf("apply %d msg : %s", n.Status().ID, msg)
				}
			}
		}
	}
}

func proposeEntries(n *node, entries []raftpb.Entry) {
	if n.Status().ID == 1 {
		for _, e := range entries {
			if e.Type == raftpb.EntryNormal {
				if len(e.Data) != 0 {
					msg := string(e.Data)
					log.Printf("propose %d msg : %s", n.Status().ID, msg)
				}
			}
		}
	}
}

func startNodes(nodes []*node) {
	for _, n := range nodes {
		n.mbox = make(chan raftpb.Message, len(nodes))
	}

	for _, n := range nodes {
		n := n
		go func() {
			ticker := time.Tick(time.Second)
			for {
				select {
				case <-ticker:
					n.Tick()
				case rd := <-n.Ready():
					if n.Status().ID==1{
						log.Println("==== node ready ===========")
					}
					if !raft.IsEmptyHardState(rd.HardState) {
						n.mu.Lock()
						n.state = rd.HardState
						n.mu.Unlock()
						n.storage.SetHardState(n.state)
					}
					proposeEntries(n, rd.Entries)
					n.storage.Append(rd.Entries)
					sendMessages(n, nodes, rd.Messages)
					applyCommits(n, rd.CommittedEntries)
					n.Advance()
				case m := <-n.mbox:
					n.Step(context.TODO(), m)
				}
			}
		}()
	}
}

func main() {
	peers := []raft.Peer{
		{ID: 1, Context: nil},
		{ID: 2, Context: nil},
		{ID: 3, Context: nil},
		{ID: 4, Context: nil},
		{ID: 5, Context: nil},
	}
	nodes := make([]*node, len(peers))
	for i, peer := range peers {
		nodes[i] = buildNode(peer.ID, peers)
	}
	startNodes(nodes)
	cnt := 0
	for {
		time.Sleep(time.Second)
		msg := fmt.Sprintf("node %d, msg %d", nodes[0].Status().ID, cnt)
		cnt++
		nodes[0].Propose(context.TODO(), []byte(msg))
		//for _, n := range nodes {
		//	fmt.Printf("node id = %d, node state = %v\n", n.Status().ID, n.Status())
		//}
	}

}
