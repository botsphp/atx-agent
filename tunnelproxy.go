package main

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/openatx/androidutils"
	"github.com/openatx/atx-server/proto"
)

func runTunnelProxy(serverAddr string) {
	var t time.Time
	n := 0
	for {
		t = time.Now()
		unsafeRunTunnelProxy(serverAddr)
		if time.Since(t) > time.Minute {
			n = 0
		}
		n++
		if n > 5 {
			n = 5
		}
		log.Printf("dial server error, reconnect after %d seconds", n*5)
		time.Sleep(time.Duration(n) * 5 * time.Second) // 5, 10, ... 50s
	}
}

func unsafeRunTunnelProxy(serverAddr string) error {
	c, _, err := websocket.DefaultDialer.Dial("ws://"+serverAddr+"/echo", nil)
	if err != nil {
		return err
	}
	defer c.Close()

	props, _ := androidutils.Properties()
	devInfo := &proto.DeviceInfo{
		Serial: props["ro.serialno"],
		Brand:  props["ro.product.brand"],
		Model:  props["ro.product.model"],
	}
	devInfo.HWAddr, _ = androidutils.HWAddrWLAN()
	c.WriteJSON(proto.CommonMessage{
		Type: proto.DeviceInfoMessage,
		Data: devInfo,
	})

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case t := <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
			if err != nil {
				log.Println(err)
				return err
			}
		}
	}
}