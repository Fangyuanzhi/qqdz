package main

import (
	"net"
	"sync"
)

var (
	cMge    *clientMge
	onceCli = &sync.Once{}
)

func getCliMge() *clientMge {
	if cMge == nil {
		onceCli.Do(func() {
			cMge = &clientMge{
				userCli: make(map[uint64]*user),
			}
		})
	}
	return cMge
}

type clientMge struct {
	userCli map[uint64]*user
	uLuck   sync.Mutex
}

type user struct {
	id    uint64
	name  string
	token string
	conn  net.Conn
}

func (this *clientMge) addUser(name, token string, conn net.Conn, id uint64) {
	this.uLuck.Lock()
	this.userCli[id] = &user{
		id:    id,
		name:  name,
		token: token,
		conn:  conn,
	}
	this.uLuck.Unlock()
}
