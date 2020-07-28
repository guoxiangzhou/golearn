package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	gcode "google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	gs "google.golang.org/grpc/status"
	"log"
	"os"
	"time"
)

func main() {
	clientv3.SetLogger(grpclog.NewLoggerV2(os.Stderr, os.Stderr, os.Stderr))
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:10001", "127.0.0.1:10002", "127.0.0.1:10003"},
		DialTimeout: 1 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	//if _, err = cli.Put(context.TODO(), "foo", "bar"); err != nil {
	//	log.Fatal(err)
	//}

	//if rsp, err := cli.Get(context.TODO(), "foo"); err == nil {
	//	for _, kv := range rsp.Kvs {
	//		log.Printf("key=%s,value=%s, version=%d\n", kv.Key, kv.Value, kv.Version)
	//	}
	//} else {
	//	log.Fatal(err)
	//}

	for i := 0; i < 100000; i++ {
		key := fmt.Sprint("key%d", i)
		value := fmt.Sprint("value%d", i)
		_, err = cli.Put(context.TODO(), key, value)
		fmt.Print(".")
		for err != nil {
			grpcErr, _ := gs.FromError(err)
			if grpcErr.Code() == gcode.InvalidArgument {
				log.Fatalf("falt error, index = %d, error = %v\n", i, err)
			}
			log.Printf("==================================\n")
			log.Printf("grpc error, code = %v \n", grpcErr.Code())
			_, err = cli.Put(context.TODO(), key, value)
			fmt.Print(".")
		}
	}
}
