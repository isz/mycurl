package http

import (
	"bufio"
	"io"
	"net/textproto"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type statusReceiver uint

type headersReceiver map[string]string

type bodyReceiver struct {
	body   []byte
	toRead int64
	read   int64
}

type chunkedBodyReceiver struct {
	currentChunkSize int64
	totalSize        int64
	bodyReceiver
}

type body interface {
	getBody() []byte
}

func (s *statusReceiver) Receive(reader *bufio.Reader) error {
	line, err := textproto.NewReader(reader).ReadLine()

	if err != nil {
		return err
	}
	return s.getStatus(line)
}

func (s statusReceiver) Code() int {
	return int(s)
}

func (s *statusReceiver) getStatus(line string) error {
	subs := strings.Split(line, " ")

	if len(subs) < 3 {
		return errors.New("protocol error")
	}

	if !checkHTTPVersion(subs[0]) {
		return errors.New("protocol version error, received" + subs[0])
	}

	stat, err := strconv.ParseInt(subs[1], 10, 32)
	if err != nil {
		err = errors.Wrap(err, "parse status code error")
	} else {
		*s = statusReceiver(stat)
		err = io.EOF
	}
	return err
}

var versions = [...]string{"HTTP/1.0", "HTTP/1.1"}

func checkHTTPVersion(version string) bool {
	for _, v := range versions {
		if version == v {
			return true
		}
	}
	return false
}

func (h *headersReceiver) Receive(reader *bufio.Reader) error {
	line, err := textproto.NewReader(reader).ReadLine()

	if err != nil {
		return err
	}

	if line == "" {
		return io.EOF
	}

	if err := h.addHeader(line); err != nil {
		return err
	}

	return nil
}

func (h *headersReceiver) addHeader(headerLine string) error {
	headerPair := strings.Split(headerLine, ": ")

	if len(headerPair) != 2 {

		return errors.New("parse header error")
	}
	(*h)[headerPair[0]] = headerPair[1]
	return nil
}

func newBodyReceiver(length int64) *bodyReceiver {
	return &bodyReceiver{
		body:   make([]byte, length),
		toRead: length,
	}
}

func (b *bodyReceiver) Receive(reader *bufio.Reader) error {
	return b.readBytes(reader)
}

func (b *bodyReceiver) readBytes(reader *bufio.Reader) error {
	r := io.LimitReader(reader, b.toRead)
	received, err := r.Read(b.body[b.read:])
	if err != nil {
		return err
	}

	b.toRead -= int64(received)
	b.read += int64(received)
	if b.toRead == 0 {
		err = io.EOF
	}
	return err
}

func (b *bodyReceiver) getBody() []byte {
	return b.body
}

func newChunkedBodyReceiver() *chunkedBodyReceiver {
	return &chunkedBodyReceiver{
		currentChunkSize: 0,
		totalSize:        0,
		bodyReceiver: bodyReceiver{
			body:   []byte{},
			toRead: 0,
			read:   0,
		},
	}
}

func (chR *chunkedBodyReceiver) Receive(reader *bufio.Reader) error {
	if chR.toRead == 0 {
		chunkSize, err := readChunkSize(reader)
		if err != nil {
			return err
		}

		if chunkSize == 0 {
			return io.EOF
		}
		chR.toRead = chunkSize
		chR.body = append(chR.body, make([]byte, chunkSize)...)
		return nil
	}

	err := chR.readBytes(reader)

	if err != nil {
		if err != io.EOF {
			return err
		}
		_, err = reader.ReadBytes('\n')
	}

	return err
}

func readChunkSize(reader *bufio.Reader) (int64, error) {
	textReader := textproto.NewReader(reader)
	line, err := textReader.ReadLine()

	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(line, 16, 64)
}
