// See LICENSE file for copyright and license details.

package wpasupplicant

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"testing"
	"time"
)

var (
	imposter  = "/tmp/wpa_supplicant_imposter"
	ourSocket string
)

// acts as a fake wpa_supplicant process
func listen(reply chan []byte, exit <-chan int) {
	// Remove this so we can create the socket
	os.Remove(imposter)
	// Start listening on the socket
	conn, err := net.ListenUnixgram("unixgram", &net.UnixAddr{Name: imposter, Net: "unixgram"})
	if err != nil {
		fmt.Printf("failed to listen: %v\n", err)
		panic(err)
	}
	// Our receive buffer
	buf := make([]byte, 2048)
	for {
		// setup read timeout
		t := time.Now()
		t = t.Add(time.Millisecond * 500)
		conn.SetReadDeadline(t)
		// attempt to read
		_, unx, err := conn.ReadFromUnix(buf)
		if err == nil {
			var r []byte
			select {
			case r = <-reply:
				fmt.Printf("Reply: %v\n", string(r))
			default:
			}

			if len(r) == 0 {
				conn.WriteToUnix([]byte("OK\n"), unx)
			} else {
				conn.WriteToUnix(r, unx)
			}
		}
		// check exit channel
		if len(exit) > 0 {
			break
		}
	}
	// Cleanup
	conn.Close()
	os.Remove(imposter)
	fmt.Println("IMPOSTER DOWN")
}

var reply chan []byte

func TestMain(m *testing.M) {
	rd := rand.New(rand.NewSource(99))
	ourSocket = fmt.Sprintf("/tmp/wpa_supplicant%v", rd.Int63())
	reply = make(chan []byte, 2)
	exit := make(chan int, 1)
	go listen(reply, exit)
	time.Sleep(time.Second * 1)
	ret := m.Run()
	exit <- 1
	time.Sleep(time.Second * 1)
	os.Exit(ret)

}
func TestConnect(t *testing.T) {
	conn, err := Connect(ourSocket, imposter)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	conn.Close()
}

func TestDoubleClose(t *testing.T) {
	conn, err := Connect(ourSocket, imposter)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	if err := conn.Close(); err != nil {
		t.Fatalf("first %v", err)
	}
	if err := conn.Close(); err == nil {
		t.Fatalf("second %v", err)
	}
}

func TestSendWithoutConnect(t *testing.T) {
	var conn *Conn
	conn = &Conn{}
	if err := conn.sendRequestOk("SET_NETWORK 0 ssid home"); err == nil {
		t.Fail()
	}
	if _, err := conn.sendRequest("SET_NETWORK 0 ssid home"); err == nil {
		t.Fail()
	}
}

func TestCheckReplyOk(t *testing.T) {
	if err := checkReplyOk([]byte("OK\n")); err != nil {
		t.Fail()
	}
	if err := checkReplyOk([]byte("OK")); err == nil {
		t.Fail()
	}
	if err := checkReplyOk([]byte{}); err == nil {
		t.Fail()
	}
}

func TestSetNetwork(t *testing.T) {
	conn, err := Connect(ourSocket, imposter)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	if err := conn.SetNetwork(0, "ssid", "home"); err != nil {
		t.Fatalf("Failed to get OK reply: %v", err)
	}
	conn.Close()
}

func TestSetNetworkQuoted(t *testing.T) {
	time.Sleep(time.Second)
	conn, err := Connect(ourSocket, imposter)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	if err := conn.SetNetworkQuoted(0, "ssid", "home"); err != nil {
		t.Errorf("Failed : %v", err)
	}
	conn.Close()
}

