# Bxgateway-go

Go package for interacting with the Bloxroute Light Gateway.

## Usage

```go
import (
	"fmt"
	"log"

	bxgateway "github.com/jonasbostoen/bxgateway-go"
)

func main() {
	gw, err := bxgateway.NewGateway("ws://127.0.0.1:28334/ws", "AUTH_KEY")
	if err != nil {
		log.Fatal(err)
	}
	defer gw.Close()

	addr := "0x..."
	filter, err := bxgateway.NewFilter().To(addr).Build()	

	c := make(chan string)

	gw.Start(c)

	if err := gw.Subscribe(bxgateway.NewTxs, bxgateway.StreamConfig{
		Include: []string{"tx_contents"},
		Filter:  filter,
	}); err != nil {
		log.Fatal(err)
	}

	for msg := range c {
		fmt.Println(msg)
	}
}

```