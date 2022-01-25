// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

package gredis_test

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/database/gredis"
	"github.com/gogf/gf/v2/test/gtest"
)

var (
	ctx = context.TODO()
)

func TestConn_DoWithTimeout(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		redis, err := gredis.New(config)
		t.AssertNil(err)
		t.AssertNE(redis, nil)
		defer redis.Close(ctx)

		conn, err := redis.Conn(ctx)
		t.AssertNil(err)
		defer conn.Close(ctx)

		_, err = conn.Do(ctx, "set", "test", "123")
		t.Assert(err, nil)
		defer conn.Do(ctx, "del", "test")

		r, err := conn.Do(ctx, "get", "test")
		t.Assert(err, nil)
		t.Assert(r.String(), "123")
	})
}

func TestConn_ReceiveVarWithTimeout(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		redis, err := gredis.New(config)
		t.AssertNil(err)
		t.AssertNE(redis, nil)
		defer redis.Close(ctx)

		conn, err := redis.Conn(ctx)
		t.AssertNil(err)
		defer conn.Close(ctx)

		_, err = conn.Do(ctx, "Subscribe", "gf")
		t.AssertNil(err)

		v, err := redis.Do(ctx, "PUBLISH", "gf", "test")

		v, err = conn.Receive(ctx)
		t.AssertNil(err)
		t.Assert(v.Val().(*gredis.Subscription).Channel, "gf")

		v, err = conn.Receive(ctx)
		t.AssertNil(err)
		t.Assert(v.Val().(*gredis.Message).Channel, "gf")
		t.Assert(v.Val().(*gredis.Message).Payload, "test")

	})
}
