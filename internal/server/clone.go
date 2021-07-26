package server

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	v1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	klog "k8s.io/klog/v2"
)

const (
	maxWebhookTimeout = 30

	errValidatingReviewReq = "[error]: failed to validate review request: %v"
	infoRequestReceived    = "[info]: request received for kind=%s, operation=%s, name=%s"
	infoWritingResponse    = "[info]: writing admission review response"
)

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

	klog.Infof(infoRequestReceived, review.Request.Kind.Kind, review.Request.Operation, review.Request.Name)

	var res reviewResponse
	ctx, cancel := context.WithTimeout(context.Background(), maxWebhookTimeout*time.Millisecond)
	defer cancel()

	switch review.Request.Kind.Kind {
	case deployment:
		var deploy appsv1.Deployment
		err = json.Unmarshal(review.Request.Object.Raw, &deploy)
		if err != nil {
			klog.Errorf("[error]: %v", err)
		}

		res, err = s.createResponse(ctx, deploy.Spec.Template.Spec.Containers, review.Request.UID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			klog.Errorf("[error]: %v", err)
		}
	case daemonset:
		var daemonset appsv1.DaemonSet
		err = json.Unmarshal(review.Request.Object.Raw, &daemonset)
		if err != nil {
			klog.Errorf("[error]: %v", err)
		}

		res, err = s.createResponse(ctx, daemonset.Spec.Template.Spec.Containers, review.Request.UID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			klog.Errorf("[error]: %v", err)
		}
	}

	writeAdmissionReviewResponse(w, res)
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

func writeAdmissionReviewResponse(w http.ResponseWriter, r reviewResponse) {
	const ()
	response := v1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       kind,
			APIVersion: version,
		},
		Response: &v1.AdmissionResponse{
			UID:     r.uid,
			Allowed: r.allowed,
			Result: &metav1.Status{
				Code:    r.status.code,
				Message: r.status.message,
				Reason:  r.status.reason,
			},
		},
	}

	if r.patch != nil {
		klog.Infof("[info]: applying patch")
		var patchType v1.PatchType = jsonPatch
		response.Response.Patch = r.patch
		response.Response.PatchType = &patchType
	}

	res, err := json.Marshal(&response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		klog.Errorf("[error]: %v", err)
	}

	klog.Infof(infoWritingResponse)
	_, err = w.Write(res)
	if err != nil {
		klog.Errorf("[error]: %v", err)
	}
}
