package WGMCS

import (
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/netip"
	"sync"

	"go.mau.fi/libsignal/ecc"
	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun"
	"golang.zx2c4.com/wireguard/tun/netstack"
)

type CloudServer struct {
	Device *device.Device
	Tun    *tun.Device
	Conn   *netstack.Net
	Prv    string
	Pub    string
	Port   uint16
	sync.Mutex
}

func (wgmcs *CloudServer) GenDevices() error {
	var err error
	wgmcs.Lock()
	defer wgmcs.Unlock()
	Tun, Conn, err := netstack.CreateNetTUN(
		[]netip.Addr{netip.MustParseAddr("10.17.0.1")},
		[]netip.Addr{netip.MustParseAddr("1.1.1.1"), netip.MustParseAddr("1.0.0.1")},
		1420,
	)
	if err != nil {
		return err
	}
	wgmcs.Conn = Conn
	wgmcs.Tun = &Tun
	keypair, err := ecc.GenerateKeyPair()
	if err != nil {
		return err
	}
	wgmcs.Device = device.NewDevice(*wgmcs.Tun, conn.NewDefaultBind(), device.NewLogger(device.LogLevelError, "WGController"))
	pubbytes := keypair.PublicKey().Serialize()
	prvbytes := keypair.PrivateKey().Serialize()
	wgmcs.Pub = hex.EncodeToString(pubbytes[:(len(pubbytes) - 1)])
	wgmcs.Prv = hex.EncodeToString(prvbytes[:])
	wgmcs.Port = uint16(8902)
	allowed_ip := "10.17.0.0/32"
	interval := 25
	uapiConf := fmt.Sprintf("private_key=%s\nlisten_port=%d\npublic_key=%s\nallowed_ip=%s\npersistent_keepalive_interval=%d\nendpoint=127.0.0.1:%d",
		wgmcs.Prv, wgmcs.Port, wgmcs.Pub, allowed_ip, interval, wgmcs.Port)
	if wgmcs.Device.IpcSet(uapiConf); err != nil {
		return err
	}
	return nil
}
func (wgmcs *CloudServer) UpDevice() error {
	if err := wgmcs.Device.Up(); err != nil {
		log.Fatal("err", err)
	}
	log.Println("prv key:", wgmcs.Prv)
	log.Println("pup key:", wgmcs.Pub)
	log.Println("WireGuard server started")
	listener, err := wgmcs.Conn.ListenTCP(&net.TCPAddr{Port: 80})
	if err != nil {
		log.Panicln(err)
	}
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		log.Printf("> %s - %s - %s", request.RemoteAddr, request.URL.String(), request.UserAgent())
		io.WriteString(writer, "Hello from userspace TCP!")
	})
	err = http.Serve(listener, nil)
	if err != nil {
		log.Panicln(err)
	}
	// Keep the server running
	select {}
}
