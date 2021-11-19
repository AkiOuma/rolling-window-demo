package main

import "hystrix-demo/testcase/server"

func main() {
	server.NewUpStreamServer().Run(":9000")
}
