package handlers

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"info7255-bigdata-app/elastic"
	"info7255-bigdata-app/models"
	"info7255-bigdata-app/services"
	"log"
	"net/http"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type PlanHandler struct {
	service   services.PlanService
	esFactory *elastic.Factory
}

func NewPlanHandler(service services.PlanService, esFactory *elastic.Factory) *PlanHandler {
	return &PlanHandler{
		service:   service,
		esFactory: esFactory,
	}
}

func (ph *PlanHandler) GetPlan(c *gin.Context) {
	objectId, found := c.Params.Get("objectId")
	if !found {
		c.JSON(http.StatusBadRequest, gin.H{"error": "objectId is required"})
		return
	}

	plan, err := ph.service.GetAnyObject(c, objectId)
	if err != nil {
		log.Printf("Failed to fetch plan with err : %v", err.Error())
		if err.Error() == "KEY_NOT_FOUND" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Plan not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	currentETag := generateETag(plan)
	clientETag := strings.TrimSpace(c.GetHeader("If-None-Match"))
	if clientETag != "" && clientETag == currentETag {
		c.Status(http.StatusNotModified)
		return
	}

	clientETag = strings.TrimSpace(c.GetHeader("If-Match"))
	if clientETag != "" && clientETag != currentETag {
		c.Status(http.StatusPreconditionFailed)
		return
	}

	c.Header("ETag", currentETag)
	c.JSON(http.StatusOK, plan)
}

func (ph *PlanHandler) CreatePlan(c *gin.Context) {
	var planRequest models.Plan

	if err := c.ShouldBindBodyWith(&planRequest, binding.JSON); err != nil {
		log.Printf("Bad request with error : %v", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid fields in the request"})
		return
	}

	existingPlan, err := ph.service.GetPlan(c, planRequest.ObjectId)
	if err == nil && existingPlan.ObjectId != "" {
		c.JSON(http.StatusConflict, gin.H{"error": "Plan already exists"})
		return
	}

	if err := ph.service.CreatePlan(c, planRequest); err != nil {
		log.Printf("Failed to create plan with error : %v", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create plan"})
		return
	}

	eTag := generateETag(planRequest)
	c.Header("ETag", eTag)
	c.JSON(http.StatusCreated, gin.H{"message": "Plan created successfully"})
}

func (ph *PlanHandler) DeletePlan(c *gin.Context) {
	objectId, found := c.Params.Get("objectId")
	if !found {
		c.JSON(http.StatusBadRequest, gin.H{"error": "objectId is required"})
		return
	}

	existingPlan, err := ph.service.GetPlan(c, objectId)
	if err != nil {
		log.Printf("Failed to delete plan with err : %v", err.Error())
		if err.Error() == "KEY_NOT_FOUND" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Plan not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	existingETag := generateETag(existingPlan)
	clientETag := strings.TrimSpace(c.GetHeader("If-Match"))
	if clientETag != "" && clientETag != existingETag {
		c.Status(http.StatusPreconditionFailed)
		return
	}

	if err := ph.service.DeletePlan(c, objectId); err != nil {
		log.Printf("Failed to delete plan with err : %v", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (ph *PlanHandler) UpdatePlan(c *gin.Context) {
	var planRequest models.Plan

	if err := c.ShouldBindBodyWith(&planRequest, binding.JSON); err != nil {
		log.Printf("Bad request with error : %v", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid fields in the request"})
		return
	}

	existingPlan, err := ph.service.GetPlan(c, planRequest.ObjectId)
	if err != nil || existingPlan.ObjectId == "" {
		// If the plan does not exist, create a new one
		if err := ph.service.CreatePlan(c, planRequest); err != nil {
			log.Printf("Failed to create plan with error : %v", err.Error())

			eTag := generateETag(planRequest)
			c.Header("ETag", eTag)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create plan"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "Plan created successfully"})
		return
	}

	currentETag := generateETag(planRequest)
	existingETag := generateETag(existingPlan)

	clientETag := strings.TrimSpace(c.GetHeader("If-None-Match"))
	if clientETag != "" && clientETag == currentETag {
		c.Status(http.StatusNotModified)
		return
	}

	clientETag = strings.TrimSpace(c.GetHeader("If-Match"))
	if clientETag != "" && clientETag != existingETag {
		c.Status(http.StatusPreconditionFailed)
		return
	}

	err = ph.service.UpdatePlan(c, planRequest.ObjectId, planRequest)
	if err != nil {
		log.Printf("Failed to update plan with error : %v", err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	eTag := generateETag(planRequest)
	c.Header("ETag", eTag)
	c.JSON(http.StatusOK, gin.H{"message": "Plan updated successfully"})
}

func (ph *PlanHandler) PatchPlan(c *gin.Context) {
	objectId, found := c.Params.Get("objectId")
	if !found {
		c.JSON(http.StatusBadRequest, gin.H{"error": "objectId is required"})
		return
	}

	var planRequest models.Plan
	if err := c.ShouldBindBodyWith(&planRequest, binding.JSON); err != nil {
		log.Printf("Bad request with error : %v", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid fields in the request"})
		return
	}

	existingPlan, err := ph.service.GetPlan(c, objectId)
	if err != nil || existingPlan.ObjectId == "" {
		log.Printf("Failed to fetch plan with err : %v", err.Error())
		if err.Error() == "KEY_NOT_FOUND" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Plan not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	currentETag := generateETag(planRequest)
	existingETag := generateETag(existingPlan)

	clientETag := strings.TrimSpace(c.GetHeader("If-None-Match"))
	if clientETag != "" && clientETag == currentETag {
		c.Status(http.StatusNotModified)
		return
	}

	clientETag = strings.TrimSpace(c.GetHeader("If-Match"))
	if clientETag != "" && clientETag != existingETag {
		c.Status(http.StatusPreconditionFailed)
		return
	}

	patchedPlan, err := ph.service.PatchPlan(c, objectId, planRequest)
	if err != nil {
		log.Printf("Failed to update plan with error : %v", err.Error())
		if strings.HasPrefix(err.Error(), "ObjectId mismatch") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		return
	}

	eTag := generateETag(patchedPlan)
	c.Header("ETag", eTag)
	c.JSON(http.StatusOK, gin.H{"message": "Plan updated successfully"})
}

func (ph *PlanHandler) GetAllPlans(c *gin.Context) {
	plans, err := ph.service.GetAllPlans(c)
	if err != nil {
		log.Printf("Failed to fetch all plans with err : %v", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, plans)
}

func (ph *PlanHandler) SearchPlans(c *gin.Context) {
	var req models.SearchPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create a match query
	matchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				req.Key: req.Value,
			},
		},
	}
	queryBytes, err := json.Marshal(matchQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Create a new search request
	searchReq := esapi.SearchRequest{
		Index: []string{"plans"},
		Body:  bytes.NewReader(queryBytes),
	}

	// Create a new Elasticsearch client
	cfg := elasticsearch.Config{
		Addresses: []string{
			"http://localhost:9200",
		},
	}
	client, err := ph.esFactory.NewClient(cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Perform the search request.
	res, err := searchReq.Do(context.Background(), client.ES)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer res.Body.Close()

	if res.IsError() {
		c.JSON(http.StatusInternalServerError, gin.H{"error": res.String()})
		return
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func generateETag(plan interface{}) string {
	dataBytes, err := json.Marshal(plan)
	if err != nil {
		log.Printf("Error marshalling plan: %v", err)
		return ""
	}
	return generateSHA1Hash(dataBytes)
}

func generateSHA1Hash(data []byte) string {
	h := sha1.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}
