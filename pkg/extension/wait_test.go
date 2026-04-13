package extension

import (
	"context"
	"fmt"
	"testing"

	"github.com/mcpchecker/mcpchecker/pkg/extension/sdk"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestHandleWait(t *testing.T) {
	tests := []struct {
		name        string
		args        any
		client      *mockClient
		wantSuccess bool
	}{
		{
			name: "condition already met",
			args: map[string]any{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata":   map[string]any{"name": "nginx", "namespace": "default"},
				"condition":  "Available",
				"timeout":    "1s",
			},
			client: &mockClient{
				getFn: func(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string) (*unstructured.Unstructured, error) {
					return &unstructured.Unstructured{
						Object: map[string]any{
							"status": map[string]any{
								"conditions": []any{
									map[string]any{"type": "Available", "status": "True"},
								},
							},
						},
					}, nil
				},
			},
			wantSuccess: true,
		},
		{
			name: "condition not met within timeout",
			args: map[string]any{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata":   map[string]any{"name": "nginx", "namespace": "default"},
				"condition":  "Available",
				"timeout":    "1s",
			},
			client: &mockClient{
				getFn: func(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string) (*unstructured.Unstructured, error) {
					return &unstructured.Unstructured{
						Object: map[string]any{
							"status": map[string]any{
								"conditions": []any{
									map[string]any{"type": "Available", "status": "False"},
								},
							},
						},
					}, nil
				},
			},
			wantSuccess: false,
		},
		{
			name: "missing condition field checks existence",
			args: map[string]any{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata":   map[string]any{"name": "test"},
				"timeout":    "1s",
			},
			client: &mockClient{
				getFn: func(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string) (*unstructured.Unstructured, error) {
					return &unstructured.Unstructured{
						Object: map[string]any{
							"apiVersion": "v1",
							"kind":       "Pod",
							"metadata":   map[string]any{"name": "test"},
						},
					}, nil
				},
			},
			wantSuccess: true,
		},
		{
			name: "no condition - resource not found times out",
			args: map[string]any{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata":   map[string]any{"name": "test"},
				"timeout":    "1s",
			},
			client: &mockClient{
				getFn: func(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string) (*unstructured.Unstructured, error) {
					return nil, fmt.Errorf("not found")
				},
			},
			wantSuccess: false,
		},
		{
			name: "no condition - succeeds when resource exists",
			args: map[string]any{
				"apiVersion": "networking.istio.io/v1",
				"kind":       "Gateway",
				"metadata":   map[string]any{"name": "my-gateway", "namespace": "istio-system"},
				"timeout":    "2s",
			},
			client: &mockClient{
				getFn: func(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string) (*unstructured.Unstructured, error) {
					return &unstructured.Unstructured{
						Object: map[string]any{
							"apiVersion": "networking.istio.io/v1",
							"kind":       "Gateway",
							"metadata": map[string]any{
								"name":      "my-gateway",
								"namespace": "istio-system",
							},
							"spec": map[string]any{
								"selector": map[string]any{
									"istio": "ingressgateway",
								},
							},
						},
					}, nil
				},
			},
			wantSuccess: true,
		},
		{
			name: "resource without status.conditions times out (e.g. Istio Gateway)",
			args: map[string]any{
				"apiVersion": "networking.istio.io/v1",
				"kind":       "Gateway",
				"metadata":   map[string]any{"name": "my-gateway", "namespace": "istio-system"},
				"condition":  "Available",
				"status":     "True",
				"timeout":    "2s",
			},
			client: &mockClient{
				getFn: func(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string) (*unstructured.Unstructured, error) {
					return &unstructured.Unstructured{
						Object: map[string]any{
							"apiVersion": "networking.istio.io/v1",
							"kind":       "Gateway",
							"metadata": map[string]any{
								"name":      "my-gateway",
								"namespace": "istio-system",
							},
							"spec": map[string]any{
								"selector": map[string]any{
									"istio": "ingressgateway",
								},
							},
						},
					}, nil
				},
			},
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ext := &Extension{
				Extension: sdk.NewExtension(sdk.ExtensionInfo{Name: "test"}),
				client:    tt.client,
			}

			req := &sdk.OperationRequest{Args: tt.args}
			result, err := ext.handleWait(context.Background(), req)

			if err != nil {
				t.Fatalf("handleWait() returned error: %v", err)
			}
			if result.Success != tt.wantSuccess {
				t.Errorf("handleWait() success = %v, want %v", result.Success, tt.wantSuccess)
			}
		})
	}
}
