package main

import "hystrix-demo/testcase/server"

func main() {
	server.NewDownStreamServer(0.2).Run(":8000")
}
