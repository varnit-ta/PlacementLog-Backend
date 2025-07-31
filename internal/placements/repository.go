package placements

import (
	"database/sql"
	"fmt"
)

type PlacementsRepo struct {
	db *sql.DB
}

func NewPlacementsRepo(db *sql.DB) *PlacementsRepo {
	return &PlacementsRepo{db: db}
}

func (r *PlacementsRepo) InsertPlacementCompany(company string, ctc *float64, placementDate string) (int, error) {
	var id int
	query := `INSERT INTO placement_companies (company, ctc, placement_date) VALUES ($1, $2, $3) RETURNING id`
	err := r.db.QueryRow(query, company, ctc, placementDate).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert placement company: %w", err)
	}
	return id, nil
}

func (r *PlacementsRepo) InsertBranchwiseRecords(placementID int, branchCounts []BranchCount) error {
	query := `INSERT INTO placement_branchwise_record (placement_id, branch, count) VALUES ($1, $2, $3)`
	for _, bc := range branchCounts {
		_, err := r.db.Exec(query, placementID, bc.Branch, bc.Count)
		if err != nil {
			return fmt.Errorf("failed to insert branchwise record: %w", err)
		}
	}
	return nil
}

func (r *PlacementsRepo) GetAllPlacements() ([]PlacementCompany, error) {
	placements := []PlacementCompany{}
	rows, err := r.db.Query(`SELECT id, company, ctc, placement_date, created_at FROM placement_companies ORDER BY placement_date DESC`)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch placements: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var p PlacementCompany
		var ctcNullable sql.NullFloat64
		err := rows.Scan(&p.ID, &p.Company, &ctcNullable, &p.PlacementDate, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		
		// Convert sql.NullFloat64 to CTCValue
		if ctcNullable.Valid {
			p.CTC = CTCValue{Value: &ctcNullable.Float64}
		} else {
			p.CTC = CTCValue{Value: nil}
		}
		
		branchRows, err := r.db.Query(`SELECT branch, count FROM placement_branchwise_record WHERE placement_id = $1`, p.ID)
		if err != nil {
			return nil, err
		}
		var branchCounts []BranchCount
		for branchRows.Next() {
			var bc BranchCount
			if err := branchRows.Scan(&bc.Branch, &bc.Count); err != nil {
				return nil, err
			}
			branchCounts = append(branchCounts, bc)
		}
		branchRows.Close()
		p.BranchCounts = branchCounts
		placements = append(placements, p)
	}
	return placements, nil
}

type CompanyBranch struct {
	Company  string        `json:"company"`
	Branches []BranchCount `json:"branches"`
}

type BranchCompany struct {
	Branch    string         `json:"branch"`
	Companies []CompanyCount `json:"companies"`
}

type CompanyCount struct {
	Company string `json:"company"`
	Count   int    `json:"count"`
}

func (r *PlacementsRepo) GetCompanyBranchMap() ([]CompanyBranch, error) {
	rows, err := r.db.Query(`
		SELECT pc.company, pbr.branch, SUM(pbr.count) as total
		FROM placement_companies pc
		JOIN placement_branchwise_record pbr ON pc.id = pbr.placement_id
		GROUP BY pc.company, pbr.branch
		ORDER BY pc.company, pbr.branch
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	companyMap := make(map[string][]BranchCount)
	for rows.Next() {
		var company, branch string
		var count int
		if err := rows.Scan(&company, &branch, &count); err != nil {
			return nil, err
		}
		companyMap[company] = append(companyMap[company], BranchCount{Branch: branch, Count: count})
	}
	var result []CompanyBranch
	for company, branches := range companyMap {
		result = append(result, CompanyBranch{Company: company, Branches: branches})
	}
	return result, nil
}

func (r *PlacementsRepo) GetBranchCompanyMap() ([]BranchCompany, error) {
	rows, err := r.db.Query(`
		SELECT pbr.branch, pc.company, SUM(pbr.count) as total
		FROM placement_companies pc
		JOIN placement_branchwise_record pbr ON pc.id = pbr.placement_id
		GROUP BY pbr.branch, pc.company
		ORDER BY pbr.branch, pc.company
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	branchMap := make(map[string][]CompanyCount)
	for rows.Next() {
		var branch, company string
		var count int
		if err := rows.Scan(&branch, &company, &count); err != nil {
			return nil, err
		}
		branchMap[branch] = append(branchMap[branch], CompanyCount{Company: company, Count: count})
	}
	var result []BranchCompany
	for branch, companies := range branchMap {
		result = append(result, BranchCompany{Branch: branch, Companies: companies})
	}
	return result, nil
}

// Ensure PlacementsRepo implements PlacementsRepository
var _ PlacementsRepository = (*PlacementsRepo)(nil)
