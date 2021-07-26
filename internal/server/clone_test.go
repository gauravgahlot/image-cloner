// Copyright 2021 The image-cloner Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCloneImage(t *testing.T) {
	d := mockDockerClient{
		ImagePullFunc: func(ctx context.Context, image string) error { return nil },
		ImageTagFunc:  func(ctx context.Context, src, dst string) error { return nil },
		ImagePushFunc: func(ctx context.Context, image string) error { return nil },
	}
	mods := []serverModifier{withRegistryUser(registryUser)}
	s := testServer(t, d, mods...)

	cases := []struct {
		name   string
		source string
	}{
		{
			name:   "admission-review-request-deployment",
			source: admissionReviewRequestDeployment,
		},
		{
			name:   "admission-review-request-daemonset",
			source: admissionReviewRequestDaemonSet,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/clone-image", bytes.NewBuffer([]byte(tc.source)))
			if err != nil {
				t.Error(err)
			}

			rr := httptest.NewRecorder()
			http.HandlerFunc(s.cloneImage).ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusOK {
				t.Errorf("Status code differs. Expected %d .\n Got %d instead.", http.StatusOK, status)
			}
			assert.JSONEq(t, expectedResponse(), rr.Body.String(), "response body differs")
		})
	}
}

func expectedResponse() string {
	var patchType v1.PatchType = jsonPatch
	response := v1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       kind,
			APIVersion: version,
		},
		Response: &v1.AdmissionResponse{
			UID:       uid,
			Allowed:   true,
			Patch:     getPatch(alpine, "", registryUser),
			PatchType: &patchType,
			Result:    &metav1.Status{},
		},
	}
	res, _ := json.Marshal(response)
	return string(res)
}

