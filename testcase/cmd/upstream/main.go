package main

import (
	"hystrix-demo/testcase/server"
	"time"
)

func main() {
	server.NewUpStreamServer(
		10,
		50,
		0.8,
		time.Second*5,
	).Run(":9000")
}
