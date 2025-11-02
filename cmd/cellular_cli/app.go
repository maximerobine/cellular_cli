package main

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/maximerobine/cellular_cli/internal/gsm_client"
	"go.bug.st/serial"
)

const serialDevice string = "/dev/serial/by-id/usb-1a86_USB_Serial-if00-port0"

type Application interface {
	Start() error
	Stop() error
	String() string
}

type cellularApp struct {
	log       *slog.Logger
	gsmClient gsm_client.GSMClient
}

func NewApplication(log *slog.Logger) (*cellularApp, error) {
	serialPort, err := initSerial(serialDevice, &serial.Mode{
		BaudRate: 115200,
		Parity:   serial.EvenParity,
		DataBits: 7,
		StopBits: serial.OneStopBit,
	})
	if err != nil {
		return nil, err
	}

	gsmClient := gsm_client.NewGSMClient(serialPort)
	return &cellularApp{
		log:       log,
		gsmClient: *gsmClient,
	}, nil
}

func (a *cellularApp) Start() chan error {
	a.log.Info("starting application")

	errCh := make(chan error)
	go func() {
		err := a.gsmClient.Ping()
		if err != nil {
			errCh <- fmt.Errorf("failed to connect to gsmClient: %v", err)
		}
	}()

	return errCh
}

func (a *cellularApp) Stop() {
	err := a.gsmClient.Close()
	if err != nil {
		a.log.Error(fmt.Sprintf("failed to close gsmClient: %v", err))
	}
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
