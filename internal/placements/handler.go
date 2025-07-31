package placements

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/varnit-ta/PlacementLog/pkg/utils"
)

type PlacementsHandler struct {
	srv *PlacementsService
}

func NewPlacementsHandler(srv *PlacementsService) *PlacementsHandler {
	return &PlacementsHandler{srv: srv}
}

// CTCValue handles both numeric CTC and "NA" string
type CTCValue struct {
	Value *float64
}

// UnmarshalJSON implements custom JSON unmarshaling for CTC
func (c *CTCValue) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first (for "NA")
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		if str == "NA" || str == "na" {
			c.Value = nil
			return nil
		}
		return fmt.Errorf("invalid CTC string value: %s, only 'NA' is allowed", str)
	}

	// Try to unmarshal as float64
	var val float64
	if err := json.Unmarshal(data, &val); err == nil {
		c.Value = &val
		return nil
	}

	return fmt.Errorf("CTC must be a number or 'NA'")
}

// MarshalJSON implements custom JSON marshaling for CTC
func (c CTCValue) MarshalJSON() ([]byte, error) {
	if c.Value == nil {
		return json.Marshal("NA")
	}
	return json.Marshal(*c.Value)
}

type PlacementRequest struct {
	Company       string   `json:"company"`
	CTC           CTCValue `json:"ctc"`
	PlacementDate string   `json:"placement_date"`
	Students      []string `json:"students"`
}

type PlacementResponse struct {
	PlacementID   int           `json:"placement_id"`
	Company       string        `json:"company"`
	CTC           CTCValue      `json:"ctc"`
	PlacementDate string        `json:"placement_date"`
	BranchCounts  []BranchCount `json:"branch_counts"`
}

// POST /placements (admin only, enforced by router middleware)
func (h *PlacementsHandler) AddPlacement(w http.ResponseWriter, r *http.Request) {
	var req PlacementRequest
	if err := utils.ReadJSON(r, &req); err != nil {
		utils.WriteError(w, err)
		return
	}
	resp, err := h.srv.AddPlacement(req)
	if err != nil {
		utils.WriteError(w, err)
		return
	}
	utils.WriteJSON(w, resp, http.StatusCreated)
}

// GET /placements (all users)
func (h *PlacementsHandler) GetAllPlacements(w http.ResponseWriter, r *http.Request) {
	placementsList, err := h.srv.GetAllPlacements()
	if err != nil {
		utils.WriteError(w, err)
		return
	}
	utils.WriteJSON(w, placementsList, http.StatusOK)
}

// GET /placements/company-branch (public)
func (h *PlacementsHandler) GetCompanyBranchMap(w http.ResponseWriter, r *http.Request) {
	result, err := h.srv.GetCompanyBranchMap()
	if err != nil {
		utils.WriteError(w, err)
		return
	}
	utils.WriteJSON(w, result, http.StatusOK)
}

// GET /placements/branch-company (public)
func (h *PlacementsHandler) GetBranchCompanyMap(w http.ResponseWriter, r *http.Request) {
	result, err := h.srv.GetBranchCompanyMap()
	if err != nil {
		utils.WriteError(w, err)
		return
	}
	utils.WriteJSON(w, result, http.StatusOK)
}