const (
	admissionReviewRequestDeployment = `{
	"kind": "AdmissionReview",
	"apiVersion": "admission.k8s.io/v1",
	"request": {
	  "uid": "4584308f-b307-455b-ab11-5765b4548b71",
	  "kind": { "group": "apps", "version": "v1", "kind": "Deployment" },
	  "resource": { "group": "apps", "version": "v1", "resource": "deployments" },
	  "requestKind": { "group": "apps", "version": "v1", "kind": "Deployment" },
	  "requestResource": {
		"group": "apps",
		"version": "v1",
		"resource": "deployments"
	  },
	  "name": "alpine",
	  "namespace": "default",
	  "operation": "CREATE",
	  "userInfo": {
		"username": "minikube-user",
		"groups": ["system:masters", "system:authenticated"]
	  },
	  "object": {
		"kind": "Deployment",
		"apiVersion": "apps/v1",
		"metadata": {
		  "name": "alpine",
		  "namespace": "default",
		  "creationTimestamp": null,
		  "labels": { "app": "alpine" },
		  "managedFields": [
			{
			  "manager": "kubectl-create",
			  "operation": "Update",
			  "apiVersion": "apps/v1",
			  "time": "2021-07-25T12:58:19Z",
			  "fieldsType": "FieldsV1",
			  "fieldsV1": {
				"f:metadata": { "f:labels": { ".": {}, "f:app": {} } },
				"f:spec": {
				  "f:progressDeadlineSeconds": {},
				  "f:replicas": {},
				  "f:revisionHistoryLimit": {},
				  "f:selector": {},
				  "f:strategy": {
					"f:rollingUpdate": {
					  ".": {},
					  "f:maxSurge": {},
					  "f:maxUnavailable": {}
					},
					"f:type": {}
				  },
				  "f:template": {
					"f:metadata": { "f:labels": { ".": {}, "f:app": {} } },
					"f:spec": {
					  "f:containers": {
						"k:{\"name\":\"alpine\"}": {
						  ".": {},
						  "f:image": {},
						  "f:imagePullPolicy": {},
						  "f:name": {},
						  "f:ports": {
							".": {},
							"k:{\"containerPort\":80,\"protocol\":\"TCP\"}": {
							  ".": {},
							  "f:containerPort": {},
							  "f:protocol": {}
							}
						  },
						  "f:resources": {},
						  "f:terminationMessagePath": {},
						  "f:terminationMessagePolicy": {}
						}
					  },
					  "f:dnsPolicy": {},
					  "f:restartPolicy": {},
					  "f:schedulerName": {},
					  "f:securityContext": {},
					  "f:terminationGracePeriodSeconds": {}
					}
				  }
				}
			  }
			}
		  ]
		},
		"spec": {
		  "replicas": 1,
		  "selector": { "matchLabels": { "app": "alpine" } },
		  "template": {
			"metadata": {
			  "creationTimestamp": null,
			  "labels": { "app": "alpine" }
			},
			"spec": {
			  "containers": [
				{
				  "name": "alpine",
				  "image": "alpine:3.12",
				  "ports": [{ "containerPort": 80, "protocol": "TCP" }],
				  "resources": {},
				  "terminationMessagePath": "/dev/termination-log",
				  "terminationMessagePolicy": "File",
				  "imagePullPolicy": "Always"
				}
			  ],
			  "restartPolicy": "Always",
			  "terminationGracePeriodSeconds": 30,
			  "dnsPolicy": "ClusterFirst",
			  "securityContext": {},
			  "schedulerName": "default-scheduler"
			}
		  },
		  "strategy": {
			"type": "RollingUpdate",
			"rollingUpdate": { "maxUnavailable": "25%", "maxSurge": "25%" }
		  },
		  "revisionHistoryLimit": 10,
		  "progressDeadlineSeconds": 600
		},
		"status": {}
	  },
	  "oldObject": null,
	  "dryRun": false,
	  "options": {
		"kind": "CreateOptions",
		"apiVersion": "meta.k8s.io/v1",
		"fieldManager": "kubectl-create"
	  }
	}
  }
`

	admissionReviewRequestDaemonSet = `{
	"kind": "AdmissionReview",
	"apiVersion": "admission.k8s.io/v1",
	"request": {
	  "uid": "4584308f-b307-455b-ab11-5765b4548b71",
	  "kind": { "group": "apps", "version": "v1", "kind": "DaemonSet" },
	  "resource": { "group": "apps", "version": "v1", "resource": "daemonsets" },
	  "requestKind": { "group": "apps", "version": "v1", "kind": "DaemonSet" },
	  "requestResource": {
		"group": "apps",
		"version": "v1",
		"resource": "daemonsets"
	  },
	  "name": "alpine",
	  "namespace": "default",
	  "operation": "CREATE",
	  "userInfo": {
		"username": "minikube-user",
		"groups": ["system:masters", "system:authenticated"]
	  },
	  "object": {
		"kind": "DaemonSet",
		"apiVersion": "apps/v1",
		"metadata": {
		  "name": "alpine",
		  "namespace": "default",
		  "creationTimestamp": null,
		  "labels": { "app": "alpine" },
		  "managedFields": [
			{
			  "manager": "kubectl-create",
			  "operation": "Update",
			  "apiVersion": "apps/v1",
			  "time": "2021-07-25T12:58:19Z",
			  "fieldsType": "FieldsV1",
			  "fieldsV1": {
				"f:metadata": { "f:labels": { ".": {}, "f:app": {} } },
				"f:spec": {
				  "f:progressDeadlineSeconds": {},
				  "f:revisionHistoryLimit": {},
				  "f:selector": {},
				  "f:strategy": {
					"f:rollingUpdate": {
					  ".": {},
					  "f:maxSurge": {},
					  "f:maxUnavailable": {}
					},
					"f:type": {}
				  },
				  "f:template": {
					"f:metadata": { "f:labels": { ".": {}, "f:app": {} } },
					"f:spec": {
					  "f:containers": {
						"k:{\"name\":\"alpine\"}": {
						  ".": {},
						  "f:image": {},
						  "f:imagePullPolicy": {},
						  "f:name": {},
						  "f:ports": {
							".": {},
							"k:{\"containerPort\":80,\"protocol\":\"TCP\"}": {
							  ".": {},
							  "f:containerPort": {},
							  "f:protocol": {}
							}
						  },
						  "f:resources": {},
						  "f:terminationMessagePath": {},
						  "f:terminationMessagePolicy": {}
						}
					  },
					  "f:dnsPolicy": {},
					  "f:restartPolicy": {},
					  "f:schedulerName": {},
					  "f:securityContext": {},
					  "f:terminationGracePeriodSeconds": {}
					}
				  }
				}
			  }
			}
		  ]
		},
		"spec": {
		  "selector": { "matchLabels": { "app": "alpine" } },
		  "template": {
			"metadata": {
			  "creationTimestamp": null,
			  "labels": { "app": "alpine" }
			},
			"spec": {
			  "containers": [
				{
				  "name": "alpine",
				  "image": "alpine:3.12",
				  "ports": [{ "containerPort": 80, "protocol": "TCP" }],
				  "resources": {},
				  "terminationMessagePath": "/dev/termination-log",
				  "terminationMessagePolicy": "File",
				  "imagePullPolicy": "Always"
				}
			  ],
			  "restartPolicy": "Always",
			  "terminationGracePeriodSeconds": 30,
			  "dnsPolicy": "ClusterFirst",
			  "securityContext": {},
			  "schedulerName": "default-scheduler"
			}
		  },
		  "strategy": {
			"type": "RollingUpdate",
			"rollingUpdate": { "maxUnavailable": "25%", "maxSurge": "25%" }
		  },
		  "revisionHistoryLimit": 10,
		  "progressDeadlineSeconds": 600
		},
		"status": {}
	  },
	  "oldObject": null,
	  "dryRun": false,
	  "options": {
		"kind": "CreateOptions",
		"apiVersion": "meta.k8s.io/v1",
		"fieldManager": "kubectl-create"
	  }
	}
  }
`
)
