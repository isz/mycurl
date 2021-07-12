package tcp

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/pkg/errors"
)

type TcpClient struct {
	conn    net.Conn
	timeout time.Duration
	reader  *bufio.Reader
}

type Receiver interface {
	Receive(reader *bufio.Reader) error
}

func NewClient(hostName string, port string, timeout time.Duration) (*TcpClient, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", hostName, port))
	if err != nil {
		return nil, err
	}
	return &TcpClient{
		conn:    conn,
		timeout: timeout,
		reader:  nil,
	}, nil
}

func (c *TcpClient) Write(data []byte) error {
	var wroteBytes int
	bytesToWrite := len(data)
	for bytesToWrite != wroteBytes {
		if err := c.conn.SetDeadline(time.Now().Add(c.timeout)); err != nil {
			return errors.Wrap(err, "set deadline error")
		}

		lastWrote, err := c.conn.Write(data[wroteBytes:])
		if err != nil {
			return errors.Wrap(err, "write to connection error")
		}
		wroteBytes += lastWrote
	}

	return nil
}

func (c *TcpClient) Read(receiver Receiver) error {
	if c.reader == nil {
		c.reader = bufio.NewReader(c.conn)
	}

	for {
		if err := c.conn.SetDeadline(time.Now().Add(c.timeout)); err != nil {
			return errors.Wrap(err, "set deadline error")
		}

		err := receiver.Receive(c.reader)
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return err
		}
	}

}

func (c *TcpClient) Close() error {
	if err := c.conn.Close(); err != nil {
		return errors.Wrap(err, "close connection error")
	}
	return nil
}
