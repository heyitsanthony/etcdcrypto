package main

import (
	"fmt"
	"os"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/namespace"
	"github.com/coreos/etcd/pkg/transport"
	"golang.org/x/net/context"

	"github.com/heyitsanthony/etcdcrypto"
)

func main() {
	tlsinfo := transport.TLSInfo{
		CAFile:   os.Args[2],
		CertFile: os.Args[3],
		KeyFile:  os.Args[4],
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"http://127.0.0.1:2379"},
	})
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	pfx := os.Args[1]
	cli.KV = namespace.NewKV(cli.KV, pfx)
	cli.Watcher = namespace.NewWatcher(cli.Watcher, pfx)
	kx, err := etcdcrypto.NewKeyExchange(cli, "", tlsinfo)
	if err != nil {
		panic(err)
	}
	defer kx.Close()

	var aeskey []byte
	select {
	case aeskey = <-kx.SymmetricKey():
	case <-time.After(5 * time.Second):
		panic("could not establish session key")
	}
	scli := clientv3.NewCtxClient(cli.Ctx())
	cipher, cerr := etcdcrypto.NewAESCipher(aeskey)
	if cerr != nil {
		panic(cerr)
	}
	scli.KV = etcdcrypto.NewKV(cli.KV, cipher)
	scli.Watcher = etcdcrypto.NewWatcher(cli.Watcher, cipher)
	scli.Lease = cli.Lease

	wch := scli.Watch(context.Background(), "hello world/", clientv3.WithPrefix())
	t := fmt.Sprintf("%v", time.Now())
	go func() {
		for {
			if _, err := scli.KV.Put(context.TODO(), "hello world/"+t, t); err != nil {
				panic(err)
			}
			time.Sleep(2 * time.Second)
		}
	}()

	for wr := range wch {
		for _, ev := range wr.Events {
			fmt.Printf("%+v\n", ev)
		}
	}
}
