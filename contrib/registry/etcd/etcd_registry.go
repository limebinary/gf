// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

package etcd

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/gsvc"
	etcd3 "go.etcd.io/etcd/client/v3"
)

func (r *Registry) Register(ctx context.Context, service *gsvc.Service) error {
	r.lease = etcd3.NewLease(r.client)
	grant, err := r.lease.Grant(ctx, int64(r.keepaliveTTL.Seconds()))
	if err != nil {
		return gerror.Wrapf(err, `etcd grant failed with keepalive ttl "%s"`, r.keepaliveTTL)
	}
	var (
		key   = service.Key()
		value = service.Value()
	)
	_, err = r.client.Put(ctx, key, value, etcd3.WithLease(grant.ID))
	if err != nil {
		return gerror.Wrapf(
			err,
			`etcd put failed with key "%s", value "%s", lease "%d"`,
			key, value, grant.ID,
		)
	}
	r.logger.Infof(
		ctx,
		`etcd put success with key "%s", value "%s", lease "%d"`,
		key, value, grant.ID,
	)
	keepAliceCh, err := r.client.KeepAlive(context.Background(), grant.ID)
	if err != nil {
		return err
	}
	go r.doKeepAlive(grant.ID, keepAliceCh)
	return nil
}

func (r *Registry) Deregister(ctx context.Context, service *gsvc.Service) error {
	_, err := r.client.Delete(ctx, service.Key())
	if r.lease != nil {
		_ = r.lease.Close()
	}
	return err
}

// doKeepAlive continuously keeps alive the lease from ETCD.
func (r *Registry) doKeepAlive(leaseID etcd3.LeaseID, keepAliceCh <-chan *etcd3.LeaseKeepAliveResponse) {
	var ctx = context.Background()
	for {
		select {
		case <-r.client.Ctx().Done():
			r.logger.Noticef(ctx, "keepalive done for lease id: %d", leaseID)
			return

		case res, ok := <-keepAliceCh:
			if res != nil {
				//r.logger.Debugf(ctx, `keepalive loop: %v, %s`, ok, res.String())
			}
			if !ok {
				r.logger.Noticef(ctx, `keepalive exit, lease id: %d`, leaseID)
				return
			}
		}
	}
}
