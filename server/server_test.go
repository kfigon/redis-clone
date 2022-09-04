package server

import (
	"redis-clone/client"
	"redis-clone/command"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testPort = 6666

func runAndSend(t *testing.T, data string) []byte {
	go StartServer(testPort)
	resp, err := client.SendData(testPort, []byte(data))
	assert.NoError(t, err)

	return resp
}

func TestSendSimpleString(t *testing.T) {
	resp := runAndSend(t, "+some string\r\n")
	assert.Equal(t, "+ok\r\n", string(resp))
}

func TestRespCommands(t *testing.T) {
	testCases := []struct {
		desc	string
		input   []byte
		expectedOut   []byte
	}{
		{
			desc: "Ping",
			input: []byte("*1\r\n$4\r\nPING\r\n"),
			expectedOut: []byte("+PONG\r\n"),
		},
		{
			desc: "Echo",
			input: []byte("*2\r\n$4\r\nECHO\r\n$3\r\nhey\r\n"),
			expectedOut: []byte("+hey\r\n"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			resp := runAndSend(t, string(tC.input))
			assert.Equal(t, tC.expectedOut, resp)
		})
	}
}

func TestInvalidCommand(t *testing.T) {
	resp := runAndSend(t, "invalid")
	assert.Equal(t, "-INVALID_CMD: invalid first character: 'i'\r\n", string(resp))
}

func TestMultipleCmds(t *testing.T) {
	go StartServer(testPort)
	client.SendData(testPort, []byte(command.BuildSetCommand("foo=123")))
	resp,err := client.SendData(testPort, []byte(command.BuildGetCommand("foo")))
	
	assert.NoError(t, err)
	assert.Equal(t, "+123\r\n", string(resp))
}