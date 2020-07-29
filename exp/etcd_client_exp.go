package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc/grpclog"
	"log"
	"os"
	"time"
)

func isHealth(endpoints []string) bool {
	ch := make(chan bool, len(endpoints))
	defer close(ch)
	for _, ep := range endpoints {
		go func(endpoint string, chH chan<- bool) {
			cfg := clientv3.Config{
				Endpoints: []string{endpoint},
			}
			cli, err := clientv3.New(cfg)
			defer cli.Close()
			if err != nil {
				chH <- false
				return
			}
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			if _, err = cli.Get(ctx, "health"); err != nil {
				chH <- false
				return
			}
			cancel()
			chH <- true
		}(ep, ch)
	}
	cntHealth := 0
	for _ = range endpoints {
		if <-ch {
			cntHealth++
		}
	}
	return cntHealth*2 > len(endpoints)
}

func main() {
	clientv3.SetLogger(grpclog.NewLoggerV2(os.Stderr, os.Stderr, os.Stderr))
	endpoints := []string{"127.0.0.1:10001", "127.0.0.1:10002", "127.0.0.1:10003"}
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 1 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	for i := 0; i < 100000; i++ {
		key := fmt.Sprint("key%d", i)
		value := fmt.Sprint("value%d", i)
		_, err = cli.Put(context.TODO(), key, value)
		fmt.Print(".")
		for err != nil {
			if isHealth(endpoints) {
				_, err = cli.Put(context.TODO(), key, value)
				fmt.Print(".")
			} else {
				log.Fatal(err)
			}
		}
	}
	fmt.Printf("\n==================\n")
}
