package findai

import "context"

// invokeTaskBody is the request body for InvokeTask. initial_state is the
// only field the endpoint accepts — anything else is rejected with 422 —
// so the SDK builds it structurally instead of passing a caller map through.
type invokeTaskBody struct {
	InitialState map[string]any `json:"initial_state,omitempty"`
}

// InvokeTask runs a task (a flow executed without a conversation) and waits
// for its result. params are the task's input parameters: the SDK nests them
// under initial_state.params, which is where the flow's nodes read them as
// {{params.key}}. Call with nil params to run the task with the defaults
// configured on it.
//
// Which parameters a task accepts is reported by GetTask (the Inputs field).
//
// The call is synchronous: it returns when the flow finishes, with the result
// in TaskInvokeResponse.Output (what the flow wrote under the "output" state
// bucket). Requires an API key with the "flows:execute" scope. A disabled
// task responds 409 (IsConflict); an unknown task — or one belonging to
// another tenant — responds 404 (IsNotFound).
func (c *Client) InvokeTask(ctx context.Context, taskID string, params map[string]any) (*TaskInvokeResponse, error) {
	var state map[string]any
	if len(params) > 0 {
		state = map[string]any{"params": params}
	}
	return c.InvokeTaskWithState(ctx, taskID, state)
}

// InvokeTaskWithState is InvokeTask with control over the full initial_state,
// for flows that read buckets other than params. The state is shallow-merged
// over the initial_state configured on the task (params combine key by key;
// other buckets replace). Most callers want InvokeTask.
func (c *Client) InvokeTaskWithState(ctx context.Context, taskID string, initialState map[string]any) (*TaskInvokeResponse, error) {
	var out TaskInvokeResponse
	path := "/api/v1/scheduled-jobs/" + taskID + "/invoke"
	if err := c.t.Do(ctx, "POST", path, nil, invokeTaskBody{InitialState: initialState}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetTask returns one task, including Inputs: the parameter names it accepts
// in InvokeTask, derived from the flow's nodes (so they follow the flow — no
// schema is declared on the task itself).
//
// The task API is served under "scheduled-jobs" paths for historical reasons:
// a task may also carry a cron schedule, but neither trigger requires the
// other, and the product calls both kinds "tasks".
func (c *Client) GetTask(ctx context.Context, taskID string) (*TaskResponse, error) {
	var out TaskResponse
	path := "/api/v1/scheduled-jobs/" + taskID
	if err := c.t.Do(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
