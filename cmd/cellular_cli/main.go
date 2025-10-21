package main

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/maximerobine/at_commands"
	"go.bug.st/serial"
)

const serialDevice string = "/dev/serial/by-id/usb-1a86_USB_Serial-if00-port0"

func sendATCommand(s serial.Port, cmd string, log slog.Logger) (string, error) {
	_ = s.ResetInputBuffer()
	s.Write([]byte("\r"))
	time.Sleep(50 * time.Millisecond)

	n, err := s.Write([]byte(cmd + "\r"))
	if err != nil {
		return "", fmt.Errorf("write failed: %v", err)
	}
	log.Debug(fmt.Sprintf("%d bytes sent: %q", n, cmd))

	var response []byte
	buff := make([]byte, 128)
	start := time.Now()

	for {
		n, err := s.Read(buff)
		if err != nil {
			return "", fmt.Errorf("read failed: %v", err)
		}
		log.Debug(fmt.Sprintf("%d bytes read", n))
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

func initSerial(dev string, mode *serial.Mode) (serial.Port, error) {
	s, err := serial.Open(dev, mode)
	if err != nil {
		return nil, fmt.Errorf("error while openning serial connection: %v", err)
	}

	if port, ok := s.(serial.Port); ok {
		port.SetReadTimeout(500 * time.Millisecond)
	}

	s.ResetInputBuffer()
	s.ResetOutputBuffer()
	time.Sleep(200 * time.Millisecond)

	_, err = s.Write([]byte("\r\n"))
	if err != nil {
		return nil, fmt.Errorf("cannot write newline %v", err)
	}
	time.Sleep(50 * time.Millisecond)

	_, err = s.Write([]byte("ATE0\r"))
	if err != nil {
		return nil, fmt.Errorf("cannot write ATE0 %v", err)
	}
	time.Sleep(50 * time.Millisecond)
	return s, nil
}

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	log.Info("Starting CLI")

	mode := &serial.Mode{
		BaudRate: 115200,
		Parity:   serial.EvenParity,
		DataBits: 7,
		StopBits: serial.OneStopBit,
	}

	s, err := initSerial(serialDevice, mode)
	if err != nil {
		panic(fmt.Errorf("fail to init serial connection: %v", err))
	}

	g, err := NewGSMClient()

	log.Info("sending Clock")
	resp, err := sendATCommand(s, at_commands.Clock.Read(), *log)
	if err != nil {
		log.Error(fmt.Sprintf("error: %v", err))
	}
	fmt.Println(resp)
}
