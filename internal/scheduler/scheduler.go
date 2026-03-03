package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"sync"
	"time"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/clock"
	ctrl "sigs.k8s.io/controller-runtime"

	monitoringv1alpha1 "github.com/LiciousTech/endpoint-monitoring-operator/api/v1alpha1"
	"github.com/LiciousTech/endpoint-monitoring-operator/internal/driver"
	"github.com/LiciousTech/endpoint-monitoring-operator/internal/notifier"
)

const defaultWorkers = 20

// DriverFactory creates a monitoring driver for a monitor spec.
type DriverFactory interface {
	CreateDriver(driverType string, endpoint string, monitor *monitoringv1alpha1.EndpointMonitor) (driver.Driver, error)
}

// NotifierFactory creates a notifier for a monitor spec.
type NotifierFactory interface {
	CreateNotifier(config *monitoringv1alpha1.NotifyConfig) (notifier.Notifier, error)
}

// ProbeResult contains the persisted result for a probe execution.
type ProbeResult struct {
	Success      bool
	Latency      time.Duration
	StatusCode   int
	ErrorMessage string
	CheckedAt    metav1.Time
}

// StatusWriter writes probe execution outcomes to EndpointMonitor status.
type StatusWriter interface {
	Write(ctx context.Context, key types.NamespacedName, result ProbeResult) error
}

// ProbeScheduler registers probes and manages background execution.
type ProbeScheduler interface {
	Upsert(key types.NamespacedName, spec monitoringv1alpha1.EndpointMonitorSpec)
	Delete(key types.NamespacedName)
	Start(ctx context.Context) error
}

// ProbeJob describes a single scheduled probe run.
type ProbeJob struct {
	Key      types.NamespacedName
	Driver   driver.Driver
	Notifier notifier.Notifier
	Status   StatusWriter
}

// Config configures scheduler worker concurrency and queue size.
type Config struct {
	Workers   int
	QueueSize int
}

type probeEntry struct {
	spec   monitoringv1alpha1.EndpointMonitorSpec
	hash   uint64
	cancel context.CancelFunc
	ticker clock.Ticker
}

// Scheduler runs timers and workers for all registered endpoint monitors.
type Scheduler struct {
	cfg             Config
	driverFactory   DriverFactory
	notifierFactory NotifierFactory
	statusWriter    StatusWriter
	clk             clock.WithTicker

	mu      sync.Mutex
	entries map[types.NamespacedName]*probeEntry
	jobs    chan ProbeJob
	logger  logr.Logger
}

// New creates a scheduler with defaults and production clock.
func New(cfg Config, driverFactory DriverFactory, notifierFactory NotifierFactory, statusWriter StatusWriter) *Scheduler {
	if cfg.Workers <= 0 {
		cfg.Workers = defaultWorkers
	}
	if cfg.QueueSize <= 0 {
		cfg.QueueSize = cfg.Workers * 4
	}

	return newWithClock(cfg, driverFactory, notifierFactory, statusWriter, clock.RealClock{})
}

func newWithClock(cfg Config, driverFactory DriverFactory, notifierFactory NotifierFactory, statusWriter StatusWriter, clk clock.WithTicker) *Scheduler {
	return &Scheduler{
		cfg:             cfg,
		driverFactory:   driverFactory,
		notifierFactory: notifierFactory,
		statusWriter:    statusWriter,
		clk:             clk,
		entries:         make(map[types.NamespacedName]*probeEntry),
		jobs:            make(chan ProbeJob, cfg.QueueSize),
		logger:          ctrl.Log.WithName("probe-scheduler"),
	}
}

// NeedLeaderElection indicates scheduler must run only on elected leader.
func (s *Scheduler) NeedLeaderElection() bool {
	return true
}

// Start runs worker pool until the manager context is canceled.
func (s *Scheduler) Start(ctx context.Context) error {
	var workers sync.WaitGroup
	workers.Add(s.cfg.Workers)
	for i := 0; i < s.cfg.Workers; i++ {
		go func(workerID int) {
			defer workers.Done()
			s.workerLoop(ctx, workerID)
		}(i)
	}

	<-ctx.Done()

	s.mu.Lock()
	for _, entry := range s.entries {
		entry.cancel()
		entry.ticker.Stop()
	}
	s.entries = map[types.NamespacedName]*probeEntry{}
	s.mu.Unlock()

	workers.Wait()
	return nil
}

