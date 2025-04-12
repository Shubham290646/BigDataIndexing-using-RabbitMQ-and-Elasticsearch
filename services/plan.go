package services

import (
	"encoding/json"
	"errors"
	"info7255-bigdata-app/models"
	"info7255-bigdata-app/rabbitmq"
	"info7255-bigdata-app/repositories"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

type PlanService interface {
	GetAnyObject(c *gin.Context, key string) (interface{}, error)
	GetPlan(c *gin.Context, key string) (models.Plan, error)
	CreatePlan(c *gin.Context, plan models.Plan) error
	DeletePlan(c *gin.Context, key string) error
	PatchPlan(c *gin.Context, key string, plan models.Plan) (models.Plan, error)
	UpdatePlan(c *gin.Context, key string, plan models.Plan) error
	GetAllPlans(ctx *gin.Context) ([]models.Plan, error)
}

type planService struct {
	repo repositories.RedisRepo
}

func NewPlanService(repo repositories.RedisRepo) PlanService {
	return &planService{
		repo: repo,
	}
}

func (ps *planService) GetPlan(ctx *gin.Context, key string) (models.Plan, error) {
	var plan models.Plan

	value, err := ps.repo.Get(ctx, key)
	if err != nil {
		log.Printf("Error getting the plan from redis : %v", err)
		return plan, err
	}

	if err := json.Unmarshal([]byte(value), &plan); err != nil {
		log.Printf("Error unmarshalling the plan from redis : %v", err)
		return plan, err
	}

	return plan, nil
}

func (ps *planService) CreatePlan(c *gin.Context, plan models.Plan) error {
	objectId := plan.ObjectId

	// Marshal the struct into a string
	value, err := json.Marshal(plan)
	if err != nil {
		log.Errorf("Error marshalling the plan struct : %v", err)
		return err
	}
	err = ps.repo.Set(c, objectId, string(value))
	if err != nil {
		log.Printf("Error setting the plan in the redis : %v", err)
		return err
	}

	pValue, err := json.Marshal(plan.PlanCostShares)
	if err != nil {
		log.Errorf("Error marshalling the plan struct : %v", err)
		return err
	}
	err = ps.repo.Set(c, plan.PlanCostShares.ObjectId, string(pValue))
	if err != nil {
		log.Printf("Error setting the plan in the redis : %v", err)
		return err
	}

	// Iterate over the linkedPlanServices array and set each object in Redis
	for _, linkedPlanService := range plan.LinkedPlanServices {
		// Marshal and store the entire linkedPlanService object in Redis
		linkedPlanServiceValue, err := json.Marshal(linkedPlanService)
		if err != nil {
			log.Errorf("Error marshalling the linkedPlanService struct : %v", err)
			return err
		}

		err = ps.repo.Set(c, linkedPlanService.ObjectId, string(linkedPlanServiceValue))
		if err != nil {
			return err
		}

		// Marshal and store the linkedService object in Redis
		linkedServiceValue, err := json.Marshal(linkedPlanService.LinkedService)
		if err != nil {
			log.Errorf("Error marshalling the linkedService struct : %v", err)
			return err
		}

		err = ps.repo.Set(c, linkedPlanService.LinkedService.ObjectId, string(linkedServiceValue))
		if err != nil {
			return err
		}

		// Marshal and store the planserviceCostShares object in Redis
		planserviceCostSharesValue, err := json.Marshal(linkedPlanService.PlanServiceCostShares)
		if err != nil {
			log.Errorf("Error marshalling the planserviceCostShares struct : %v", err)
			return err
		}

		err = ps.repo.Set(c, linkedPlanService.PlanServiceCostShares.ObjectId, string(planserviceCostSharesValue))
		if err != nil {
			return err
		}
	}

	// Publish the plan creation message to RabbitMQ
	message := models.PlanMessage{
		Operation: "create",
		Plan:      plan,
	}

	rmq := &rabbitmq.Factory{}
	err = rmq.PublishMessage("plans_queue", message)
	if err != nil {
		log.Errorf("Error publishing create message to RabbitMQ: %v", err)
		return err
	}

	return nil
}

func (ps *planService) DeletePlan(c *gin.Context, objectId string) error {
	// Fetch the plan
	plan, err := ps.GetPlan(c, objectId)
	if err != nil {
		log.Printf("Error getting the plan from the redis : %v", err)
		return err
	}

	// Delete the plan
	err = ps.repo.Delete(c, objectId)
	if err != nil {
		log.Printf("Error deleting the plan from the redis : %v", err)
		return err
	}

	// Delete the PlanCostShares
	err = ps.repo.Delete(c, plan.PlanCostShares.ObjectId)
	if err != nil {
		log.Printf("Error deleting the PlanCostShares from the redis : %v", err)
		return err
	}

	// Delete each LinkedPlanService and its related objects
	for _, linkedPlanService := range plan.LinkedPlanServices {
		// Delete the LinkedPlanService
		err = ps.repo.Delete(c, linkedPlanService.ObjectId)
		if err != nil {
			log.Printf("Error deleting the LinkedPlanService from the redis : %v", err)
			return err // This line was uncommented
		}

		// Delete the LinkedService
		err = ps.repo.Delete(c, linkedPlanService.LinkedService.ObjectId)
		if err != nil {
			log.Printf("Error deleting the LinkedService from the redis : %v", err)
			return err
		}

		// Delete the PlanServiceCostShares
		err = ps.repo.Delete(c, linkedPlanService.PlanServiceCostShares.ObjectId)
		if err != nil {
			log.Printf("Error deleting the PlanServiceCostShares from the redis : %v", err)
			return err
		}
	}

	// Publish the plan deletion message to RabbitMQ
	message := models.PlanMessage{
		Operation: "delete",
		Plan:      plan,
	}

	rmq := &rabbitmq.Factory{}
	err = rmq.PublishMessage("plans_queue", message)
	if err != nil {
		log.Errorf("Error publishing delete message to RabbitMQ: %v", err)
		return err
	}

	return nil
}

func (ps *planService) PatchPlan(ctx *gin.Context, key string, plan models.Plan) (models.Plan, error) {
	existingPlan, err := ps.GetPlan(ctx, key)
	if err != nil || existingPlan.ObjectId == "" {
		return models.Plan{}, err
	}

	if plan.PlanCostShares != nil {
		var pValue []byte
		if existingPlan.PlanCostShares != nil {
			if existingPlan.PlanCostShares.ObjectId != plan.PlanCostShares.ObjectId {
				validationErr := errors.New("ObjectId mismatch in planCostShares")
				log.Errorf("Error updating planCostShares : %v", validationErr)
				return models.Plan{}, validationErr
			}
			existingPlan.PlanCostShares.UpdatePlanCostShares(*plan.PlanCostShares)
			pValue, err = json.Marshal(existingPlan.PlanCostShares)
			if err != nil {
				log.Errorf("Error marshalling the plan struct : %v", err)
				return models.Plan{}, err
			}
			err = ps.repo.Set(ctx, plan.PlanCostShares.ObjectId, string(pValue))
			if err != nil {
				log.Printf("Error setting the plan in the redis : %v", err)
				return models.Plan{}, err
			}
		} else {
			pValue, err = json.Marshal(plan.PlanCostShares)
			if err != nil {
				log.Errorf("Error marshalling the plan struct : %v", err)
				return models.Plan{}, err
			}
		}

		err = ps.repo.Set(ctx, plan.PlanCostShares.ObjectId, string(pValue))
		if err != nil {
			log.Printf("Error setting the plan in the redis : %v", err)
			return models.Plan{}, err
		}
	}

	// Create a map of new LinkedPlanServices for easy lookup
	newLinkedPlanServices := make(map[string]models.LinkedPlanService)
	for _, newLinkedPlanService := range plan.LinkedPlanServices {
		newLinkedPlanServices[newLinkedPlanService.ObjectId] = newLinkedPlanService
	}

	// Update existing LinkedPlanServices if they are in the newLinkedPlanServices map
	for i, existingLinkedPlanService := range existingPlan.LinkedPlanServices {
		if newLinkedPlanService, ok := newLinkedPlanServices[existingLinkedPlanService.ObjectId]; ok {
			existingPlan.LinkedPlanServices[i] = newLinkedPlanService
			delete(newLinkedPlanServices, existingLinkedPlanService.ObjectId)
		}
	}

	// Append any remaining new LinkedPlanServices that were not in the existing plan
	for _, newLinkedPlanService := range newLinkedPlanServices {
		existingPlan.LinkedPlanServices = append(existingPlan.LinkedPlanServices, newLinkedPlanService)

		linkedPlanServiceValue, err := json.Marshal(newLinkedPlanService)
		if err != nil {
			log.Errorf("Error marshalling the linkedPlanService struct : %v", err)
			return models.Plan{}, err
		}

		err = ps.repo.Set(ctx, newLinkedPlanService.ObjectId, string(linkedPlanServiceValue))
		if err != nil {
			return models.Plan{}, err
		}

		// Marshal and store the linkedService object in Redis
		linkedServiceValue, err := json.Marshal(newLinkedPlanService.LinkedService)
		if err != nil {
			log.Errorf("Error marshalling the linkedService struct : %v", err)
			return models.Plan{}, err
		}

		err = ps.repo.Set(ctx, newLinkedPlanService.LinkedService.ObjectId, string(linkedServiceValue))
		if err != nil {
			return models.Plan{}, err
		}

		// Marshal and store the planserviceCostShares object in Redis
		planserviceCostSharesValue, err := json.Marshal(newLinkedPlanService.PlanServiceCostShares)
		if err != nil {
			log.Errorf("Error marshalling the planserviceCostShares struct : %v", err)
			return models.Plan{}, err
		}

		err = ps.repo.Set(ctx, newLinkedPlanService.PlanServiceCostShares.ObjectId, string(planserviceCostSharesValue))
		if err != nil {
			return models.Plan{}, err
		}
	}

	existingPlan.Org = plan.Org
	// existingPlan.PlanStatus = plan.PlanStatus
	existingPlan.CreationDate = plan.CreationDate

	if plan.ObjectId != "" && existingPlan.ObjectId != plan.ObjectId {
		validationErr := errors.New("ObjectId mismatch in plan")
		log.Errorf("Error updating plan : %v", validationErr)
		return models.Plan{}, validationErr
	}

	value, err := json.Marshal(existingPlan)
	if err != nil {
		log.Printf("Error marshalling plan: %v", err)
		return models.Plan{}, err
	}

	err = ps.repo.Set(ctx, key, string(value))
	if err != nil {
		log.Printf("Error saving plan to redis: %v", err)
		return models.Plan{}, err
	}

	// FIXED: Use existingPlan instead of input plan for RabbitMQ message
	message := models.PlanMessage{
		Operation: "patch",
		Plan:      existingPlan, // Changed from plan to existingPlan
	}

	rmq := &rabbitmq.Factory{}
	err = rmq.PublishMessage("plans_queue", message)
	if err != nil {
		log.Errorf("Error publishing patch message to RabbitMQ: %v", err)
		return models.Plan{}, err
	}

	return existingPlan, nil
}

func (ps *planService) UpdatePlan(ctx *gin.Context, key string, plan models.Plan) error {
	// Delete the existing plan and all its associated objects
	err := ps.DeletePlan(ctx, key)
	if err != nil {
		log.Printf("Failed to delete existing plan with error : %v", err.Error())
		return err
	}

	// Create a new plan with the new request body
	err = ps.CreatePlan(ctx, plan)
	if err != nil {
		log.Printf("Failed to create new plan with error : %v", err.Error())
		return err
	}

	return nil
}

func (ps *planService) GetAllPlans(ctx *gin.Context) ([]models.Plan, error) {
	plans := make([]models.Plan, 0)
	keys, err := ps.repo.Keys(ctx, "*")
	if err != nil {
		log.Printf("Error fetching all the keys from the redis : %v", err)
		return nil, err
	}
	for _, key := range keys {
		value, err := ps.repo.Get(ctx, key)
		if err != nil {
			log.Printf("Error fetching the value from the redis : %v", err)
			return nil, err
		}

		var plan models.Plan
		err = json.Unmarshal([]byte(value), &plan)
		if err != nil {
			log.Printf("Error unmarshalling the plan from the redis : %v", err)
			continue
		}

		if plan.ObjectId != "" && plan.ObjectType != "" && plan.PlanCostShares.ObjectId != "" {
			plans = append(plans, plan)
		}
	}

	return plans, nil
}

func (ps *planService) GetAnyObject(ctx *gin.Context, key string) (interface{}, error) {
	value, err := ps.repo.Get(ctx, key)
	if err != nil {
		log.Printf("Error getting the plan from the redis : %v", err)
		return nil, err
	}

	var plan models.Plan
	err = json.Unmarshal([]byte(value), &plan)
	if err != nil {
		log.Printf("Error unmarshalling the plan from the redis : %v", err)
		return nil, err
	}

	// Check the ObjectType and return the corresponding struct
	switch plan.ObjectType {
	case "membercostshare":
		var pcs models.PlanCostShares
		err = json.Unmarshal([]byte(value), &pcs)
		if err != nil {
			return nil, err
		}
		return pcs, nil
	case "service":
		var ls models.LinkedService
		err = json.Unmarshal([]byte(value), &ls)
		if err != nil {
			return nil, err
		}
		return ls, nil
	case "PlanServiceCostShares":
		var pscs models.PlanServiceCostShares
		err = json.Unmarshal([]byte(value), &pscs)
		if err != nil {
			return nil, err
		}
		return pscs, nil
	case "planservice":
		var lps models.LinkedPlanService
		err = json.Unmarshal([]byte(value), &lps)
		if err != nil {
			return nil, err
		}
		return lps, nil
	default:
		return plan, nil
	}
}
