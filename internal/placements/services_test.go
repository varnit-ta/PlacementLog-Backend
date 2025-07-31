package placements

import (
	"errors"
	"reflect"
	"testing"
)

type mockPlacementsRepo struct {
	InsertPlacementCompanyFunc  func(company string, ctc *float64, placementDate string) (int, error)
	InsertBranchwiseRecordsFunc func(placementID int, branchCounts []BranchCount) error
	GetAllPlacementsFunc        func() ([]PlacementCompany, error)
	GetCompanyBranchMapFunc     func() ([]CompanyBranch, error)
	GetBranchCompanyMapFunc     func() ([]BranchCompany, error)
}

func (m *mockPlacementsRepo) InsertPlacementCompany(company string, ctc *float64, placementDate string) (int, error) {
	return m.InsertPlacementCompanyFunc(company, ctc, placementDate)
}
func (m *mockPlacementsRepo) InsertBranchwiseRecords(placementID int, branchCounts []BranchCount) error {
	return m.InsertBranchwiseRecordsFunc(placementID, branchCounts)
}
func (m *mockPlacementsRepo) GetAllPlacements() ([]PlacementCompany, error) {
	return m.GetAllPlacementsFunc()
}
func (m *mockPlacementsRepo) GetCompanyBranchMap() ([]CompanyBranch, error) {
	return m.GetCompanyBranchMapFunc()
}
func (m *mockPlacementsRepo) GetBranchCompanyMap() ([]BranchCompany, error) {
	return m.GetBranchCompanyMapFunc()
}

func TestPlacementsService_AddPlacement(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mockPlacementsRepo{
			InsertPlacementCompanyFunc: func(company string, ctc *float64, placementDate string) (int, error) {
				return 1, nil
			},
			InsertBranchwiseRecordsFunc: func(placementID int, branchCounts []BranchCount) error {
				return nil
			},
		}
		s := NewPlacementsService(repo)
		ctcValue := 10.5
		resp, err := s.AddPlacement(PlacementRequest{
			Company:       "TestCo",
			CTC:           CTCValue{Value: &ctcValue},
			PlacementDate: "2024-01-01",
			Students:      []string{"22bcs1234", "22bcs5678"},
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if resp.Company != "TestCo" || resp.CTC.Value == nil || *resp.CTC.Value != 10.5 {
			t.Errorf("unexpected response: %+v", resp)
		}
	})
	t.Run("success with NA CTC", func(t *testing.T) {
		repo := &mockPlacementsRepo{
			InsertPlacementCompanyFunc: func(company string, ctc *float64, placementDate string) (int, error) {
				if ctc != nil {
					t.Errorf("expected nil CTC, got %v", *ctc)
				}
				return 1, nil
			},
			InsertBranchwiseRecordsFunc: func(placementID int, branchCounts []BranchCount) error {
				return nil
			},
		}
		s := NewPlacementsService(repo)
		resp, err := s.AddPlacement(PlacementRequest{
			Company:       "TestCo",
			CTC:           CTCValue{Value: nil}, // NA value
			PlacementDate: "2024-01-01",
			Students:      []string{"22bcs1234", "22bcs5678"},
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if resp.Company != "TestCo" || resp.CTC.Value != nil {
			t.Errorf("unexpected response: %+v", resp)
		}
	})
	t.Run("placement company insert error", func(t *testing.T) {
		repo := &mockPlacementsRepo{
			InsertPlacementCompanyFunc: func(company string, ctc *float64, placementDate string) (int, error) {
				return 0, errors.New("insert error")
			},
		}
		s := NewPlacementsService(repo)
		ctcValue := 10.5
		_, err := s.AddPlacement(PlacementRequest{Company: "TestCo", CTC: CTCValue{Value: &ctcValue}, Students: []string{"22bcs1234"}})
		if err == nil || err.Error() != "insert error" {
			t.Errorf("expected insert error, got %v", err)
		}
	})
	t.Run("branchwise records insert error", func(t *testing.T) {
		repo := &mockPlacementsRepo{
			InsertPlacementCompanyFunc: func(company string, ctc *float64, placementDate string) (int, error) {
				return 1, nil
			},
			InsertBranchwiseRecordsFunc: func(placementID int, branchCounts []BranchCount) error {
				return errors.New("branchwise error")
			},
		}
		s := NewPlacementsService(repo)
		ctcValue := 10.5
		_, err := s.AddPlacement(PlacementRequest{Company: "TestCo", CTC: CTCValue{Value: &ctcValue}, Students: []string{"22bcs1234"}})
		if err == nil || err.Error() != "branchwise error" {
			t.Errorf("expected branchwise error, got %v", err)
		}
	})
}

func TestPlacementsService_GetAllPlacements(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		placements := []PlacementCompany{{ID: 1, Company: "TestCo"}}
		repo := &mockPlacementsRepo{
			GetAllPlacementsFunc: func() ([]PlacementCompany, error) { return placements, nil },
		}
		s := NewPlacementsService(repo)
		got, err := s.GetAllPlacements()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !reflect.DeepEqual(got, placements) {
			t.Errorf("expected %v, got %v", placements, got)
		}
	})
	t.Run("repo error", func(t *testing.T) {
		repo := &mockPlacementsRepo{
			GetAllPlacementsFunc: func() ([]PlacementCompany, error) { return nil, errors.New("db error") },
		}
		s := NewPlacementsService(repo)
		_, err := s.GetAllPlacements()
		if err == nil || err.Error() != "db error" {
			t.Errorf("expected db error, got %v", err)
		}
	})
}

