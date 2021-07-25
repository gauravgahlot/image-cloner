package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	v1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	klog "k8s.io/klog/v2"
)

const (
	errValidatingReviewReq = "[error]: failed to validate review request: %v"
	errCreateDockerClient  = "[error]: error creating docker client: %v"
	errDockerOperation     = "[error]: failed to %s docker image: %v"
	infoRequestReceived    = "[info]: review request received for kind: %s\t operation: %s\t name: %s"
)

type patch struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func (s *server) cloneImage(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		klog.Errorf("[error]: %v", err)
	}

	review, err := validateReviewRequest(body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		klog.Errorf(errValidatingReviewReq, err)
	}

	klog.Infof(infoRequestReceived, review.Request.Kind, review.Request.Operation, review.Request.Name)

	var deploy appsv1.Deployment
	err = json.Unmarshal(review.Request.Object.Raw, &deploy)
	if err != nil {
		klog.Errorf("[error]: %v", err)
	}

	patchList := make([]patch, len(deploy.Spec.Template.Spec.Containers))
	for i, c := range deploy.Spec.Template.Spec.Containers {
		err = s.client.ImagePull(context.Background(), c.Image)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			klog.Errorf(errDockerOperation, "pull", err)
		}

		newImage := newImage(c.Image, s.registry, s.registryUser)
		err = s.client.ImageTag(context.Background(), c.Image, newImage)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			klog.Errorf(errDockerOperation, "tag", err)
		}

		err = s.client.ImagePush(context.Background(), newImage)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			klog.Errorf(errDockerOperation, "push", err)
		}

		patchList[i] = patch{
			Op:    "replace",
			Path:  fmt.Sprintf("/spec/template/spec/containers/%d/image", i),
			Value: newImage,
		}
	}

	patches, err := json.Marshal(patchList)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		klog.Errorf("[error]: %v", err)
	}

	writeAdmissionReviewResponse(w, review.Request.UID, true, patches)
}

func validateReviewRequest(body []byte) (v1.AdmissionReview, error) {
	deserializer := serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer()
	var reviewReq v1.AdmissionReview

	if _, _, err := deserializer.Decode(body, nil, &reviewReq); err != nil {
		return reviewReq, err
	} else if reviewReq.Request == nil {
		return reviewReq, err
	}

	return reviewReq, nil
}

func newImage(src, registry, user string) string {
	img := strings.Split(src, "/")
	if registry == "" {
		return strings.Join([]string{user, img[len(img)-1]}, "/")
	}
	return strings.Join([]string{registry, user, img[len(img)-1]}, "/")
}

func writeAdmissionReviewResponse(w http.ResponseWriter, uid types.UID, allowed bool, patch []byte) {
	const (
		kind    = "AdmissionReview"
		version = "admission.k8s.io/v1"
	)
	response := v1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       kind,
			APIVersion: version,
		},
		Response: &v1.AdmissionResponse{
			UID:     uid,
			Allowed: allowed,
		},
	}

	if patch != nil {
		klog.Infof("[info]: applying patch")
		var patchType v1.PatchType = "JSONPatch"
		response.Response.Patch = patch
		response.Response.PatchType = &patchType
	}

	res, err := json.Marshal(&response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		klog.Errorf("[error]: %v", err)
	}

	klog.Infof("[info]: writing admission review response")
	w.Write(res)
}
