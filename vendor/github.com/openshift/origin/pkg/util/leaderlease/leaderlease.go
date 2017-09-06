package leaderlease

import (
	"fmt"
	"time"

	etcdclient "github.com/coreos/etcd/client"
	"github.com/golang/glog"
	"golang.org/x/net/context"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	utilwait "k8s.io/apimachinery/pkg/util/wait"
	etcdutil "k8s.io/apiserver/pkg/storage/etcd/util"
)

// Leaser allows a caller to acquire a lease and be notified when it is lost.
type Leaser interface {
	// AcquireAndHold tries to acquire the lease and hold it until it expires, the lease is lost,
	// or we observe another party take the lease. The provided function will be invoked when the
	// lease is acquired, and the provided channel will be closed when the lease is lost. If the
	// function returns true, the lease will be released on exit. If the function returns false,
	// the lease will be held.
	AcquireAndHold(chan error)
	// Release returns any active leases
	Release()
}

// Etcd takes and holds a leader lease until it can no longer confirm it owns
// the lease, then returns.
type Etcd struct {
	client     etcdclient.Client
	keysClient etcdclient.KeysAPI
	key        string
	value      string
	ttl        uint64

	// the fraction of the ttl to wait before trying to renew - for instance, 0.75 with TTL 20
	// will wait 15 seconds before attempting to renew the lease, then retry over the next 5
	// seconds in the event of an error no more than maxRetries times.
	waitFraction float32
	// the interval to wait when an error occurs acquiring the lease
	pauseInterval time.Duration
	// the maximum retries when releasing or renewing the lease
	maxRetries int
	// the shortest time between attempts to renew the lease
	minimumRetryInterval time.Duration
}

// NewEtcd creates a Lease in etcd, storing value at key with expiration ttl
// and continues to refresh it until the key is lost, expires, or another
// client takes it.
func NewEtcd(client etcdclient.Client, key, value string, ttl uint64) Leaser {
	return &Etcd{
		client:     client,
		keysClient: etcdclient.NewKeysAPI(client),
		key:        key,
		value:      value,
		ttl:        ttl,

		waitFraction:         0.66,
		pauseInterval:        time.Second,
		maxRetries:           10,
		minimumRetryInterval: 100 * time.Millisecond,
	}
}

const autoSyncInterval = 10 * time.Second

// AcquireAndHold implements an acquire and release of a lease.
func (e *Etcd) AcquireAndHold(notify chan error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		// Because the call to e.keysClient.Set in tryAcquire is using PrevNoExist, etcd considers this
		// to be a "one-shot" attempt, meaning that if the connection attempt to one of the etcd cluster
		// members fails, it will not fail over to any of the other cluster members. Calling
		// e.client.AutoSync is not a one-shot call, and it will try to contact each cluster member
		// until it succeeds. Assuming it does, the client's list of endpoints is updated, and any
		// unavailable members are removed from the list.
		for {
			err := e.client.AutoSync(ctx, autoSyncInterval)
			if err == context.DeadlineExceeded || err == context.Canceled {
				break
			}
			utilruntime.HandleError(err)
			time.Sleep(e.pauseInterval)
		}
	}()

	for {
		ok, ttl, index, err := e.tryAcquire()
		if err != nil {
			utilruntime.HandleError(err)
			time.Sleep(e.pauseInterval)
			continue
		}
		if !ok {
			time.Sleep(e.pauseInterval)
			continue
		}

		// notify
		notify <- nil
		defer close(notify)

		// hold the lease
		if err := e.tryHold(ttl, index); err != nil {
			notify <- err
		}
		break
	}
}

// tryAcquire tries to create the lease key in etcd, or if it already exists
// and belongs to another user, to wait until the lease expires or is deleted.
// It returns true if the lease was acquired, the current TTL, the nextIndex
// to watch from, or an error.
func (e *Etcd) tryAcquire() (ok bool, ttl uint64, nextIndex uint64, err error) {
	ttl = e.ttl

	resp, err := e.keysClient.Set(
		context.Background(),
		e.key,
		e.value,
		&etcdclient.SetOptions{
			TTL:       time.Duration(ttl) * time.Second,
			PrevExist: etcdclient.PrevNoExist,
		},
	)
	if err == nil {
		// we hold the lease
		index := resp.Index
		glog.V(4).Infof("Lease %s acquired at %d, ttl %d seconds", e.key, index, e.ttl)
		return true, ttl, index + 1, nil
	}

	if !etcdutil.IsEtcdNodeExist(err) {
		return false, 0, 0, fmt.Errorf("unable to check lease %s: %v", e.key, err)
	}

	latest, err := e.keysClient.Get(context.Background(), e.key, nil)
	if err != nil {
		return false, 0, 0, fmt.Errorf("unable to retrieve lease %s: %v", e.key, err)
	}

	nextIndex = eventIndexFor(latest)
	if latest.Node.TTL > 0 {
		ttl = uint64(latest.Node.TTL)
	}

	if latest.Node.Value != e.value {
		glog.V(4).Infof("Lease %s owned by %s at %d ttl %d seconds, waiting for expiration", e.key, latest.Node.Value, nextIndex-1, ttl)
		// waits until the lease expires or changes to us.
		// TODO: it's possible we were given the lease during the watch, but we just expect to go
		//   through this loop again and let this condition check
		if _, err := e.waitForExpiration(false, nextIndex, nil); err != nil {
			return false, 0, 0, fmt.Errorf("unable to wait for lease expiration %s: %v", e.key, err)
		}
		return false, 0, 0, nil
	}

	glog.V(4).Infof("Lease %s already held, expires in %d seconds", e.key, ttl)
	return true, ttl, nextIndex, nil
}

