package placements

import (
	"strings"
	"time"
)

type BranchCount struct {
	Branch string `json:"branch"`
	Count  int    `json:"count"`
}

type PlacementCompany struct {
	ID            int           `json:"id"`
	Company       string        `json:"company"`
	CTC           CTCValue      `json:"ctc"`
	PlacementDate string        `json:"placement_date"`
	CreatedAt     string        `json:"created_at"`
	BranchCounts  []BranchCount `json:"branch_counts,omitempty"`
}

func GetBranchFromRegNo(regNo string) string {
	if len(regNo) >= 5 {
		return strings.ToLower(regNo[2:5])
	}
	return ""
}

func CountBranches(regNos []string) []BranchCount {
	branchMap := make(map[string]int)
	for _, regNo := range regNos {
		branch := GetBranchFromRegNo(regNo)
		if branch != "" {
			branchMap[branch]++
		}
	}
	branchCounts := []BranchCount{}
	for branch, count := range branchMap {
		branchCounts = append(branchCounts, BranchCount{Branch: branch, Count: count})
	}
	return branchCounts
}

// Define PlacementsRepository interface for testability
//go:generate mockgen -destination=mock_placements_repo.go -package=placements . PlacementsRepository

type PlacementsRepository interface {
	InsertPlacementCompany(company string, ctc *float64, placementDate string) (int, error)
	InsertBranchwiseRecords(placementID int, branchCounts []BranchCount) error
	GetAllPlacements() ([]PlacementCompany, error)
	GetCompanyBranchMap() ([]CompanyBranch, error)
	GetBranchCompanyMap() ([]BranchCompany, error)
}

type PlacementsService struct {
	repo PlacementsRepository
}

func NewPlacementsService(repo PlacementsRepository) *PlacementsService {
	return &PlacementsService{repo: repo}
}

func (s *PlacementsService) AddPlacement(req PlacementRequest) (PlacementResponse, error) {
	placementDate := req.PlacementDate
	if placementDate == "" {
		placementDate = time.Now().Format("2006-01-02")
	}
	branchCounts := CountBranches(req.Students)
	placementID, err := s.repo.InsertPlacementCompany(req.Company, req.CTC.Value, placementDate)
	if err != nil {
		return PlacementResponse{}, err
	}
	err = s.repo.InsertBranchwiseRecords(placementID, branchCounts)
	if err != nil {
		return PlacementResponse{}, err
	}
	return PlacementResponse{
		PlacementID:   placementID,
		Company:       req.Company,
		CTC:           req.CTC,
		PlacementDate: placementDate,
		BranchCounts:  branchCounts,
	}, nil
}

func (s *PlacementsService) GetAllPlacements() ([]PlacementCompany, error) {
	return s.repo.GetAllPlacements()
}

func (s *PlacementsService) GetCompanyBranchMap() ([]CompanyBranch, error) {
	return s.repo.GetCompanyBranchMap()
}

func (s *PlacementsService) GetBranchCompanyMap() ([]BranchCompany, error) {
	return s.repo.GetBranchCompanyMap()
}
