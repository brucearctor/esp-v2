// Copyright 2019 Google Cloud Platform Proxy Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bootstrap

import (
	"flag"
	"reflect"
	"strings"
	"testing"

	"github.com/golang/protobuf/ptypes"

	opencensuspb "github.com/census-instrumentation/opencensus-proto/gen-go/trace/v1"
	tracepb "github.com/envoyproxy/data-plane-api/api/trace"
)

const (
	fakeFlagProjectId      = "fake-flag-project-id"
	fakeStackdriverAddress = "dns:non-existent-address:2840"
)

func TestTracingConfig(t *testing.T) {
	testData := []struct {
		desc             string
		flags            map[string]string
		tracingProjectId string
		wantError        string
		wantResult       *tracepb.OpenCensusConfig
	}{
		{
			desc:             "Success with default tracing",
			flags:            map[string]string{},
			tracingProjectId: fakeFlagProjectId,
			wantResult: &tracepb.OpenCensusConfig{
				TraceConfig: &opencensuspb.TraceConfig{
					Sampler: &opencensuspb.TraceConfig_ProbabilitySampler{
						ProbabilitySampler: &opencensuspb.ProbabilitySampler{
							SamplingProbability: *TracingSamplingRate,
						},
					},
				},
				StackdriverExporterEnabled: true,
				StackdriverProjectId:       fakeFlagProjectId,
			},
		},
		{
			desc: "Failed with invalid tracing_incoming_context",
			flags: map[string]string{
				"tracing_incoming_context": "aaa",
			},
			tracingProjectId: fakeFlagProjectId,
			wantError:        "Invalid trace context: aaa",
		},
		{
			desc: "Failed with invalid tracing_outgoing_context",
			flags: map[string]string{
				"tracing_incoming_context": "",
				"tracing_outgoing_context": "bbb",
			},
			tracingProjectId: fakeFlagProjectId,
			wantError:        "Invalid trace context: bbb",
		},
		{
			desc: "Success with some tracing contexts",
			flags: map[string]string{
				"tracing_incoming_context": "traceparent,grpc-trace-bin",
				"tracing_outgoing_context": "x-cloud-trace-context",
			},
			tracingProjectId: fakeFlagProjectId,
			wantResult: &tracepb.OpenCensusConfig{
				TraceConfig: &opencensuspb.TraceConfig{
					Sampler: &opencensuspb.TraceConfig_ProbabilitySampler{
						ProbabilitySampler: &opencensuspb.ProbabilitySampler{
							SamplingProbability: 0.001,
						},
					},
				},
				StackdriverExporterEnabled: true,
				StackdriverProjectId:       fakeFlagProjectId,
				IncomingTraceContext: []tracepb.OpenCensusConfig_TraceContext{
					tracepb.OpenCensusConfig_TRACE_CONTEXT,
					tracepb.OpenCensusConfig_GRPC_TRACE_BIN,
				},
				OutgoingTraceContext: []tracepb.OpenCensusConfig_TraceContext{
					tracepb.OpenCensusConfig_CLOUD_TRACE_CONTEXT,
				},
			},
		},
		{
			desc: "Failed with invalid sampling rate",
			flags: map[string]string{
				"tracing_incoming_context": "",
				"tracing_outgoing_context": "",
				"tracing_sample_rate":      "2.1",
			},
			tracingProjectId: fakeFlagProjectId,
			wantError:        "Invalid trace sampling rate: 2.1",
		},
		{
			desc: "Success with sample rate 0.0",
			flags: map[string]string{
				"tracing_incoming_context": "",
				"tracing_outgoing_context": "",
				"tracing_sample_rate":      "0.0",
			},
			tracingProjectId: fakeFlagProjectId,
			wantResult: &tracepb.OpenCensusConfig{
				TraceConfig: &opencensuspb.TraceConfig{
					Sampler: &opencensuspb.TraceConfig_ConstantSampler{
						ConstantSampler: &opencensuspb.ConstantSampler{
							Decision: opencensuspb.ConstantSampler_ALWAYS_PARENT,
						},
					},
				},
				StackdriverExporterEnabled: true,
				StackdriverProjectId:       fakeFlagProjectId,
			},
		},
		{
			desc: "Success with sample rate 1.0",
			flags: map[string]string{
				"tracing_incoming_context": "",
				"tracing_outgoing_context": "",
				"tracing_sample_rate":      "1.0",
			},
			tracingProjectId: fakeFlagProjectId,
			wantResult: &tracepb.OpenCensusConfig{
				TraceConfig: &opencensuspb.TraceConfig{
					Sampler: &opencensuspb.TraceConfig_ConstantSampler{
						ConstantSampler: &opencensuspb.ConstantSampler{
							Decision: opencensuspb.ConstantSampler_ALWAYS_ON,
						},
					},
				},
				StackdriverExporterEnabled: true,
				StackdriverProjectId:       fakeFlagProjectId,
			},
		},
		{
			desc: "Success with sample rate 0.5",
			flags: map[string]string{
				"tracing_incoming_context": "",
				"tracing_outgoing_context": "",
				"tracing_sample_rate":      "0.5",
			},
			tracingProjectId: fakeFlagProjectId,
			wantResult: &tracepb.OpenCensusConfig{
				TraceConfig: &opencensuspb.TraceConfig{
					Sampler: &opencensuspb.TraceConfig_ProbabilitySampler{
						ProbabilitySampler: &opencensuspb.ProbabilitySampler{
							SamplingProbability: 0.5,
						},
					},
				},
				StackdriverExporterEnabled: true,
				StackdriverProjectId:       fakeFlagProjectId,
			},
		},
		{
			desc: "Success with custom stackdriver address",
			flags: map[string]string{
				"tracing_stackdriver_address": fakeStackdriverAddress,
				"tracing_incoming_context":    "",
				"tracing_outgoing_context":    "",
				"tracing_sample_rate":         "0.5",
			},
			tracingProjectId: fakeFlagProjectId,
			wantResult: &tracepb.OpenCensusConfig{
				TraceConfig: &opencensuspb.TraceConfig{
					Sampler: &opencensuspb.TraceConfig_ProbabilitySampler{
						ProbabilitySampler: &opencensuspb.ProbabilitySampler{
							SamplingProbability: 0.5,
						},
					},
				},
				StackdriverExporterEnabled: true,
				StackdriverProjectId:       fakeFlagProjectId,
				StackdriverAddress:         fakeStackdriverAddress,
			},
		},
	}

	for _, tc := range testData {

		for fk, fv := range tc.flags {
			flag.Set(fk, fv)
		}

		got, err := CreateTracing(tc.tracingProjectId)

		if tc.wantError != "" && (err == nil || !strings.Contains(err.Error(), tc.wantError)) {
			t.Errorf("Test (%s): failed, expected err: %v, got: %v", tc.desc, tc.wantError, err)
		}

		if tc.wantResult != nil {
			if got == nil {
				t.Errorf("Test (%s): failed, expected result should not be nil", tc.desc)
			}
			if got.Http.Name != "envoy.tracers.opencensus" {
				t.Errorf("Test (%s): failed, expected config name is wrong", tc.desc)
			}

			gotCfg := &tracepb.OpenCensusConfig{}
			if err := ptypes.UnmarshalAny(got.Http.GetTypedConfig(), gotCfg); err != nil {
				t.Errorf("Test (%s): failed, failed to unmarshall any", tc.desc)
			}
			if !reflect.DeepEqual(gotCfg, tc.wantResult) {
				t.Errorf("Test (%s): failed, got : %v, want: %v", tc.desc, gotCfg, tc.wantResult)
			}
		}

	}
}
