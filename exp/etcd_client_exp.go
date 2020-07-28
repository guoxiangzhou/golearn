package main

import (
	"context"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc/grpclog"
	"log"
	"os"
	"time"
)

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

	for _, ep := range endpoints {
		cfg := clientv3.Config{
			Endpoints:            []string{ep},
			DialTimeout:          2 * time.Second,
			DialKeepAliveTime:    2 * time.Second,
			DialKeepAliveTimeout: 6 * time.Second,
		}
		log.Println(cfg)
		localCli, err := clientv3.New(cfg)
		if err != nil {
			log.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.TODO(), 1*time.Second)
		rsp, err := localCli.Get(ctx, "health")
		cancel()
		log.Println("===============")
		log.Println(rsp)
		log.Println(err)
	}

	//for i := 0; i < 100000; i++ {
	//	key := fmt.Sprint("key%d", i)
	//	value := fmt.Sprint("value%d", i)
	//	_, err = cli.Put(context.TODO(), key, value)
	//	fmt.Print(".")
	//	for err != nil {
	//		log.Println("=========================================================")
	//		for _, ep := range endpoints {
	//			resp, err := cli.Status(context.TODO(), ep)
	//			if err == nil {
	//				log.Println("=======================================================")
	//				log.Println(resp)
	//			}
	//		}
	//		return
	//	}
	//}
}