// Upsert registers or updates scheduling for a monitor key/spec.
func (s *Scheduler) Upsert(key types.NamespacedName, spec monitoringv1alpha1.EndpointMonitorSpec) {
	hash, err := hashSpec(spec)
	if err != nil {
		s.logger.Error(err, "unable to hash spec", "key", key)
		return
	}

	interval := time.Duration(spec.CheckInterval) * time.Second
	if interval <= 0 {
		interval = time.Second
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if existing, ok := s.entries[key]; ok {
		if existing.hash == hash {
			return
		}
		existing.cancel()
		existing.ticker.Stop()
	}

	probeCtx, cancel := context.WithCancel(context.Background())
	ticker := s.clk.NewTicker(interval)
	entry := &probeEntry{spec: spec, hash: hash, cancel: cancel, ticker: ticker}
	s.entries[key] = entry

	go s.dispatchProbeLoop(probeCtx, key, entry)
}

// Delete cancels and removes an existing scheduled monitor.
func (s *Scheduler) Delete(key types.NamespacedName) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.entries[key]
	if !ok {
		return
	}
	entry.cancel()
	entry.ticker.Stop()
	delete(s.entries, key)
}

func (s *Scheduler) dispatchProbeLoop(ctx context.Context, key types.NamespacedName, entry *probeEntry) {
	s.enqueueJob(key, entry.spec)
	for {
		select {
		case <-ctx.Done():
			return
		case <-entry.ticker.C():
			s.enqueueJob(key, entry.spec)
		}
	}
}

func (s *Scheduler) enqueueJob(key types.NamespacedName, spec monitoringv1alpha1.EndpointMonitorSpec) {
	monitor := &monitoringv1alpha1.EndpointMonitor{Spec: spec}
	drv, err := s.driverFactory.CreateDriver(spec.Driver, spec.Endpoint, monitor)
	if err != nil {
		s.logger.Error(err, "unable to create driver", "key", key)
		return
	}

	n, err := s.notifierFactory.CreateNotifier(&spec.Notify)
	if err != nil {
		s.logger.Error(err, "unable to create notifier", "key", key)
		return
	}

	job := ProbeJob{Key: key, Driver: drv, Notifier: n, Status: s.statusWriter}

	select {
	case s.jobs <- job:
	default:
		s.logger.Info("dropping probe job because queue is full", "key", key)
	}
}

func (s *Scheduler) workerLoop(ctx context.Context, workerID int) {
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-s.jobs:
			s.runJob(ctx, workerID, job)
		}
	}
}

func (s *Scheduler) runJob(ctx context.Context, workerID int, job ProbeJob) {
	logger := s.logger.WithValues("worker", workerID, "key", job.Key)
	defer func() {
		if recovered := recover(); recovered != nil {
			logger.Error(fmt.Errorf("panic: %v", recovered), "recovered panic in probe worker")
		}
	}()

	checkResult, err := job.Driver.Check()
	if err != nil {
		checkResult = &driver.CheckResult{Success: false, Error: err, Message: err.Error()}
	}

	status := "failure"
	if checkResult.Success {
		status = "success"
	}

	message := fmt.Sprintf("%s monitor for %s: %s", job.Driver.GetType(), job.Driver.GetEndpoint(), checkResult.Message)
	if notifyErr := job.Notifier.SendAlert(status, message); notifyErr != nil {
		logger.Error(notifyErr, "failed to notify")
	}

	probeResult := ProbeResult{Success: checkResult.Success, Latency: checkResult.ResponseTime, CheckedAt: metav1.Now()}
	if checkResult.Error != nil {
		probeResult.ErrorMessage = checkResult.Error.Error()
	}

	if err := job.Status.Write(ctx, job.Key, probeResult); err != nil {
		logger.Error(err, "failed to write probe status")
	}
}

func hashSpec(spec monitoringv1alpha1.EndpointMonitorSpec) (uint64, error) {
	b, err := json.Marshal(spec)
	if err != nil {
		return 0, err
	}
	h := fnv.New64a()
	if _, err := h.Write(b); err != nil {
		return 0, err
	}
	return h.Sum64(), nil
}
