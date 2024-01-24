package main

import "sync"

/*
*	管理全部房间
 */

var (
	rManage *roomMge
)

func getRoomManage() *roomMge {
	if rManage == nil {
		rManage = &roomMge{
			rooms: make(map[uint64]*Room),
		}
	}
	return rManage
}

type roomMge struct {
	rooms  map[uint64]*Room
	rMLuck sync.Mutex
}
