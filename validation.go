package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"encoding/json"
	"net/http"
)

var allowedHttpMethod = []string{
	http.MethodGet,
	http.MethodPatch,
	http.MethodPut,
	http.MethodPost,
	http.MethodDelete,
}

func ValidateEnv() []error {
	var errs []error
	if err := validateTargetUrl(); err != nil {
		errs = append(errs, err)
	}
	if err := validateReqHttpMethodRatio(); err != nil {
		errs = append(errs, err)
	}
	if err := validateHttpHeaders(); err != nil {
		errs = append(errs, err)
	}
	if err := validateHttpRequestBody(); err != nil {
		errs = append(errs, err)
	}
	if err := validateThreadNum(); err != nil {
		errs = append(errs, err)
	}
	if err := validateTrialNum(); err != nil {
		errs = append(errs, err)
	}
	if err := validatePermanent(); err != nil {
		errs = append(errs, err)
	}
	if err := validateAlert(); err != nil {
		errs = append(errs, err)
	}
	return errs
}

// Validate TARGET_URL env
func validateTargetUrl() error {
	env := os.Getenv(EnvTargetUrl)
	if validateEmpty(env) {
		return errors.New(fmt.Sprintf("Environment valiable %s is required.", EnvTargetUrl))
	}
	if !strings.HasPrefix(env, "http") || !strings.HasPrefix(env, "https") {
		return errors.New(fmt.Sprintf("Environment valiable %s has only http or https protocol.", EnvTargetUrl))
	}
	return nil
}

// Validate HTTP_HEADERS env
func validateHttpHeaders() error {
	env := os.Getenv(EnvHttpHeaders)
	if validateEmpty(env) {
		return errors.New(fmt.Sprintf("Environment valiable %s is required.", EnvHttpHeaders))
	}
	headers := make(map[string]interface{})
	if err := json.Unmarshal([]byte(env), &headers); err != nil {
		return errors.New(fmt.Sprintf("Environment valiable %s not hashmap structure.", EnvHttpHeaders))
	}
	return nil
}

// Validate REQ_HTTP_METHOD_RATIO env
func validateReqHttpMethodRatio() error {
	env := os.Getenv(EnvReqHttpMethodRatio)
	if validateEmpty(env) {
		return errors.New(fmt.Sprintf("Environment valiable %s is required.", EnvReqHttpMethodRatio))
	}
	percentages := make(map[string]int)
	if err := json.Unmarshal([]byte(env), &percentages); err != nil {
		return errors.New(fmt.Sprintf("Environment valiable %s not hashmap structure.", EnvReqHttpMethodRatio))
	}

	var totalPercent int
	for _, v := range percentages {
		totalPercent = totalPercent + v
	}
	if totalPercent != 10 {
		return errors.New(fmt.Sprintf("Environment valiable %s requires percentage of 10.", EnvReqHttpMethodRatio))
	}
	return nil
}

// Validate HTTP_REQ_BODY env
func validateHttpRequestBody() error {
	env := os.Getenv(EnvHttpRequestBody)
	ratio := os.Getenv(EnvReqHttpMethodRatio)
	if strings.Contains(ratio, http.MethodPost) || strings.Contains(ratio, http.MethodPut) || strings.Contains(ratio, http.MethodPatch){
		if validateEmpty(env) {
			return errors.New(fmt.Sprintf("Environment valiable %s is required.", EnvHttpRequestBody))
		}
		body := make(map[string]interface{})
		if err := json.Unmarshal([]byte(env), &body); err != nil {
			return errors.New(fmt.Sprintf("Environment valiable %s not hashmap structure.", EnvHttpRequestBody))
		}
	}
	return nil
}

// Validate THREAD_NUM env
func validateThreadNum() error {
	env := os.Getenv(EnvThreadNum)
	if validateEmpty(env) {
		return errors.New(fmt.Sprintf("Environment valiable %s is required.", EnvThreadNum))
	}
	if _, err := strconv.Atoi(env); err != nil {
		return errors.New(fmt.Sprintf("Environment valiable %s is not number.", EnvThreadNum))
	}
	return nil
}

// Validate TRIAL_NUM env
func validateTrialNum() error {
	env := os.Getenv(EnvTrialNum)
	if validateEmpty(env) {
		return errors.New(fmt.Sprintf("Environment valiable %s is required.", EnvTrialNum))
	}
	if _, err := strconv.Atoi(env); err != nil {
		return errors.New(fmt.Sprintf("Environment valiable %s is not number.", EnvTrialNum))
	}
	return nil
}

// Validate PERMANENT env
// Optional
func validatePermanent() error {
	env := os.Getenv(EnvPermanent)
	if validateEmpty(env) {
		return nil
	}
	if !strings.EqualFold(env, "true") || !strings.EqualFold(env, "false") {
		return errors.New(fmt.Sprintf("Environment valiable %s is true or false.", EnvPermanent))
	}
	return nil
}

// Validate ENABLE_ALERT,SLACK_NOTIFY_THRESHOLD_LATENCY_MILLIS,SLACK_NOTIFY_THRESHOLD_RPS env
// Optional
func validateAlert() error {
	env := os.Getenv(EnvEnableAlert)
	if validateEmpty(env) {
		return nil
	}

	isEnable, err := strconv.ParseBool(env)
	if err != nil {
		return errors.New(fmt.Sprintf("Environment valiable %s is true or false.", EnvEnableAlert))
	}

	if !isEnable {
		return nil
	}

	latencyErr := validateSlackNotifyThreshHoldLatencyMillis()
	rpsErr := validateSlackNotifyThreshHoldRps()
	if latencyErr != nil && rpsErr != nil{
		return errors.New(fmt.Sprintf("Environment valiable %s or %s is required.", EnvSlackNotifyThreshHoldLatencyMillis, EnvSlackNotifyThreshHoldRps))
	}
	return nil
}

// Validate SLACK_NOTIFY_THRESHOLD_LATENCY_MILLIS env
// Optional
func validateSlackNotifyThreshHoldLatencyMillis() error {
	env := os.Getenv(EnvSlackNotifyThreshHoldLatencyMillis)
	if validateEmpty(env) {
		return nil
	}
	if _, err := strconv.Atoi(env); err != nil {
		return errors.New(fmt.Sprintf("Environment valiable %s is not number.", EnvSlackNotifyThreshHoldLatencyMillis))
	}
	return nil
}

// Validate SLACK_NOTIFY_THRESHOLD_RPS env
// Optional
func validateSlackNotifyThreshHoldRps() error {
	env := os.Getenv(EnvSlackNotifyThreshHoldRps)
	if validateEmpty(env) {
		return nil
	}
	if _, err := strconv.Atoi(env); err != nil {
		return errors.New(fmt.Sprintf("Environment valiable %s is not number.", EnvSlackNotifyThreshHoldRps))
	}
	return nil
}

// Validate empty
func validateEmpty(data string) bool {
	if data == "" {
		return true
	}
	return false
}