// Release tries to delete the leader lock.
func (e *Etcd) Release() {
	for i := 0; i < e.maxRetries; i++ {
		_, err := e.keysClient.Delete(context.Background(), e.key, &etcdclient.DeleteOptions{PrevValue: e.value})
		if err == nil {
			break
		}
		// If the value has changed, we don't hold the lease. If the key is missing we don't
		// hold the lease.
		if etcdutil.IsEtcdTestFailed(err) || etcdutil.IsEtcdNotFound(err) {
			break
		}
		utilruntime.HandleError(fmt.Errorf("unable to release %s: %v", e.key, err))
	}
}

// tryHold attempts to hold on to the lease by repeatedly refreshing its TTL.
// If the lease hold fails, is deleted, or changed to another user. The provided
// index is used to watch from.
// TODO: currently if we miss the watch window, we will error and try to recreate
// the lock. It's likely we will lose the lease due to that.
func (e *Etcd) tryHold(ttl, index uint64) error {
	// watch for termination
	stop := make(chan struct{})
	lost := make(chan struct{})
	closedLost := false
	watchIndex := index
	go utilwait.Until(func() {
		index, err := e.waitForExpiration(true, watchIndex, stop)
		watchIndex = index
		if err != nil {
			utilruntime.HandleError(fmt.Errorf("error watching for lease expiration %s: %v", e.key, err))
			return
		}
		glog.V(4).Infof("Lease %s lost due to deletion at %d", e.key, watchIndex)
		if !closedLost {
			closedLost = true
			close(lost)
		}
	}, 100*time.Millisecond, stop)
	defer close(stop)

	duration := time.Duration(ttl) * time.Second
	after := time.Duration(float32(duration) * e.waitFraction)
	last := duration - after
	interval := last / time.Duration(e.maxRetries)
	if interval < e.minimumRetryInterval {
		interval = e.minimumRetryInterval
	}

	// as long as we can renew the lease, loop
	for {
		select {
		case <-time.After(after):
			err := wait.Poll(interval, last, func() (bool, error) {
				glog.V(4).Infof("Renewing lease %s at %d", e.key, index-1)
				resp, err := e.keysClient.Set(context.Background(), e.key, e.value,
					&etcdclient.SetOptions{
						TTL:       time.Duration(e.ttl) * time.Second,
						PrevValue: e.value,
						PrevIndex: index - 1,
					},
				)
				switch {
				case err == nil:
					index = eventIndexFor(resp)
					return true, nil
				case etcdutil.IsEtcdTestFailed(err):
					return false, fmt.Errorf("another client has taken the lease %s: %v", e.key, err)
				case etcdutil.IsEtcdNotFound(err):
					return false, fmt.Errorf("another client has revoked the lease %s", e.key)
				default:
					utilruntime.HandleError(fmt.Errorf("unexpected error renewing lease %s: %v", e.key, err))
					index = etcdIndexFor(err, index)
					// try again
					return false, nil
				}
			})

			switch err {
			case nil:
				// wait again
				glog.V(4).Infof("Lease %s renewed at %d", e.key, index-1)
			case wait.ErrWaitTimeout:
				return fmt.Errorf("unable to renew lease %s at %d: %v", e.key, index, err)
			default:
				return fmt.Errorf("lost lease %s at %d: %v", e.key, index, err)
			}

		case <-lost:
			return fmt.Errorf("the lease has been lost %s at %d", e.key, index)
		}
	}
}

// waitForExpiration waits until the lease value changes in etcd through deletion, expiration,
// or explicit change. Held indicates whether the current process owns the lease. The appropriate
// next watch index is returned.
func (e *Etcd) waitForExpiration(held bool, from uint64, stop chan struct{}) (uint64, error) {
	for {
		lost, index, err := e.waitExpiration(held, from, stop)
		if err != nil {
			return index, err
		}
		if lost {
			return index, nil
		}
	}
}

// waitExpiration watches etcd until the lease is deleted, expired, or changed. If the lease is
// held and a change to the value no longer matches the local value, the lease will be considered
// to be lost. If the lease is not held, and the value changes to match our value, we'll consider
// the existing lease to be lost and we are a candidate to acquire it. The appropriate next watch
// index is returned.
func (e *Etcd) waitExpiration(held bool, from uint64, stop chan struct{}) (bool, uint64, error) {
	for {
		select {
		case <-stop:
			return false, from, nil
		default:
		}
		glog.V(5).Infof("watching for expiration of lease %s from %d", e.key, from)
		w := e.keysClient.Watcher(e.key, &etcdclient.WatcherOptions{AfterIndex: from - 1})
		resp, err := w.Next(context.Background())
		if err != nil {
			return false, etcdIndexFor(err, from), err
		}

		index := eventIndexFor(resp)

		if resp.Action == "delete" || resp.Action == "compareAndDelete" || resp.Action == "expire" {
			// the lease has expired
			return true, index, nil
		}

		switch {
		case resp.Node == nil:
		case resp.Node.Value == e.value && !held:
			// given to us
			return true, index, nil
		case resp.Node.Value != e.value && held:
			// taken away from us
			return true, index, nil
		}

		from = index
	}
}

// eventIndexFor returns the next etcd index to watch based on a response
func eventIndexFor(resp *etcdclient.Response) uint64 {
	if resp.Node != nil {
		return resp.Node.ModifiedIndex + 1
	}
	if resp.PrevNode != nil {
		return resp.PrevNode.ModifiedIndex + 1
	}
	return resp.Index
}

// etcdIndexFor returns index, or if err is an EtcdError, the current
// etcd index.
func etcdIndexFor(err error, index uint64) uint64 {
	if etcderr, ok := err.(*etcdclient.Error); ok {
		return etcderr.Index
	}
	return index
}
