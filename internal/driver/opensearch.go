package driver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type OpenSearchDriver struct {
	endpoint string
	client   *http.Client
}

type OpenSearchClusterHealth struct {
	ClusterName                 string  `json:"cluster_name"`
	Status                      string  `json:"status"`
	TimedOut                    bool    `json:"timed_out"`
	NumberOfNodes               int     `json:"number_of_nodes"`
	NumberOfDataNodes           int     `json:"number_of_data_nodes"`
	ActivePrimaryShards         int     `json:"active_primary_shards"`
	ActiveShards                int     `json:"active_shards"`
	RelocatingShards            int     `json:"relocating_shards"`
	InitializingShards          int     `json:"initializing_shards"`
	UnassignedShards            int     `json:"unassigned_shards"`
	DelayedUnassignedShards     int     `json:"delayed_unassigned_shards"`
	NumberOfPendingTasks        int     `json:"number_of_pending_tasks"`
	NumberOfInFlightFetch       int     `json:"number_of_in_flight_fetch"`
	TaskMaxWaitingInQueueMillis int     `json:"task_max_waiting_in_queue_millis"`
	ActiveShardsPercentAsNumber float64 `json:"active_shards_percent_as_number"`
}

func NewOpenSearchDriver(endpoint string) (Driver, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("endpoint cannot be empty")
	}

	endpoint = strings.TrimSuffix(endpoint, "/")

	return &OpenSearchDriver{
		endpoint: endpoint,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (o *OpenSearchDriver) Check() (*CheckResult, error) {
	start := time.Now()

	healthURL := o.endpoint + "/_cluster/health"

	resp, err := o.client.Get(healthURL)
	duration := time.Since(start)

	result := &CheckResult{
		ResponseTime: duration,
	}

	if err != nil {
		result.Success = false
		result.Error = err
		result.Message = fmt.Sprintf("OpenSearch check failed: %v", err)
		return result, nil
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		result.Success = false
		result.Message = fmt.Sprintf("OpenSearch check failed (status: %d, response time: %v)", resp.StatusCode, duration)
		return result, nil
	}

	var health OpenSearchClusterHealth
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		result.Success = false
		result.Error = err
		result.Message = fmt.Sprintf("OpenSearch check failed to parse response: %v", err)
		return result, nil
	}

	switch strings.ToLower(health.Status) {
	case "green":
		result.Success = true
		result.Message = fmt.Sprintf("OpenSearch cluster is healthy (status: %s, nodes: %d, active_shards: %d, response time: %v)",
			health.Status, health.NumberOfNodes, health.ActiveShards, duration)
	case "yellow":
		result.Success = false
		result.Message = fmt.Sprintf("OpenSearch cluster has issues (status: %s, nodes: %d, unassigned_shards: %d, response time: %v)",
			health.Status, health.NumberOfNodes, health.UnassignedShards, duration)
	case "red":
		result.Success = false
		result.Message = fmt.Sprintf("OpenSearch cluster is unhealthy (status: %s, nodes: %d, unassigned_shards: %d, response time: %v)",
			health.Status, health.NumberOfNodes, health.UnassignedShards, duration)
	default:
		result.Success = false
		result.Message = fmt.Sprintf("OpenSearch cluster status unknown (status: %s, response time: %v)",
			health.Status, duration)
	}

	return result, nil
}

func (o *OpenSearchDriver) GetEndpoint() string {
	return o.endpoint
}

func (o *OpenSearchDriver) GetType() string {
	return "opensearch"
}
