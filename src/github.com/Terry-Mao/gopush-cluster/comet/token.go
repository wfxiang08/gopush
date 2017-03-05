// Copyright © 2014 Terry Mao, LiuDing All rights reserved.
// This file is part of gopush-cluster.

// gopush-cluster is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// gopush-cluster is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with gopush-cluster.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	log "github.com/alecthomas/log4go"
	"container/list"
	"errors"
	"time"
)

var (
	// Token exists
	ErrTokenExist = errors.New("token exist")
	// Token not exists
	ErrTokenNotExist = errors.New("token not exist")
	// Token expired
	ErrTokenExpired = errors.New("token expired")
)

//
// 通过LRUMap来记录token, 验证token的有效性
// Token struct
type Token struct {
	token map[string]*list.Element // token map
	lru   *list.List               // lru double linked list
}

// Token Element
type TokenData struct {
	Ticket string
	Expire time.Time
}

// NewToken create a token struct ptr
func NewToken() *Token {
	return &Token{
		token: map[string]*list.Element{},
		lru:   list.New(),
	}
}

// Add add a token
func (t *Token) Add(ticket string) error {
	if e, ok := t.token[ticket]; !ok {
		// new element add to lru back
		e = t.lru.PushBack(&TokenData{Ticket: ticket, Expire: time.Now().Add(Conf.TokenExpire)})
		t.token[ticket] = e
	} else {
		log.Warn("token \"%s\" exist", ticket)
		return ErrTokenExist
	}

	// 添加新元素之后清理
	t.clean()
	return nil
}

// Auth auth a token is valid
func (t *Token) Auth(ticket string) error {
	// 如何认证ticket?
	// 1. ticket不存在
	// 2. ticket
	if e, ok := t.token[ticket]; !ok {
		log.Warn("token \"%s\" not exist", ticket)
		return ErrTokenNotExist
	} else {
		td, _ := e.Value.(*TokenData)
		if time.Now().After(td.Expire) {
			t.clean()
			log.Warn("token \"%s\" expired", ticket)
			return ErrTokenExpired
		}

		// 重新修改Expire
		td.Expire = time.Now().Add(Conf.TokenExpire)
		t.lru.MoveToBack(e)
	}
	t.clean()
	return nil
}

// clean scan the lru list expire the element
func (t *Token) clean() {
	now := time.Now()
	e := t.lru.Front()

	// 从前往后遍历，删除过期的tokens
	for {
		if e == nil {
			break
		}
		td, _ := e.Value.(*TokenData)
		if now.After(td.Expire) {
			log.Warn("token \"%s\" expired", td.Ticket)
			o := e.Next()

			// 从map中删除ticket
			delete(t.token, td.Ticket)

			// 从双向链表
			t.lru.Remove(e)
			e = o
			continue
		}

		// 碰到没有Expire的元素，终止
		break
	}
}
