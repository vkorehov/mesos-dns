package records

import (
	// "fmt"
	"testing"
)

func TestHostBySlaveId(t *testing.T) {

	slaves := []slave{
		{Id: "20140827-000744-3041283216-5050-2116-1", Hostname: "blah.com"},
		{Id: "33333333-333333-3333333333-3333-3333-2", Hostname: "blah.blah.com"},
	}

	rg := RecordGenerator{Slaves: slaves}

	for i := 0; i < len(slaves); i++ {
		host, err := rg.hostBySlaveId(slaves[i].Id)
		if err != nil {
			t.Error(err)
		}

		if host != slaves[i].Hostname {
			t.Error("wrong slave/hostname")
		}
	}

}

func TestYankPort(t *testing.T) {
	p := "[31328-31328]"

	port := yankPort(p)

	if port != "31328" {
		t.Error("not parsing port")
	}
}

func TestLeaderIP(t *testing.T) {
	l := "master@144.76.157.37:5050"

	ip := leaderIP(l)

	if ip != "144.76.157.37" {
		t.Error("not parsing ip")
	}
}

func TestStripUID(t *testing.T) {
	tname := "reviewbot.8c9b3434-615a-11e4-a088-c20493233aa5"

	name := stripUID(tname)

	if name != "reviewbot" {
		t.Error("not parsing task name")
	}
}
