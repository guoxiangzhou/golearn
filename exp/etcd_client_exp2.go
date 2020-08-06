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

const (
	totalRecords = 1000000
)

func getOp(numClient int, numQuery int, keyPrefix int, outCnt int, endpoints []string) {

	if (outCnt != 1) && (outCnt != 10) && (outCnt != 100) && (outCnt != 1000) && (outCnt != 10000) {
		log.Fatal("output count should be 10,100,1000,10000")
	}

	ch := make(chan int)
	log.Printf("get, numClient = %d, numQuery=%d, keyPrefix=%d, outCnt=%d\n", numClient, numQuery, keyPrefix, outCnt)
	clients := make([]*clientv3.Client, numClient)
	defer func() {
		for i := 0; i < numClient; i++ {
			if clients[i] != nil {
				clients[i].Close()
			}
		}
	}()

	prefix := ""
	for i := 0; i < keyPrefix; i++ {
		prefix += "0"
	}

	for i := 0; i < numClient; i++ {
		cli, err := clientv3.New(clientv3.Config{
			Endpoints: endpoints,
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
		if outCnt == 1 {
			go func() {
				//log.Printf("start client %d\n", i)
				for j := 0; j < numQuery; j++ {
					key := fmt.Sprintf("%skey%d", prefix, rand.Int()%totalRecords)
					if _, err := cli.Get(context.TODO(), key); err != nil {
						log.Fatal(err)
					}
				}
				ch <- 0
				//log.Printf("finish client %d\n", i)
			}()
		} else {
			go func() {
				//log.Printf("start client %d\n", i)
				rngBegin := totalRecords / (outCnt * 10)
				rngEnd := totalRecords / outCnt
				rngRange := rngEnd - rngBegin
				expCnt := ((outCnt - 1) / 9) + outCnt
				for j := 0; j < numQuery; j++ {
					rnd := rand.Int() % rngRange
					rnd += rngBegin
					key := fmt.Sprintf("%skey%d", prefix, rnd)
					if rsp, err := cli.Get(context.TODO(), key, clientv3.WithPrefix()); err != nil || rsp.Count != int64(expCnt) {
						if err != nil {
							log.Fatal(err)
						} else {
							log.Fatalf("output count = %d, expect output = %d\n", rsp.Count, expCnt)
						}
					}
				}
				ch <- 0
				//log.Printf("finish client %d\n", i)
			}()
		}
	}
	for i := 0; i < numClient; i++ {
		<-ch
	}
	end := time.Now()
	span := end.Sub(start)
	log.Printf("end query, cost %v\n", span)
}

func putOp(numClient int, numQuery int, keyPrefix int, valPrefix int, endpoints []string) {
	ch := make(chan int)
	log.Printf("put, numClient = %d, numQuery=%d, keyPrefix=%d, valPrefix=%d\n", numClient, numQuery, keyPrefix, valPrefix)
	clients := make([]*clientv3.Client, numClient)
	defer func() {
		for i := 0; i < numClient; i++ {
			if clients[i] != nil {
				clients[i].Close()
			}
		}
	}()

	preKey := ""
	for i := 0; i < keyPrefix; i++ {
		preKey += "0"
	}
	preVal := ""
	for i := 0; i < valPrefix; i++ {
		preVal += "0"
	}

	for i := 0; i < numClient; i++ {
		cli, err := clientv3.New(clientv3.Config{
			Endpoints: endpoints,
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
			for j := 0; j < numQuery; j++ {
				key := fmt.Sprintf("%skey%d", preKey, i*numQuery+j)
				val := fmt.Sprintf("%svalue%d", preVal, i*numQuery+j)
				if _, err := cli.Put(context.TODO(), key, val); err != nil {
					log.Fatal(err)
				}
			}
			ch <- 0
		}()
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
	keyPrefix := flag.Int("keyPrefix", 0, "size of key prefix")
	valPrefix := flag.Int("valPrefix", 0, "size of value prefix")
	outCnt := flag.Int("outCnt", 1, "output count 10, 100, 1000, 10000")
	flag.Parse()

	endpoints := []string{"127.0.0.1:10001", "127.0.0.1:10002", "127.0.0.1:10003"}
	if *mod == "get" {
		getOp(*numClient, *numQuery, *keyPrefix, *outCnt, endpoints)
	} else if *mod == "put" {
		putOp(*numClient, *numQuery, *keyPrefix, *valPrefix, endpoints)
	} else {
		log.Fatal("mod should be either 'put' or 'get'")
	}
}
