package gsm_client

import (
	"bytes"
	"fmt"
	"time"

	"go.bug.st/serial"
)

type GSMClient struct {
	port serial.Port
}

func NewGSMClient(port serial.Port) *GSMClient {
	return &GSMClient{
		port: port,
	}
}

func (c *GSMClient) Ping() error {
	_, err := c.SendATCommand("AT")
	if err != nil {
		return fmt.Errorf("failed to ping modem: %v", err)
	}
	return nil
}

func (c *GSMClient) SendATCommand(cmd string) (string, error) {
	_ = c.port.ResetInputBuffer()
	c.port.Write([]byte("\r"))
	time.Sleep(50 * time.Millisecond)

	_, err := c.port.Write([]byte(cmd + "\r"))
	if err != nil {
		return "", fmt.Errorf("write failed: %v", err)
	}
	// c.log.Debug(fmt.Sprintf("%d bytes sent: %q", n, cmd))

	var response []byte
	buff := make([]byte, 128)
	start := time.Now()
	for {
		n, err := c.port.Read(buff)
		if err != nil {
			return "", fmt.Errorf("read failed: %v", err)
		}
		// c.log.Debug(fmt.Sprintf("RAW: %q", buff))
		// c.log.Debug(fmt.Sprintf("%d bytes read", n))
		response = append(response, buff[:n]...)
		if bytes.HasSuffix(response, []byte("\r\nOK\r\n")) {
			break
		}
		if bytes.HasSuffix(response, []byte("\r\nERROR\r\n")) ||
			bytes.Contains(response, []byte("+CME ERROR:")) {
			return string(response), fmt.Errorf("error from device")
		}
		if time.Since(start) > 2*time.Second {
			return string(response), fmt.Errorf("timeout waiting for response")
		}
	}

	return string(response), nil
}

func (c *GSMClient) Close() error {
	if c.port != nil {
		return nil
	}

	//c.log.Debug("closing serial port")
	return c.port.Close()
}
