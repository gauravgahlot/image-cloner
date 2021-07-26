package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	deployment = "Deployment"
	daemonset  = "DaemonSet"
	jsonPatch  = "JSONPatch"
	kind       = "AdmissionReview"
	version    = "admission.k8s.io/v1"
)

const (
	errDockerOperation  = "[error]: failed to %s docker image: %v"
	errCreatingPatch    = "Internal server error creating a patch. Please check the logs."
	errMarshallingPatch = "Internal server error marshalling the patch. Please check the logs."
)

type patch struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

type reviewResponse struct {
	uid     types.UID
	allowed bool
	patch   []byte
	status  status
}

type status struct {
	code    int32
	message string
	reason  metav1.StatusReason
}

func (s *server) createResponse(ctx context.Context, containers []v1.Container, uid types.UID) (reviewResponse, error) {
	patches, err := s.tryCreatePatches(ctx, containers)
	if err != nil {
		return createErrorResponse(uid, 500, metav1.StatusReasonInternalError, errCreatingPatch), err
	}

	var patch []byte
	if len(patches) != 0 {
		patch, err = json.Marshal(patches)
		if err != nil {
			return createErrorResponse(uid, 500, metav1.StatusReasonInternalError, errMarshallingPatch), err
		}
	} else {
		patch = nil
	}

	return reviewResponse{
		uid:     uid,
		allowed: true,
		patch:   patch,
	}, nil
}

func (s *server) tryCreatePatches(ctx context.Context, containers []v1.Container) ([]patch, error) {
	patches := []patch{}
	for i, c := range containers {
		if s.isUsingBackupRegistry(c.Image) {
			continue
		}

		err := s.client.ImagePull(ctx, c.Image)
		if err != nil {
			return nil, fmt.Errorf(errDockerOperation, "pull", err)
		}

		newImage := newImage(c.Image, s.registry, s.registryUser)
		err = s.client.ImageTag(ctx, c.Image, newImage)
		if err != nil {
			return nil, fmt.Errorf(errDockerOperation, "tag", err)
		}

		err = s.client.ImagePush(ctx, newImage)
		if err != nil {
			return nil, fmt.Errorf(errDockerOperation, "push", err)
		}

		patches = append(patches, patch{
			Op:    "replace",
			Path:  fmt.Sprintf("/spec/template/spec/containers/%d/image", i),
			Value: newImage,
		})
	}

	return patches, nil
}

func (s *server) isUsingBackupRegistry(src string) bool {
	if s.registry != "" {
		return strings.HasPrefix(src, s.registry) && strings.Contains(src, s.registryUser)
	}
	return strings.HasPrefix(src, s.registryUser)
}

func createErrorResponse(uid types.UID, code int32, reason metav1.StatusReason, msg string) reviewResponse {
	return reviewResponse{
		uid:     uid,
		allowed: false,
		status: status{
			code:    code,
			reason:  reason,
			message: msg,
		},
	}
}