func TestWepKeys(t *testing.T) {
	conn, err := Connect(ourSocket, imposter)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	if err := conn.SetNetworkWepKeys(0, KeyASCII, []string{"A", "B", "C", "D"}); err != nil {
		t.Fatalf("%v", err)
	}
	if err := conn.SetNetworkWepKeys(0, KeyHex, []string{"A", "B", "C", "D"}); err != nil {
		t.Fatalf("%v", err)
	}
	reply <- []byte("FAIL")
	if err := conn.SetNetworkWepKeys(0, KeyASCII, []string{"BADKEY"}); err == nil {
		t.Fatalf("%v", err)
	}
	reply <- []byte("FAIL")
	if err := conn.SetNetworkWepKeys(0, KeyHex, []string{"BADKEY"}); err == nil {
		t.Fatalf("%v", err)
	}
	conn.Close()
}

func TestConn_GetNetwork(t *testing.T) {
	type args struct {
		id    int
		field string
	}
	tests := []struct {
		name      string
		args      args
		reply     string
		wantValue string
		wantErr   bool
	}{
		{
			name:      "good",
			args:      args{id: 0, field: "ssid"},
			reply:     "home",
			wantValue: "home",
			wantErr:   false,
		},
	}

	// connect
	c, err := Connect(ourSocket, imposter)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reply <- []byte(tt.reply)
			gotValue, err := c.GetNetwork(tt.args.id, tt.args.field)
			if (err != nil) != tt.wantErr {
				t.Errorf("Conn.GetNetwork() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotValue != tt.wantValue {
				t.Errorf("Conn.GetNetwork() = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}

	c.Close()
}

func TestAddNetwork(t *testing.T) {
	conn, err := Connect(ourSocket, imposter)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	reply <- []byte{'1'}
	if v, err := conn.AddNetwork(); err != nil {
		t.Fatalf("%v", err)
	} else if v < 0 {
		t.Fatalf("network value should not be zero")
	}
	conn.Close()
}

func TestRemoveNetwork(t *testing.T) {
	conn, err := Connect(ourSocket, imposter)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	if err := conn.RemoveNetwork(1); err != nil {
		t.Fatalf("%v", err)
	}
	conn.Close()
}

func TestSetGlobalParameters(t *testing.T) {
	conn, err := Connect(ourSocket, imposter)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	if err := conn.SetGlobalParameter("foo", "bar"); err != nil {
		t.Fatalf("%v", err)
	}
	conn.Close()
}

func TestSelectEnableDisableNetwork(t *testing.T) {
	conn, err := Connect(ourSocket, imposter)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	if err := conn.SelectNetwork(1); err != nil {
		t.Fatalf("%v", err)
	}
	if err := conn.EnableNetwork(1); err != nil {
		t.Fatalf("%v", err)
	}
	if err := conn.DisableNetwork(1); err != nil {
		t.Fatalf("%v", err)
	}
	conn.Close()
}

func TestReassociate(t *testing.T) {
	conn, err := Connect(ourSocket, imposter)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	if err := conn.Reassociate(); err != nil {
		t.Fatalf("%v", err)
	}
	conn.Close()
}

func TestReconnect(t *testing.T) {
	conn, err := Connect(ourSocket, imposter)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	if err := conn.Reconnect(); err != nil {
		t.Fatalf("%v", err)
	}
	conn.Close()
}

func TestListNetworks(t *testing.T) {
	conn, err := Connect(ourSocket, imposter)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	reply <- []byte(`network id / ssid / bssid / flags
		0       P7MANU        any
		1       P7MANU5       any
		2       P7MANU6       any
		3       P7MANU7       any
		`)
	if v, err := conn.ListNetworks(); err != nil {
		t.Fatalf("%v", err)
	} else if len(v) < 1 {
		t.Fatalf("should have received more")
	}
	conn.Close()
}

func TestNumOfNetworks(t *testing.T) {
	conn, err := Connect(ourSocket, imposter)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	if v, err := conn.NumOfNetworks(); err != nil {
		t.Fatalf("%v", err)
	} else if v != 0 {
		t.Fatalf("should have been zero")
	}
	conn.Close()
}

func TestReconfigure(t *testing.T) {
	conn, err := Connect(ourSocket, imposter)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	if err := conn.Reconfigure(); err != nil {
		t.Fatalf("%v", err)
	}
	conn.Close()
}
