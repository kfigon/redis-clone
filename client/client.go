package client

import (
	"fmt"
	"net"
	"redis-clone/util"
)

func SendData(port int, data []byte) ([]byte, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return nil, fmt.Errorf("error connecting to the server: %v", err)
	}
	defer conn.Close()

	n, err := conn.Write(data)
	if err != nil {
		return nil, fmt.Errorf("error sending data to the server: %v", err)
	}
	fmt.Println("send", n, "bytes")
	out, b, err := util.ReadSocket(conn)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	return out[:b], nil
}