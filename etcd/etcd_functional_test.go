package etcd

import (
	"math/rand"
	"testing"
	"time"
)

func TestKillLeader(t *testing.T) {
	tests := []int{3, 5, 9, 11}

	for i, tt := range tests {
		es, hs := buildCluster(tt, false)
		waitCluster(t, es)
		waitLeader(es)

		lead := es[0].node.Leader()
		es[lead].Stop()

		time.Sleep(es[0].tickDuration * defaultElection * 2)

		waitLeader(es)
		if es[1].node.Leader() == 0 {
			t.Errorf("#%d: lead = %d, want not 0", i, es[1].node.Leader())
		}

		for i := range es {
			es[len(es)-i-1].Stop()
		}
		for i := range hs {
			hs[len(hs)-i-1].Close()
		}
	}
	afterTest(t)
}

func TestRandomKill(t *testing.T) {
	tests := []int{3, 5, 9, 11}

	for _, tt := range tests {
		es, hs := buildCluster(tt, false)
		waitCluster(t, es)
		waitLeader(es)

		toKill := make(map[int64]struct{})
		for len(toKill) != tt/2-1 {
			toKill[rand.Int63n(int64(tt))] = struct{}{}
		}
		for k := range toKill {
			es[k].Stop()
		}

		time.Sleep(es[0].tickDuration * defaultElection * 2)

		waitLeader(es)

		for i := range es {
			es[len(es)-i-1].Stop()
		}
		for i := range hs {
			hs[len(hs)-i-1].Close()
		}
	}
	afterTest(t)
}

type leadterm struct {
	lead int64
	term int64
}

func waitLeader(es []*Server) {
	for {
		ls := make([]leadterm, 0, len(es))
		for i := range es {
			switch es[i].mode {
			case participant:
				ls = append(ls, reportLead(es[i]))
			case standby:
				//TODO(xiangli) add standby support
			case stop:
			}
		}
		if isSameLead(ls) {
			return
		}
		time.Sleep(es[0].tickDuration * defaultElection)
	}
}

func reportLead(s *Server) leadterm {
	return leadterm{s.node.Leader(), s.node.Term()}
}

func isSameLead(ls []leadterm) bool {
	m := make(map[leadterm]int)
	for i := range ls {
		m[ls[i]] = m[ls[i]] + 1
	}
	if len(m) == 1 {
		return true
	}
	// todo(xiangli): printout the current cluster status for debugging....
	return false
}
