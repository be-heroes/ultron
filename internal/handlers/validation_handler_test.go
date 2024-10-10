package handlers_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	handlers "github.com/be-heroes/ultron/internal/handlers"
	"github.com/be-heroes/ultron/mocks" // Import the generated mocks
	ultron "github.com/be-heroes/ultron/pkg"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestValidatePods_Success(t *testing.T) {
	mockComputeService := new(mocks.IComputeService)

	mockComputeService.On("MatchPodSpec", mock.AnythingOfType("*v1.Pod")).
		Return(&ultron.WeightedNode{
			Selector: map[string]string{"node-type": "mock-node"},
		}, nil)

	handler := handlers.NewValidationHandler(mockComputeService, nil)

	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod",
		},
		Spec: corev1.PodSpec{},
	}
	rawPod, _ := json.Marshal(pod)

	admissionReviewReq := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:  "1234",
			Kind: metav1.GroupVersionKind{Kind: "Pod"},
			Object: runtime.RawExtension{
				Raw: rawPod,
			},
		},
	}

	reqBody, _ := json.Marshal(admissionReviewReq)
	req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	handler.ValidatePodSpec(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status code 200")

	var admissionReviewResp admissionv1.AdmissionReview
	err := json.Unmarshal(body, &admissionReviewResp)
	assert.NoError(t, err, "Expected valid AdmissionReview response")
	assert.Equal(t, admissionReviewReq.Request.UID, admissionReviewResp.Response.UID, "Expected response UID to match request UID")
	assert.True(t, admissionReviewResp.Response.Allowed, "Expected Allowed to be true")
}

func TestValidatePods_InvalidBody(t *testing.T) {
	mockComputeService := new(mocks.IComputeService)

	handler := handlers.NewValidationHandler(mockComputeService, nil)

	req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewBuffer([]byte("invalid body")))
	w := httptest.NewRecorder()

	handler.ValidatePodSpec(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Expected status code 400")
}

func TestValidationHandleAdmissionReview_NonPodKind(t *testing.T) {
	mockComputeService := new(mocks.IComputeService)

	handler := handlers.NewValidationHandler(mockComputeService, nil)

	admissionRequest := &admissionv1.AdmissionRequest{
		Kind: metav1.GroupVersionKind{Kind: "Service"},
	}

	admissionResponse, err := handler.HandleAdmissionReview(admissionRequest)
	assert.NoError(t, err, "HandleAdmissionReview should not return an error")
	assert.True(t, admissionResponse.Allowed, "Expected Allowed to be true for non-pod kind")
}

func TestValidationHandleAdmissionReview_PodSpecFailure(t *testing.T) {
	mockComputeService := new(mocks.IComputeService)

	mockComputeService.On("MatchPodSpec", mock.AnythingOfType("*v1.Pod")).Return(nil, nil)

	handler := handlers.NewValidationHandler(mockComputeService, nil)

	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod",
		},
		Spec: corev1.PodSpec{},
	}
	rawPod, _ := json.Marshal(pod)

	admissionRequest := &admissionv1.AdmissionRequest{
		UID:  "1234",
		Kind: metav1.GroupVersionKind{Kind: "Pod"},
		Object: runtime.RawExtension{
			Raw: rawPod,
		},
	}

	admissionResponse, err := handler.HandleAdmissionReview(admissionRequest)
	assert.NoError(t, err, "HandleAdmissionReview should not return an error")
	assert.True(t, admissionResponse.Allowed, "Expected Allowed to be true")
}
