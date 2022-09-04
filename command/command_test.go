package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInvalidCommands(t *testing.T) {
	testCases := []struct {
		desc  string
		input []byte
	}{
		{
			desc:  "too short1",
			input: []byte{},
		},
		{
			desc:  "too short2",
			input: []byte("ab"),
		},
		{
			desc:  "invalid cmd",
			input: []byte("^asd"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			cmd := command(tC.input)
			assert.Error(t, cmd.basicValidation())
		})
	}
}

func TestValidStringCommands(t *testing.T) {
	testCases := []struct {
		desc  string
		input []byte
	}{
		{
			desc:  "short string",
			input: []byte("+OK\r\n"),
		},
		{
			desc:  "short string2",
			input: []byte{'+', 'O', 'K', 0x0D, 0x0A},
		},
		{
			desc:  "short string3",
			input: []byte{'+', 'O', 'K', '\r', '\n'},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			cmd := command(tC.input)
			require.NoError(t, cmd.basicValidation())
			_, err := newSimpleString(tC.input)
			require.NoError(t, err)
		})
	}
}

func TestParseString(t *testing.T) {
	testCases := []struct {
		desc  string
		input []byte
		exp   string
	}{
		{
			desc:  "ok1",
			input: []byte("+OK\r\n"),
			exp:   "OK",
		},
		{
			desc:  "ok2",
			input: []byte{'+', 'O', 'K', 0x0D, 0x0A},
			exp:   "OK",
		},
		{
			desc:  "ok3",
			input: []byte{'+', 'O', 'K', '\r', '\n'},
			exp:   "OK",
		},
		{
			desc:  "some long msg",
			input: []byte("+hello world\r\n"),
			exp:   "hello world",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			cmd, err := newSimpleString(tC.input)
			assert.NoError(t, cmd.basicValidation())
			assert.NoError(t, err)
			assert.Equal(t, tC.exp, cmd.SimpleString())
		})
	}
}

func TestInvalidSimpleString(t *testing.T) {
	data := []byte("+\r\n")
	_, err := newSimpleString(data)
	assert.Error(t, err)
}

func TestBulkString(t *testing.T) {
	testCases := []struct {
		desc            string
		input           []byte
		expected        string
		expectedByteLen int
		expectedLen     int
	}{
		{
			desc:            "short string",
			input:           []byte("$3\r\nHEY\r\n"),
			expected:        "HEY",
			expectedByteLen: 3,
			expectedLen:     9,
		},
		{
			desc:            "longer string",
			input:           []byte("$18\r\nHELLO WORLD my man\r\n"),
			expected:        "HELLO WORLD my man",
			expectedByteLen: 18,
			expectedLen:     25,
		},
		{
			desc:            "string with delimiters inside",
			input:           []byte("$28\r\nTHIS CONTAINS A \r\n INSIDE IT\r\n"),
			expected:        "THIS CONTAINS A \r\n INSIDE IT",
			expectedByteLen: 28,
			expectedLen:     35,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			cmd, err := newBulkString(tC.input)
			require.NoError(t, err)

			assert.Equal(t, tC.expected, cmd.BulkString(), "invalid string parsed")
			assert.Equal(t, tC.expectedByteLen, cmd.byteLen, "invalid len")
			assert.Equal(t, tC.expectedLen, cmd.len(), "invalid len")
		})
	}
}

func TestInvalidBulkStrings(t *testing.T) {
	testCases := []struct {
		desc          string
		input         []byte
		expectedError string
	}{
		{
			desc:          "Invalid first byte",
			input:         []byte("+5\r\nHEY\r\n"),
			expectedError: "invalid first byte",
		},
		{
			desc:          "missing length",
			input:         []byte("$\r\nHEY\r\n"),
			expectedError: "missing length",
		},
		{
			desc:          "missing delimiter",
			input:         []byte("$3HEY\r\n"),
			expectedError: "missing delimiter",
		},
		{
			desc:          "too big size",
			input:         []byte("$15\r\nHEY\r\n"),
			expectedError: "invalid length",
		},
		{
			desc:          "missing termination",
			input:         []byte("$3\r\nHEY"),
			expectedError: "invalid length",
		},
		{
			desc:          "Too little length",
			input:         []byte("$3\r\nHELLO WORLD\r\n"),
			expectedError: "invalid termination",
		},
		{
			desc:          "Empty element",
			input:         []byte("$0\r\n"),
			expectedError: "empty element",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			cmd, err := newBulkString([]byte(tC.input))
			assert.Nil(t, cmd)
			assert.Error(t, err)
			assert.ErrorContains(t, err, tC.expectedError)
		})
	}
}

func TestArrayCommand(t *testing.T) {
	build := func(t *testing.T, data []byte) *ArrayCommand {
		arr, err := newArrayString(data)
		require.NoError(t, err)
		return arr
	}

	t.Run("simple array", func(t *testing.T) {
		data := []byte("*2\r\n" + "+OK\r\n" + "+hello world\r\n")
		arr := build(t, data)

		require.Len(t, arr.Commands(), 2)

		assert.Equal(t, "OK", arr.Commands()[0])
		assert.Equal(t, "hello world", arr.Commands()[1])
	})

	t.Run("array with bulk", func(t *testing.T) {
		data := []byte("*2\r\n" + "+OK\r\n" + "$28\r\nTHIS CONTAINS A \r\n INSIDE IT\r\n")
		arr := build(t, data)

		require.Len(t, arr.Commands(), 2)

		assert.Equal(t, "OK", arr.Commands()[0])
		assert.Equal(t, "THIS CONTAINS A \r\n INSIDE IT", arr.Commands()[1])
	})

	t.Run("array with bulks", func(t *testing.T) {
		data := []byte("*2\r\n" + "$2\r\nOK\r\n" + "$28\r\nTHIS CONTAINS A \r\n INSIDE IT\r\n")
		arr := build(t, data)

		require.Len(t, arr.Commands(), 2)

		assert.Equal(t, "OK", arr.Commands()[0])
		assert.Equal(t, "THIS CONTAINS A \r\n INSIDE IT", arr.Commands()[1])
	})

	t.Run("many arrays", func(t *testing.T) {
		t.Skip("todo")

		data := []byte("*3\r\n" + "$2\r\nOK\r\n" + "*3\r\n" + "+first\r\n" + "+second\r\n" + "$5\r\nthird\r\n" + "+end\r\n")
		arr := build(t, data)

		require.Len(t, arr.Commands(), 4)

		assert.Equal(t, "OK", arr.Commands()[0])
		assert.Equal(t, "first", arr.Commands()[1])
		assert.Equal(t, "second", arr.Commands()[2])
		assert.Equal(t, "third", arr.Commands()[3])
		assert.Equal(t, "end", arr.Commands()[4])
	})
}

func TestInvalidArrays(t *testing.T) {
	t.Run("Invalid number of elements", func(t *testing.T) {
		data := []byte("*2\r\n" + "+OK\r\n")
		_, err := newArrayString(data)
		assert.Error(t, err)
	})

	t.Run("Empty element", func(t *testing.T) {
		data := []byte("*2\r\n$6\r\nDELETE\r\n$0\r\n")
		_, err := newArrayString(data)
		assert.Error(t, err)
	})

	t.Run("Empty array", func(t *testing.T) {
		data := []byte("*0\r\n")
		_, err := newArrayString(data)
		assert.Error(t, err)
	})
}
