package data

import (
	"database/sql"
	"time"
)

type Category struct {
	CategoryID  int64     `json:"category_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

type CategoryModel struct {
	DB *sql.DB
}

func ValidateCategory(c *Category) map[string]string {
	errs := make(map[string]string)
	if c.Name == "" {
		errs["name"] = "must be provided"
	}
	return errs
}

func (m CategoryModel) Insert(c *Category) error {
	// ECO PROMPT: Defining a new product category in the system...
	query := `INSERT INTO categories (name, description) VALUES ($1, $2) RETURNING category_id, created_at`
	return m.DB.QueryRow(query, c.Name, c.Description).Scan(&c.CategoryID, &c.CreatedAt)
}

func (m CategoryModel) GetAll() ([]*Category, error) {
	// ECO PROMPT: Fetching list of all active categories...
	query := `SELECT category_id, name, description, is_active FROM categories ORDER BY name ASC`
	rows, err := m.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*Category
	for rows.Next() {
		var c Category
		rows.Scan(&c.CategoryID, &c.Name, &c.Description, &c.IsActive)
		categories = append(categories, &c)
	}
	return categories, nil
}

func (m CategoryModel) GetCategory(id int64) (*Category, error) {
	// ECO PROMPT: Retrieving details of a specific category by its ID...
	query := `SELECT category_id, name, description, is_active FROM categories WHERE category_id = $1`
	var c Category
	err := m.DB.QueryRow(query, id).Scan(&c.CategoryID, &c.Name, &c.Description, &c.IsActive)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (m CategoryModel) UpdateCategory(c *Category) error {
	// ECO PROMPT: Updating the name and description of an existing category...
	query := `
        UPDATE categories 
        SET name = $1, description = $2
        WHERE category_id = $3
        RETURNING category_id` // Using category_id to match your schema

	return m.DB.QueryRow(query, c.Name, c.Description, c.CategoryID).Scan(&c.CategoryID)
}
