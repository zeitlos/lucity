package main

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

const (
	leaseName          = "lucity-conductor"
	leaseDuration      = 15 * time.Second
	leaseRenewDeadline = 10 * time.Second
	leaseRetryPeriod   = 2 * time.Second
	loopStopTimeout    = 10 * time.Second
)

func runAsLeader(ctx context.Context, client kubernetes.Interface, namespace string, loops ...func(context.Context)) (<-chan struct{}, error) {
	identity, err := os.Hostname()

	if err != nil {
		return nil, err
	}

	var mu sync.Mutex
	var running sync.WaitGroup

	electionCtx, stopElection := context.WithCancel(context.Background())

	election := leaderelection.LeaderElectionConfig{
		Lock: &resourcelock.LeaseLock{
			LeaseMeta:  metav1.ObjectMeta{Name: leaseName, Namespace: namespace},
			Client:     client.CoordinationV1(),
			LockConfig: resourcelock.ResourceLockConfig{Identity: identity},
		},
		Name:            leaseName,
		LeaseDuration:   leaseDuration,
		RenewDeadline:   leaseRenewDeadline,
		RetryPeriod:     leaseRetryPeriod,
		ReleaseOnCancel: true,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(leaderCtx context.Context) {
				mu.Lock()
				defer mu.Unlock()

				running.Wait()

				if ctx.Err() != nil {
					return
				}

				loopCtx, stopLoops := context.WithCancel(leaderCtx)
				context.AfterFunc(ctx, stopLoops)

				slog.Info("leading background loops", "identity", identity)

				for _, loop := range loops {
					running.Go(func() { loop(loopCtx) })
				}
			},
			OnStoppedLeading: func() {},
			OnNewLeader: func(leader string) {
				if leader != "" && leader != identity {
					slog.Info("following conductor leader", "leader", leader)
				}
			},
		},
	}

	context.AfterFunc(ctx, func() {
		mu.Lock()
		defer mu.Unlock()

		if !stoppedWithin(&running, loopStopTimeout) {
			slog.Warn("background loops still running at shutdown", "timeout", loopStopTimeout)
		}

		stopElection()
	})

	done := make(chan struct{})

	go func() {
		defer close(done)

		for electionCtx.Err() == nil {
			leaderelection.RunOrDie(electionCtx, election)
		}
	}()

	return done, nil
}

func stoppedWithin(running *sync.WaitGroup, timeout time.Duration) bool {
	stopped := make(chan struct{})

	go func() {
		running.Wait()
		close(stopped)
	}()

	select {
	case <-stopped:
		return true
	case <-time.After(timeout):
		return false
	}
}
