package bxgateway

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestFilter(t *testing.T) {
	f := NewFilter()
	a := common.HexToAddress("0x7a250d5630b4cf539739df2c5dacb4c659f2488d")
	newf := f.To(a.Hex())
	filterStr, _ := newf.Build()

	gw, err := NewGateway("ws://127.0.0.1:28334/ws", "OWNhOTNhNTktNjg1Yi00N2JhLWFiY2QtMjEzMGE4NTEzZDBkOjliYTg4NjI5NjNlNzlhMTI3ZGRmODcxYTI1ZTY0Yzc3")
	if err != nil {
		t.Error(err)
	}

	c := make(chan string)

	gw.Start(c)

	gw.Subscribe(NewTxs, StreamConfig{
		Include: []string{"tx_hash", "tx_contents"},
		Filter:  filterStr,
	})

	for msg := range c {
		fmt.Println(msg)
	}
}
