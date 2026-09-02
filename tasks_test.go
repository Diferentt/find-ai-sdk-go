package findai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestInvokeTaskNestsParamsUnderInitialState(t *testing.T) {
	// Top-level params (instead of initial_state.params) is the mistake the
	// backend rejects with 422; the SDK must make it unrepresentable.
	var gotBody map[string]any
	c := newTestServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/scheduled-jobs/task_1/invoke" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("method = %q", r.Method)
		}
		raw, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(raw, &gotBody); err != nil {
			t.Fatalf("request body is not JSON: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"job_id": "task_1", "success": true, "duration_ms": 2100,
			"output": {"traduccion": "hello world"},
			"final_state": null,
			"nodes_executed": ["start", "generate", "end"],
			"error": null
		}`))
	})

	res, err := c.InvokeTask(context.Background(), "task_1", map[string]any{"texto": "hola mundo", "idioma": "en"})
	if err != nil {
		t.Fatalf("InvokeTask() error = %v", err)
	}

	state, ok := gotBody["initial_state"].(map[string]any)
	if !ok {
		t.Fatalf("body has no initial_state object: %v", gotBody)
	}
	params, ok := state["params"].(map[string]any)
	if !ok || params["texto"] != "hola mundo" || params["idioma"] != "en" {
		t.Fatalf("params not nested under initial_state.params: %v", gotBody)
	}
	if _, leaked := gotBody["params"]; leaked {
		t.Fatalf("params leaked to the top level of the body: %v", gotBody)
	}

	if !res.Success || res.Output["traduccion"] != "hello world" || res.DurationMS != 2100 {
		t.Fatalf("unexpected result: %+v", res)
	}
	if res.FinalState != nil {
		t.Fatalf("FinalState should be nil when output is present: %+v", res.FinalState)
	}
}

func TestInvokeTaskWithNilParamsSendsEmptyBody(t *testing.T) {
	c := newTestServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		// initial_state carries omitempty: a defaults-only run must not send
		// {"initial_state": null}, which is not the same as absent for a
		// strict backend.
		if string(raw) != "{}" {
			t.Errorf("body = %q, want {}", raw)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"job_id": "task_1", "success": true, "duration_ms": 10, "nodes_executed": []}`))
	})

	if _, err := c.InvokeTask(context.Background(), "task_1", nil); err != nil {
		t.Fatalf("InvokeTask() error = %v", err)
	}
}

func TestInvokeTaskFlowFailureIsSuccessFalseNotError(t *testing.T) {
	c := newTestServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"job_id": "task_1", "success": false, "duration_ms": 900,
			"output": null, "final_state": {"system": {}},
			"nodes_executed": ["start", "generate"],
			"error": "Node generate (generate) failed: boom"
		}`))
	})

	res, err := c.InvokeTask(context.Background(), "task_1", nil)
	if err != nil {
		t.Fatalf("a flow that ran and failed is a 200, got err = %v", err)
	}
	if res.Success || res.Error == nil {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestInvokeTaskDisabledIsConflict(t *testing.T) {
	c := newTestServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"detail": "Task is disabled"}`))
	})

	_, err := c.InvokeTask(context.Background(), "task_1", nil)
	if !IsConflict(err) {
		t.Fatalf("IsConflict(err) = false, err = %v", err)
	}
}

func TestGetTaskReportsInputs(t *testing.T) {
	c := newTestServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/scheduled-jobs/task_1" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "task_1", "tenant_id": "t_1", "name": "Traducir texto",
			"flow_id": "flow_translate", "version_id": null,
			"schedule_type": null, "schedule_expression": null,
			"timezone": "UTC", "initial_state": {"params": {"idioma": "en"}},
			"enabled": true,
			"created_at": "2026-09-01T00:00:00Z", "updated_at": "2026-09-01T00:00:00Z",
			"last_run_at": null, "last_run_status": null,
			"eventbridge_status": "not_applicable", "eventbridge_error": null,
			"warnings": [], "inputs": ["idioma", "texto"]
		}`))
	})

	task, err := c.GetTask(context.Background(), "task_1")
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}
	if task.Name != "Traducir texto" || task.ScheduleType != nil || !task.Enabled {
		t.Fatalf("unexpected task: %+v", task)
	}
	if len(task.Inputs) != 2 || task.Inputs[0] != "idioma" {
		t.Fatalf("Inputs = %v, want [idioma texto]", task.Inputs)
	}
}
