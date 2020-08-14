package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/gob"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"go.etcd.io/etcd/pkg/pbutil"
	"go.etcd.io/etcd/raft"
	"go.etcd.io/etcd/raft/raftpb"
	"log"
	"os"
	"sync"
	"time"
)

const (
	walPath = "/tmp/wal.db"
)

type node struct {
	raft.Node
	storage *raft.MemoryStorage

	kvMu    sync.Mutex
	kvstore map[string]string

	walMu sync.Mutex
	wal   *sql.DB
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
				n.kvMu.Lock()
				n.kvstore[datakv.Key] = datakv.Value
				n.kvMu.Unlock()
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

func (n *node) createWalTables() {
	if _, err := n.wal.Exec("CREATE TABLE wal (idx INTEGER ,data BLOB)"); err != nil {
		log.Fatal(err)
	}
	// idx = 0 for hardstate
	if _, err := n.wal.Exec("INSERT INTO wal VALUES (0,NULL)"); err != nil {
		log.Fatal(err)
	}
}

func (n *node) openWal() {
	db, err := sql.Open("sqlite3", walPath)
	if err != nil {
		log.Fatal(err)
	}
	n.wal = db
}

func (n *node) saveWal(st raftpb.HardState, ents []raftpb.Entry) {
	n.walMu.Lock()
	defer n.walMu.Unlock()

	if !raft.IsEmptyHardState(st) {
		stmt, err := n.wal.Prepare("UPDATE wal SET data = ? WHERE idx=0")
		if err != nil {
			log.Fatal(err)
		}
		stmt.Exec(pbutil.MustMarshal(&st))
		stmt.Close()
	}
	if len(ents) > 0 {
		stmt, err := n.wal.Prepare("INSERT INTO wal VALUES (?,?)")
		if err != nil {
			log.Fatal(err)
		}
		for _, ent := range ents {
			stmt.Exec(ent.Index, pbutil.MustMarshal(&ent))
		}
		stmt.Close()
	}
}

func (n *node) startServer(stopc chan struct{}) {
	ticker := time.Tick(10 * time.Millisecond)
	for {
		select {
		case <-ticker:
			n.Tick()
		case rd := <-n.Ready():
			n.saveWal(rd.HardState, rd.Entries)

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
}

func createNode(id uint64) *node {
	st := raft.NewMemoryStorage()
	c := &raft.Config{
		ID:              id,
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
	return n
}

func loadWal() (state raftpb.HardState, entries []raftpb.Entry) {
	db, err := sql.Open("sqlite3", walPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	stRows, err := db.Query("SELECT data FROM wal WHERE idx=0")
	if err != nil {
		log.Fatal(err)
	}
	defer stRows.Close()
	for stRows.Next() {
		var stBytes []byte
		err := stRows.Scan(&stBytes)
		if err != nil {
			log.Fatal(err)
		}
		pbutil.MustUnmarshal(&state, stBytes)
	}
	walRows, err := db.Query("SELECT idx, data FROM wal WHERE idx >0 ORDER BY idx ASC")
	if err != nil {
		log.Fatal(err)
	}
	defer walRows.Close()
	for walRows.Next() {
		var idx uint64
		var walBytes []byte
		err := walRows.Scan(&idx, &walBytes)
		if err != nil {
			log.Fatal(err)
		}
		var entry raftpb.Entry
		pbutil.MustUnmarshal(&entry, walBytes)
		entries = append(entries, entry)
	}
	return state, entries
}

func restart() {
	st := raft.NewMemoryStorage()
	c := &raft.Config{
		ID:              0x01,
		ElectionTick:    10,
		HeartbeatTick:   1,
		Storage:         st,
		MaxSizePerMsg:   4096,
		MaxInflightMsgs: 256,
	}
	stat, ents := loadWal()
	st.SetHardState(stat)
	st.Append(ents)
	rn := raft.RestartNode(c)

	n := &node{
		Node:    rn,
		storage: st,
		kvstore: make(map[string]string),
	}

	stopc := make(chan struct{})

	go n.startServer(stopc)

	time.Sleep(time.Second)
	stopc <- struct{}{}
	time.Sleep(time.Second)
	for key, value := range n.kvstore {
		log.Printf("%s,%s", key, value)
	}
}

func main() {
	n := createNode(0x01)
	if _, err := os.Stat(walPath); err == nil {
		os.Remove(walPath)
	}

	n.openWal()
	defer n.wal.Close()

	n.createWalTables()

	stopc := make(chan struct{})

	go n.startServer(stopc)
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
	log.Println("============== restart ======================")
	restart()
}
