package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	services "ultron/internal/services"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
)

type MutationHandler interface {
	MutatePods(w http.ResponseWriter, r *http.Request)
	HandleAdmissionReview(request *admissionv1.AdmissionRequest) (*admissionv1.AdmissionResponse, error)
}

type IMutationHandler struct {
	computeService services.ComputeService
}

func NewIMutationHandler(computeService services.ComputeService) *IMutationHandler {
	return &IMutationHandler{
		computeService: computeService,
	}
}

func (mh IMutationHandler) MutatePods(w http.ResponseWriter, r *http.Request) {
	var admissionReviewReq admissionv1.AdmissionReview
	var admissionReviewResp admissionv1.AdmissionReview

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Could not read request body: %v", err)
		http.Error(w, "could not read request body", http.StatusBadRequest)

		return
	}

	if err := json.Unmarshal(body, &admissionReviewReq); err != nil {
		log.Printf("Could not unmarshal request: %v", err)
		http.Error(w, "could not unmarshal request", http.StatusBadRequest)

		return
	}

	admissionResponse, err := mh.HandleAdmissionReview(admissionReviewReq.Request)
	if err != nil {
		log.Printf("Could not handle addmission review: %v", err)
		http.Error(w, "could not handle addmission review", http.StatusInternalServerError)

		return
	}

	admissionReviewResp.Response = admissionResponse
	admissionReviewResp.Kind = admissionReviewReq.Kind
	admissionReviewResp.APIVersion = admissionReviewReq.APIVersion
	admissionReviewResp.Response.UID = admissionReviewReq.Request.UID

	respBytes, err := json.Marshal(admissionReviewResp)
	if err != nil {
		log.Printf("Could not marshal response: %v", err)
		http.Error(w, "could not marshal response", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(respBytes); err != nil {
		log.Printf("Could not write response: %v", err)
		http.Error(w, "could not write response", http.StatusInternalServerError)
	}
}

func (mh IMutationHandler) HandleAdmissionReview(request *admissionv1.AdmissionRequest) (*admissionv1.AdmissionResponse, error) {
	if request.Kind.Kind != "Pod" {
		return &admissionv1.AdmissionResponse{
			Allowed: true,
		}, nil
	}

	var pod corev1.Pod
	if err := json.Unmarshal(request.Object.Raw, &pod); err != nil {
		return &admissionv1.AdmissionResponse{
			Allowed: true,
		}, err
	}

	wNode, err := mh.computeService.ComputePodSpec(&pod)
	if err != nil {
		return nil, err
	}

	pod.Spec.NodeSelector = wNode.Selector
	patchBytes, err := json.Marshal([]map[string]interface{}{
		{
			"op":    "add",
			"path":  "/spec/nodeSelector",
			"value": pod.Spec.NodeSelector,
		},
	})
	if err != nil {
		return &admissionv1.AdmissionResponse{
			Allowed: true,
		}, err
	}

	return &admissionv1.AdmissionResponse{
		Allowed:   true,
		Patch:     patchBytes,
		PatchType: func() *admissionv1.PatchType { pt := admissionv1.PatchTypeJSONPatch; return &pt }(),
	}, nil
}
