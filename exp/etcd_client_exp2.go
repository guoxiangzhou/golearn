package main

import (
	"context"
	"flag"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"log"
	"math/rand"
	"time"
)

func main() {
	numClient := flag.Int("numClient", 1000, "num of client")
	numQuery := flag.Int("numQuery", 1000, "num of query per client")
	flag.Parse()

	endpoints := []string{"127.0.0.1:10001", "127.0.0.1:10002", "127.0.0.1:10003"}
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 1 * time.Second,
	})
	defer cli.Close()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("start query")
	ch := make(chan int)

	log.Printf("numClient=%d, numQuery=%d\n", *numClient, *numQuery)

	start := time.Now()
	for i := 0; i < *numClient; i++ {
		go func() {
			for i := 0; i < *numQuery; i++ {
				key := fmt.Sprintf("key", rand.Int()%1000000)
				if _, err := cli.Get(context.TODO(), key); err != nil {
					log.Fatal(err)
				}
			}
			ch <- 0
		}()
	}
	for i := 0; i < *numClient; i++ {
		<-ch
	}
	end := time.Now()
	span := end.Sub(start)
	log.Printf("end query, cost %v\n", span)
}
