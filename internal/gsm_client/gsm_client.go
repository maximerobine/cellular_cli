package gsm_client

import (
	"bytes"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"go.bug.st/serial"
)

type GSMClient struct {
	port serial.Port
	log  slog.Logger
	mu   sync.Mutex
}

func NewGSMClient(port serial.Port, log slog.Logger) *GSMClient {
	return &GSMClient{
		port: port,
		log:  log,
	}
}

func (c *GSMClient) SendATCommand(cmd string) (string, error) {
	c.mu.Lock()
	defer c.mu.Lock()
	_ = c.port.ResetInputBuffer()
	c.port.Write([]byte("\r"))
	time.Sleep(50 * time.Millisecond)

	n, err := c.port.Write([]byte(cmd + "\r"))
	if err != nil {
		return "", fmt.Errorf("write failed: %v", err)
	}
	c.log.Debug(fmt.Sprintf("%d bytes sent: %q", n, cmd))

	var response []byte
	buff := make([]byte, 128)
	start := time.Now()

	for {
		n, err := c.port.Read(buff)
		if err != nil {
			return "", fmt.Errorf("read failed: %v", err)
		}
		c.log.Debug(fmt.Sprintf("%d bytes read", n))
		response = append(response, buff[:n]...)
		if bytes.HasSuffix(response, []byte("\r\nOK\r\n")) ||
			bytes.HasSuffix(response, []byte("\r\nERROR\r\n")) {
			break
		}

		if time.Since(start) > 2*time.Second {
			return string(response), fmt.Errorf("timeout waiting for response")
		}
	}

	return string(response), nil
}

func (c *GSMClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.port != nil {
		return nil
	}

	c.log.Debug("closing serial port")
	return c.port.Close()
}
