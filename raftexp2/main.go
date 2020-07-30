package main

import (
	"context"
	"go.etcd.io/etcd/raft"
	"go.etcd.io/etcd/raft/raftpb"
	"math/rand"
	"sync"
	"time"
)

type node struct {
	raft.Node
	id      uint64
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
		id:      id,
		storage: st,
	}
	return n
}

func startNodes(nodes []*node) {
	for _, n := range nodes {
		n.mbox = make(chan raftpb.Message, len(nodes))
	}

	for _, n := range nodes {
		n := n
		go func() {
			ticker := time.Tick(5 * time.Millisecond)
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
					for _, m := range rd.Messages {
						m := m
						for _, recvNode := range nodes {
							if recvNode.id != n.id {
								recvNode := recvNode
								go func() {
									b, err := m.Marshal()
									if err != nil {
										panic(err)
									}
									var cm raftpb.Message
									err = cm.Unmarshal(b)
									if err != nil {
										panic(err)
									}
									time.Sleep(time.Duration(rand.Int63n(10)) * time.Millisecond)
									recvNode.mbox <- cm
								}()
							}
						}
					}
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
	for {
		time.Sleep(time.Second)
		//fmt.Print(".")
	}

}
