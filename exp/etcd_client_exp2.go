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

func getOp(numClient int, numQuery int, endpoints []string) {
	ch := make(chan int)
	log.Printf("get, numClient = %d, numQuery=%d\n", numClient, numQuery)
	clients := make([]*clientv3.Client, numClient)
	defer func() {
		for i := 0; i < numClient; i++ {
			if clients[i] != nil {
				clients[i].Close()
			}
		}
	}()

	for i := 0; i < numClient; i++ {
		cli, err := clientv3.New(clientv3.Config{
			Endpoints:   endpoints,
			DialTimeout: 1 * time.Second,
		})
		if err != nil {
			log.Fatal(err)
		}
		clients[i] = cli
	}
	start := time.Now()
	for i := 0; i < numClient; i++ {
		i := i
		cli := clients[i]
		go func() {
			for i := 0; i < numQuery; i++ {
				key := fmt.Sprintf("key%d", rand.Int()%1000000)
				if _, err := cli.Get(context.TODO(), key); err != nil {
					log.Fatal(err)
				}
			}
			ch <- 0
			//log.Printf("finish client %d\n", i)
		}()
		//log.Printf("start client %d\n", i)
	}
	for i := 0; i < numClient; i++ {
		<-ch
	}
	end := time.Now()
	span := end.Sub(start)
	log.Printf("end query, cost %v\n", span)
}

func main() {
	numClient := flag.Int("numClient", 1000, "num of client")
	numQuery := flag.Int("numQuery", 1000, "num of query per client")
	mod := flag.String("mod", "get", "put/get")
	flag.Parse()

	endpoints := []string{"127.0.0.1:10001", "127.0.0.1:10002", "127.0.0.1:10003"}
	if *mod == "get" {
		getOp(*numClient, *numQuery, endpoints)
	} else if *mod == "put" {

	} else {
		log.Fatal("mod should be either 'put' or 'get'")
	}
}
