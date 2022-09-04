package server

import (
	"fmt"
	"net"
	"redis-clone/command"
	"redis-clone/util"
)

const DefaultPort int = 6380

type datastore[T any] interface {
	get(string) (T, bool)
	store(string, T)
	delete(string)
}

type redisServer struct {
	store datastore[string]
}

func StartServer(port int) {
	ln, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		fmt.Println("Error during creating connection", err)
		return
	}
	defer ln.Close()

	r := &redisServer{newDataStore()}

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("error when accepting connection", err)
			return
		}
		go r.handleConnection(conn) // not exactly like redis. Redis has single threaded event loop
	}
}

// todo: accept conn io.ReadWriteCloser for better testability
func (r *redisServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	data, b, err := util.ReadSocket(conn)
	if err != nil {
		fmt.Println("error reading socket", err)
		return
	}
	resp := r.handleCommand(data[0:b])
	conn.Write(resp)
}

func (r *redisServer) handleCommand(data []byte) []byte {
	fmt.Println(string(data))

	c, err := command.ParseCommand(data)
	if err != nil {
		fmt.Println("invalid command:", err)
		return []byte("-INVALID_CMD: " + err.Error() + "\r\n")
	}

	switch e := c.(type) {
	case *command.SimpleStringCommand:
		fmt.Println("got string", e.SimpleString())
	case *command.BulkCommand:
		fmt.Println("got bulk", e.BulkString())
	case *command.ArrayCommand:
		cmds := e.Commands()
		fmt.Println("got array", )
		return r.handleRespCommands(cmds)
	default:
		fmt.Println("invalid cmd")
	}
	return []byte("+ok\r\n")
}

func (r *redisServer) handleRespCommands(cmds []string) []byte {
	ok := func() []byte {
		return buildOkResponse("OK")
	}

	if len(cmds) == 0 {
		return buildBadResponse("no command")
	} else if len(cmds) == 1 && cmds[0] == "PING" {
		return buildOkResponse("PONG")
	} else if len(cmds) == 2 {
		switch cmds[0] {
		case "ECHO":
			return buildOkResponse(cmds[1])
		case "GET":
			{
				v, ok := r.store.get(cmds[1])
				if !ok {
					return buildBadResponse("missing key")
				}
				return buildOkResponse(v)
			}
		case "DELETE":
			{
				r.store.delete(cmds[1])
				return ok()
			}
		}
	} else if len(cmds) == 3 && cmds[0] == "SET" {
		r.store.store(cmds[1], cmds[2])
		return ok()
	}

	return buildBadResponse("unknown cmd")
}

func buildBadResponse(resp string) []byte {
	return []byte("-" + resp + "\r\n")
}

func buildOkResponse(resp string) []byte {
	return []byte("+" + resp + "\r\n")
}
