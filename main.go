package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"net/http"
)

const (
	// =========================================================================
	// ============================== HTTP Request =============================
	// =========================================================================

	// EnvTargetUrl Required
	// ex: https://example.com
	EnvTargetUrl = "INPUT_TARGET_URL"

	// EnvHttpHeaders Required
	// HashMap data structure
	// ex: {"Authorization": "Bearer ", "Content-Type": "application/json"}
	EnvHttpHeaders = "INPUT_HTTP_HEADERS"

	// EnvThreadNum Required
	// Maximum is 3.
	// ex: 2
	EnvThreadNum = "INPUT_THREAD_NUM"

	// EnvTrialNum Required
	// Maximum is 20. Takes up to 5 minutes
	// ex: 20
	EnvTrialNum = "INPUT_TRIAL_NUM"

	// EnvReqHttpMethodRatio Optional
	// HashMap data structure
	// If only one http method, always 100 percent set method
	// ex: {"POST": 4, "GET": 6}
	EnvReqHttpMethodRatio = "INPUT_REQ_HTTP_METHOD_RATIO"

	// EnvHttpRequestBody Optional
	// If not empty, always use body when not GET method
	// ex: {"email": "test@gmail.com", "password": "A_test12345-"}
	EnvHttpRequestBody = "INPUT_HTTP_REQ_BODY"

	// =========================================================================
	// ============================== Report data ==============================
	// =========================================================================

	// EnvPermanent Optional
	// Using GitHub pages
	// ex: true || false
	EnvPermanent = "INPUT_PERMANENT"

	// =========================================================================
	// ============================== Slack Alert ==============================
	// =========================================================================

	// EnvEnableAlert Optional
	// ex: true | false
	EnvEnableAlert = "INPUT_ENABLE_ALERT"

	// EnvSlackWebHookUrl Optional
	// ex: https://slack.com
	EnvSlackWebHookUrl = "INPUT_SLACK_WEB_HOOK_URL"

	// EnvSlackChannel Optional
	// ex: #test
	EnvSlackChannel = "INPUT_SLACK_CHANNEL"

	// EnvSlackNotifyThreshHoldLatencyMillis Optional
	// If set this one, notify slack when latency is not achieve threshold
	// ex: 200
	EnvSlackNotifyThreshHoldLatencyMillis = "INPUT_SLACK_NOTIFY_THRESHOLD_LATENCY_MILLIS"

	// EnvSlackNotifyThreshHoldRps Optional
	// If set this one, notify slack when The request per seconds is not achieve threshold
	// ex: 20
	EnvSlackNotifyThreshHoldRps = "INPUT_SLACK_NOTIFY_THRESHOLD_RPS"
)

var (
	allowedHttpMethods = []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodPatch, http.MethodDelete}
)

func main() {
	errs := ValidateEnv()
	if errs != nil {
		for _, v := range errs {
			fmt.Println(v)
		}
		os.Exit(1)
		return
	}

	runtime := NewRuntimeInfo()
	client := NewBenchmarkClient(
		runtime.TargetUrl,
		runtime.HttpMethods,
		runtime.HttpHeaders,
		runtime.HttpRequestBody,
		runtime.HttpRequestMethodRatio,
	)
	calculator := NewCalculator(runtime.TrialNum)

	// Worm up
	client.Warmup()

	// Benchmarking
	startTime := time.Now().UTC()
	fmt.Println(fmt.Sprintf("Start benchmarking. time = %d\n", startTime.Unix()))
	for i := 1; i <= runtime.TrialNum; i++ {
		var wg sync.WaitGroup
		var result Result
		var mutex = &sync.Mutex{}
		for index := 1; index <= runtime.ThreadNum; index++ {
			wg.Add(1)
			go func(index int) {
				data := client.Attack(index)
				mutex.Lock()
				result.Get = append(result.Get, data.Get...)
				result.Post = append(result.Post, data.Post...)
				result.Put = append(result.Put, data.Put...)
				result.Patch = append(result.Patch, data.Patch...)
				result.Delete = append(result.Delete, data.Delete...)
				result.ErrData = calculator.CalculateMethodErrors(result.ErrData, data.ErrData)
				mutex.Unlock()
				wg.Done()
			}(index)
		}
		wg.Wait()

		// Calculate metrics per trial
		fmt.Println()
		calculator.CalculatePerTrial(result.Get, http.MethodGet, i, result.ErrData[http.MethodGet])
		calculator.CalculatePerTrial(result.Post, http.MethodPost, i, result.ErrData[http.MethodPost])
		calculator.CalculatePerTrial(result.Put, http.MethodPut, i, result.ErrData[http.MethodPut])
		calculator.CalculatePerTrial(result.Patch, http.MethodPatch, i, result.ErrData[http.MethodPatch])
		calculator.CalculatePerTrial(result.Delete, http.MethodDelete, i, result.ErrData[http.MethodDelete])

		time.Sleep(1 * time.Second)
	}
	endTime := time.Now().UTC()
	fmt.Println(fmt.Sprintf("End benchmarking. time = %d", endTime.Unix()))

	// Graph
	metrics := calculator.GetMetricsResult()
	graph := NewGraph(metrics)
	charts := graph.GenerateCharts(metrics.TimeRange)
	graph.Output(charts, startTime, endTime)

	// Alert
	if runtime.EnableAlert {
		fmt.Println("\nCheck alert threshold")
		if runtime.SlackNotifyThreshHoldLatencyMillis != 0 {
			for _, method := range allowedHttpMethods {
				// TODO: Choose percentile type
				if calculator.IsOverThreshHold(PercentileAvg, runtime.SlackNotifyThreshHoldLatencyMillis, method) {
					NewSlack(runtime.SlackWebHookUrl, runtime.SlackChannel).NotifyAlert(PercentileAvg)
					return
				}
			}
		}
		if runtime.SlackNotifyThreshHoldRps != 0 {
			for _, method := range allowedHttpMethods {
				if calculator.IsOverThreshHold(Rps, runtime.SlackNotifyThreshHoldRps, method) {
					NewSlack(runtime.SlackWebHookUrl, runtime.SlackChannel).NotifyAlert(PercentileAvg)
					return
				}
			}
		}
	}
}
