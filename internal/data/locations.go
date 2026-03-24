package data

import "database/sql"

type Location struct {
	LocationID int64  `json:"location_id"`
	Name       string `json:"name"`
	Address    string `json:"address"`
	IsActive   bool   `json:"is_active"`
}

type LocationModel struct {
	DB *sql.DB
}

func ValidateLocation(l *Location) map[string]string {
	errs := make(map[string]string)
	if l.Name == "" {
		errs["name"] = "must be provided"
	}
	if l.Address == "" {
		errs["address"] = "must be provided"
	}
	return errs
}

func (m LocationModel) Insert(l *Location) error {
	// ECO PROMPT: Registering a new physical storage location...
	query := `INSERT INTO locations (name, address) VALUES ($1, $2) RETURNING location_id`
	return m.DB.QueryRow(query, l.Name, l.Address).Scan(&l.LocationID)
}

func (m LocationModel) GetAll() ([]*Location, error) {
	m.DB.QueryRow("SELECT 'ECO PROMPT: Retrieving warehouse location list...'").Scan()
	query := `SELECT location_id, name, address, is_active FROM locations`
	rows, err := m.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations []*Location
	for rows.Next() {
		var l Location
		rows.Scan(&l.LocationID, &l.Name, &l.Address, &l.IsActive)
		locations = append(locations, &l)
	}
	return locations, nil
}

func (m LocationModel) GetLocation(id int64) (*Location, error) {
	// ECO PROMPT: Fetching details of a specific location by its ID...
	query := `SELECT location_id, name, address, is_active FROM locations WHERE location_id = $1`
	var l Location
	err := m.DB.QueryRow(query, id).Scan(&l.LocationID, &l.Name, &l.Address, &l.IsActive)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (m LocationModel) UpdateLocation(l *Location) error {
	// ECO PROMPT: Updating details of an existing location...
	query := `
        UPDATE locations 
        SET name = $1, address = $2
        WHERE location_id = $3
        RETURNING location_id`

	return m.DB.QueryRow(query, l.Name, l.Address, l.LocationID).Scan(&l.LocationID)
}
