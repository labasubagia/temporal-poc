package activities

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"runtime"
	"strings"
)

const (
	STATUS_STARTED   = "started"
	STATUS_COMPLETED = "completed"
	STATUS_FAILED    = "failed"
)


func NotifyProgress(ctx context.Context, workflowID, fn any, status string, totalActivities int) error {
	activityName := GetFunctionName(fn)
	payload := map[string]interface{}{
		"workflow_id":      workflowID,
		"activity":        activityName,
		"status":          status,
		"total_activities": totalActivities,
	}
	body, _ := json.Marshal(payload)
	_, _ = http.Post("http://localhost:8081/ws/notify", "application/json", bytes.NewReader(body))
	return nil
}

func GetFunctionName(i any) string {
	if fullName, ok := i.(string); ok {
		return fullName
	}
	fullName := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
	// Full function name that has a struct pointer receiver has the following format
	// <prefix>.(*<type>).<function>
	elements := strings.Split(fullName, ".")
	shortName := elements[len(elements)-1]
	// This allows to call activities by method pointer
	// Compiler adds -fm suffix to a function name which has a receiver
	// Note that this works even if struct pointer used to get the function is nil
	// It is possible because nil receivers are allowed.
	// For example:
	// var a *Activities
	// ExecuteActivity(ctx, a.Foo)
	// will call this function which is going to return "Foo"
	return strings.TrimSuffix(shortName, "-fm")
}