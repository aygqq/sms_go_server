package com

import (
	"bufio"
	"log"
	"syscall"

	"github.com/schleibinger/sio"
)

var port *sio.Port
var callback func([]byte)

// Init function
func Init(f func([]byte)) error {
	// устанавливаем соединение
	porter, err := sio.Open("/dev/ttyRPMSG0", syscall.B115200)
	if err != nil {
		log.Println("COM Open error ", err)
		return err
	}
	port = porter
	callback = f

	go comRecv()
	return nil
}

// Send - send data to COM
func Send(data []byte) {
	var err error
	// отправляем данные
	_, err = port.Write(data)
	if err != nil {
		log.Println("COM Send error ", err)
		return
	}
}

func comRecv() {
	reader := bufio.NewReader(port)
	for {
		//time.Sleep(time.Second)
		// получаем данные
		reply, err := reader.ReadBytes(0xFE)
		if err != nil {
			log.Println("COM Recv error ", err)
			continue
		}
		callback(reply)
	}
}
