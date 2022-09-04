package command

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
)

const DELIMITER_LENGTH = 2

// because we lack sum types in go...
type commandBase interface {
	dummy()
}

func ParseCommand(b []byte) (commandBase, error) {
	cmd := command(b)
	if err := cmd.basicValidation(); err != nil {
		return nil, err
	}
	if cmd.isStringCmd() {
		return newSimpleString(cmd)
	} else if cmd.isBulk() {
		return newBulkString(cmd)
	} else if cmd.isArray() {
		return newArrayString(cmd)
	}
	return nil, fmt.Errorf("invalid command")
}

type command []byte

func (c command) basicValidation() error {
	if len(c) < 3 {
		return fmt.Errorf("too short")
	} else if !c.isStringCmd() && !c.isBulk() && !c.isArray() {
		return fmt.Errorf("invalid first character: %q", c[0])
	}
	return nil
}

func terminationError(termination []byte) error {
	return fmt.Errorf("invalid termination: %q", string(termination))
}

func stringTerminated(termination []byte) bool {
	return equal(termination, []byte{0x0d, 0x0a})
}

func (c command) isStringCmd() bool {
	return c[0] == '+'
}

func (c command) isArray() bool {
	return c[0] == '*'
}

func (c command) isBulk() bool {
	return c[0] == '$'
}

func equal[T comparable](a []T, b []T) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func parseLengthToken(c command) (int, error) {
	byteLenStr := ""
	for i := 1; i < len(c)-1; i++ {
		this := c[i]
		next := c[i+1]
		if unicode.IsDigit(rune(this)) {
			byteLenStr += string(this)
		} else if byteLenStr == "" && this == '\r' && next == '\n' {
			return 0, fmt.Errorf("missing length")
		} else if byteLenStr != "" && this == '\r' && next == '\n' {
			break
		} else {
			return 0, fmt.Errorf("missing delimiter")
		}
	}
	byteLen, err := strconv.Atoi(byteLenStr)
	if err != nil {
		return 0, fmt.Errorf("invalid len %q: %v", byteLenStr, err)
	}
	return byteLen, nil
}

type SimpleStringCommand struct {
	command
}

func newSimpleString(c command) (*SimpleStringCommand, error) {
	if !c.isStringCmd() {
		return nil, fmt.Errorf("invalid first byte: %q", c[0])
	} else if len(c) == 3 {
		return nil, fmt.Errorf("empty command")
	}
	ln := 1
	for ln < len(c)-1 {
		this := c[ln]
		next := c[ln+1]
		if this == '\r' && next == '\n' {
			break
		}
		ln++
	}
	if !stringTerminated(c[ln : ln+2]) {
		return nil, terminationError(c[ln : ln+2])
	}
	return &SimpleStringCommand{c[0 : ln+2]}, nil
}

func (s *SimpleStringCommand) termination() []byte {
	c := s.command
	return c[len(c)-DELIMITER_LENGTH:]
}

func (s *SimpleStringCommand) SimpleString() string {
	c := s.command
	return string(c[1 : len(c)-DELIMITER_LENGTH])
}

func (_ *SimpleStringCommand) dummy() {}

type BulkCommand struct {
	command
	byteLen int
}

func newBulkString(c command) (*BulkCommand, error) {
	if !c.isBulk() {
		return nil, fmt.Errorf("invalid first byte: %q", c[0])
	}

	byteLen, err := parseLengthToken(c)
	if err != nil {
		return nil, err
	} else if byteLen == 0 {
		return nil, fmt.Errorf("empty element")
	} else if expectedBulkLen(byteLen) > len(c) {
		return nil, fmt.Errorf("invalid length")
	} else if !stringTerminated(c[expectedBulkLen(byteLen)-2 : expectedBulkLen(byteLen)]) {
		return nil, terminationError(c[expectedBulkLen(byteLen)-2 : expectedBulkLen(byteLen)])
	}
	return &BulkCommand{c[0:expectedBulkLen(byteLen)], byteLen}, nil
}

func (b *BulkCommand) BulkString() string {
	charsOfByteLen := charLenOfNum(b.byteLen)
	start := 1 + charsOfByteLen + 2
	end := start + b.byteLen
	return string(b.command)[start:end]
}
func (_ *BulkCommand) dummy() {}

func expectedBulkLen(ln int) int {
	byteLenStr := charLenOfNum(ln)
	return 1 + byteLenStr + DELIMITER_LENGTH + ln + DELIMITER_LENGTH
}

func charLenOfNum(ln int) int {
	return int(math.Log10(float64(ln))) + 1
}

func (b *BulkCommand) len() int {
	return expectedBulkLen(b.byteLen)
}

type ArrayCommand struct {
	cmds []commandBase
}

func newArrayString(c command) (*ArrayCommand, error) {
	if !c.isArray() {
		return nil, fmt.Errorf("invalid first byte: %q", c[0])
	}
	arrayLen, err := parseLengthToken(c)
	if err != nil {
		return nil, err
	} else if arrayLen <= 0 {
		return nil, fmt.Errorf("empty array")
	}
	cmds := []commandBase{}

	i := 1 + charLenOfNum(arrayLen) + DELIMITER_LENGTH
	for i < len(c) {
		subCmd := command(c[i:])

		if err := subCmd.basicValidation(); err != nil {
			return nil, err
		}

		switch {
		case subCmd.isArray():
			a, err := newArrayString(subCmd)
			if err != nil {
				return nil, err
			}
			cmds = append(cmds, a)
			i++ // todo
			break
		case subCmd.isStringCmd():
			s, err := newSimpleString(subCmd)
			if err != nil {
				return nil, err
			}
			cmds = append(cmds, s)
			i += len(s.command)
		case subCmd.isBulk():
			b, err := newBulkString(subCmd)
			if err != nil {
				return nil, err
			}
			cmds = append(cmds, b)
			i += len(b.command)
		default:
			return nil, fmt.Errorf("invalid fist byte %q", subCmd[0])
		}
	}
	if len(cmds) != arrayLen {
		return nil, fmt.Errorf("invalid number of elements detected. Declared %d, got %d", arrayLen, len(cmds))	
	}

	return &ArrayCommand{cmds}, nil
}

func (_ *ArrayCommand) dummy() {}

func (a *ArrayCommand) Commands() []string {
	var out []string
	for _, c := range a.cmds {
		switch e := c.(type) {
		case *SimpleStringCommand:
			out = append(out, e.SimpleString())
		case *BulkCommand:
			out = append(out, e.BulkString())
		case *ArrayCommand:
			out = append(out, e.Commands()...) // probably wrong, todo when we support nested arrays
		}
	}
	return out
}

func buildArrayBulkCmd(data []string) string {
	buf := fmt.Sprintf("*%d\r\n", len(data))
	for _, v := range data {
		buf += fmt.Sprintf("$%d\r\n%s\r\n", len(v), v)
	}
	return buf
}

func BuildGetCommand(data string) string {
	return buildArrayBulkCmd([]string{"GET", data})
}

func BuildDeleteCommand(data string) string {
	return buildArrayBulkCmd([]string{"DELETE", data})
}

func BuildSetCommand(data string) string {
	splitted := strings.Split(data, "=")

	in := []string{"SET"}
	for _, v := range splitted {
		in = append(in, v)
	}
	return buildArrayBulkCmd(in)
}

func BuildPingCommand() string {
	return buildArrayBulkCmd([]string{"PING"})
}

func BuildEchoCommand(data string) string {
	return buildArrayBulkCmd([]string{"ECHO", data})
}