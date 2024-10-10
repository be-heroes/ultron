package handlers

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	ultron "github.com/be-heroes/ultron/pkg"
	services "github.com/be-heroes/ultron/pkg/services"
	"github.com/redis/go-redis/v9"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
)

type IValidationHandler interface {
	ValidatePodSpec(w http.ResponseWriter, r *http.Request)
	HandleAdmissionReview(request *admissionv1.AdmissionRequest) (*admissionv1.AdmissionResponse, error)
}

type ValidationHandler struct {
	computeService *services.IComputeService
	redisClient    *redis.Client
}

func NewValidationHandler(computeService *services.IComputeService, redisClient *redis.Client) *ValidationHandler {
	return &ValidationHandler{
		computeService: computeService,
		redisClient:    redisClient,
	}
}

func (vh *ValidationHandler) ValidatePodSpec(w http.ResponseWriter, r *http.Request) {
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

	admissionResponse, err := vh.HandleAdmissionReview(admissionReviewReq.Request)
	if err != nil {
		log.Printf("Could not handle admission review: %v", err)
		http.Error(w, "could not handle admission review", http.StatusInternalServerError)

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

func (vh *ValidationHandler) HandleAdmissionReview(request *admissionv1.AdmissionRequest) (*admissionv1.AdmissionResponse, error) {
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

	wNode, err := (*vh.computeService).MatchPodSpec(&pod)
	if err != nil {
		return nil, err
	}

	if wNode == nil && vh.redisClient != nil {
		vh.redisClient.Publish(context.Background(), ultron.TopicPodObserve, pod)
	} else if vh.redisClient != nil {
		vh.redisClient.Publish(context.Background(), ultron.TopicNodeObserve, wNode)
	}

	return &admissionv1.AdmissionResponse{
		Allowed: true,
	}, nil
}
