package handlers_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	handlers "github.com/be-heroes/ultron/internal/handlers"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestValidatePods_Success(t *testing.T) {
	mockComputeService := &MockComputeService{}
	handler := handlers.NewValidationHandler(mockComputeService)

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

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code 200, got %d", resp.StatusCode)
	}

	var admissionReviewResp admissionv1.AdmissionReview
	if err := json.Unmarshal(body, &admissionReviewResp); err != nil {
		t.Fatalf("Expected valid AdmissionReview response, but got error: %v", err)
	}

	if admissionReviewResp.Response.UID != admissionReviewReq.Request.UID {
		t.Errorf("Expected response UID to match request UID, got %s", admissionReviewResp.Response.UID)
	}

	if admissionReviewResp.Response.Allowed != true {
		t.Errorf("Expected Allowed to be true, but got false")
	}
}

func TestValidatePods_InvalidBody(t *testing.T) {
	mockComputeService := &MockComputeService{}
	handler := handlers.NewValidationHandler(mockComputeService)

	req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewBuffer([]byte("invalid body")))
	w := httptest.NewRecorder()

	handler.ValidatePodSpec(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected status code 400, got %d", resp.StatusCode)
	}
}

func TestValidationHandleAdmissionReview_NonPodKind(t *testing.T) {
	mockComputeService := &MockComputeService{}
	handler := handlers.NewValidationHandler(mockComputeService)

	admissionRequest := &admissionv1.AdmissionRequest{
		Kind: metav1.GroupVersionKind{Kind: "Service"},
	}

	admissionResponse, err := handler.HandleAdmissionReview(admissionRequest)
	if err != nil {
		t.Fatalf("HandleAdmissionReview returned an error: %v", err)
	}

	if admissionResponse.Allowed != true {
		t.Errorf("Expected Allowed to be true for non-pod kind, got false")
	}
}

func TestValidationHandleAdmissionReview_PodSpecFailure(t *testing.T) {
	mockComputeService := &MockComputeService{}
	handler := handlers.NewMutationHandler(mockComputeService)

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
	if err != nil {
		t.Fatalf("HandleAdmissionReview returned an error: %v", err)
	}

	if admissionResponse.Allowed != true {
		t.Errorf("Expected Allowed to be true, but got false")
	}
}