func TestPlacementsService_GetCompanyBranchMap(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cb := []CompanyBranch{{Company: "TestCo"}}
		repo := &mockPlacementsRepo{
			GetCompanyBranchMapFunc: func() ([]CompanyBranch, error) { return cb, nil },
		}
		s := NewPlacementsService(repo)
		got, err := s.GetCompanyBranchMap()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !reflect.DeepEqual(got, cb) {
			t.Errorf("expected %v, got %v", cb, got)
		}
	})
	t.Run("repo error", func(t *testing.T) {
		repo := &mockPlacementsRepo{
			GetCompanyBranchMapFunc: func() ([]CompanyBranch, error) { return nil, errors.New("db error") },
		}
		s := NewPlacementsService(repo)
		_, err := s.GetCompanyBranchMap()
		if err == nil || err.Error() != "db error" {
			t.Errorf("expected db error, got %v", err)
		}
	})
}

func TestPlacementsService_GetBranchCompanyMap(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		bc := []BranchCompany{{Branch: "bcs"}}
		repo := &mockPlacementsRepo{
			GetBranchCompanyMapFunc: func() ([]BranchCompany, error) { return bc, nil },
		}
		s := NewPlacementsService(repo)
		got, err := s.GetBranchCompanyMap()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !reflect.DeepEqual(got, bc) {
			t.Errorf("expected %v, got %v", bc, got)
		}
	})
	t.Run("repo error", func(t *testing.T) {
		repo := &mockPlacementsRepo{
			GetBranchCompanyMapFunc: func() ([]BranchCompany, error) { return nil, errors.New("db error") },
		}
		s := NewPlacementsService(repo)
		_, err := s.GetBranchCompanyMap()
		if err == nil || err.Error() != "db error" {
			t.Errorf("expected db error, got %v", err)
		}
	})
}

func TestGetBranchFromRegNo(t *testing.T) {
	cases := []struct {
		regNo string
		want  string
	}{
		{"22bcs1234", "bcs"},
		{"22mec5678", "mec"},
		{"", ""},
		{"12", ""},
		{"22EEE1234", "eee"}, // upper case, should be lower
		{"22bcs", "bcs"},
		{"22", ""},
	}
	for _, c := range cases {
		got := GetBranchFromRegNo(c.regNo)
		if got != c.want {
			t.Errorf("GetBranchFromRegNo(%q) = %q; want %q", c.regNo, got, c.want)
		}
	}
}

func TestCountBranches(t *testing.T) {
	cases := []struct {
		regNos []string
		want   map[string]int
	}{
		{[]string{"22bcs1234", "22bcs5678", "22mec1234"}, map[string]int{"bcs": 2, "mec": 1}},
		{[]string{}, map[string]int{}},
		{[]string{"", "12"}, map[string]int{}},
		{[]string{"22bcs1234", "22bcs1234"}, map[string]int{"bcs": 2}},
	}
	for _, c := range cases {
		got := CountBranches(c.regNos)
		gotMap := map[string]int{}
		for _, bc := range got {
			gotMap[bc.Branch] = bc.Count
		}
		if !reflect.DeepEqual(gotMap, c.want) {
			t.Errorf("CountBranches(%v) = %v; want %v", c.regNos, gotMap, c.want)
		}
	}
}

func TestCTCValue_JSON(t *testing.T) {
	t.Run("marshal numeric CTC", func(t *testing.T) {
		ctcValue := 15.5
		ctc := CTCValue{Value: &ctcValue}
		
		data, err := ctc.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}
		
		expected := "15.5"
		if string(data) != expected {
			t.Errorf("expected %s, got %s", expected, string(data))
		}
	})
	
	t.Run("marshal NA CTC", func(t *testing.T) {
		ctc := CTCValue{Value: nil}
		
		data, err := ctc.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}
		
		expected := `"NA"`
		if string(data) != expected {
			t.Errorf("expected %s, got %s", expected, string(data))
		}
	})
	
	t.Run("unmarshal numeric CTC", func(t *testing.T) {
		var ctc CTCValue
		data := []byte("25.75")
		
		err := ctc.UnmarshalJSON(data)
		if err != nil {
			t.Fatalf("UnmarshalJSON failed: %v", err)
		}
		
		if ctc.Value == nil || *ctc.Value != 25.75 {
			t.Errorf("expected 25.75, got %v", ctc.Value)
		}
	})
	
	t.Run("unmarshal NA CTC", func(t *testing.T) {
		var ctc CTCValue
		data := []byte(`"NA"`)
		
		err := ctc.UnmarshalJSON(data)
		if err != nil {
			t.Fatalf("UnmarshalJSON failed: %v", err)
		}
		
		if ctc.Value != nil {
			t.Errorf("expected nil, got %v", ctc.Value)
		}
	})
	
	t.Run("unmarshal case insensitive NA", func(t *testing.T) {
		var ctc CTCValue
		data := []byte(`"na"`)
		
		err := ctc.UnmarshalJSON(data)
		if err != nil {
			t.Fatalf("UnmarshalJSON failed: %v", err)
		}
		
		if ctc.Value != nil {
			t.Errorf("expected nil, got %v", ctc.Value)
		}
	})
	
	t.Run("unmarshal invalid string", func(t *testing.T) {
		var ctc CTCValue
		data := []byte(`"invalid"`)
		
		err := ctc.UnmarshalJSON(data)
		if err == nil {
			t.Fatal("expected error for invalid string")
		}
		
		expectedError := "invalid CTC string value: invalid, only 'NA' is allowed"
		if err.Error() != expectedError {
			t.Errorf("expected error %s, got %s", expectedError, err.Error())
		}
	})
}
