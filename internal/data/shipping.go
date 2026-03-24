package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type Shipping struct {
	MethodID     int64     `json:"method_id"`
	ProviderName string    `json:"provider_name"`
	ServiceType  string    `json:"service_type"`
	BaseRate     float64   `json:"base_rate"`
	ContactPhone *string   `json:"contact_phone,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type ShippingModel struct {
	DB *sql.DB
}

func ValidateShipping(s *Shipping) []string {
	var errs []string

	if s.ProviderName == "" {
		errs = append(errs, "Provider name is required")
	}
	if s.ServiceType == "" {
		errs = append(errs, "Service type is required")
	}
	if s.BaseRate < 0 {
		errs = append(errs, "Base rate cannot be negative")
	}

	return errs
}

func (m *ShippingModel) CreateShipping(ctx context.Context, s *Shipping) error {
	query := `
    INSERT INTO shipping_methods (provider_name, service_type, base_rate, contact_phone)
    VALUES ($1, $2, $3, $4)
    RETURNING method_id, created_at`

	return m.DB.QueryRowContext(ctx, query,
		s.ProviderName,
		s.ServiceType,
		s.BaseRate,
		s.ContactPhone,
	).Scan(&s.MethodID, &s.CreatedAt)
}

func (m *ShippingModel) GetShipping(ctx context.Context, id int64) (*Shipping, error) {
	query := `
	SELECT method_id, provider_name, service_type, base_rate, contact_phone, created_at
	FROM shipping_methods
	WHERE method_id = $1`
	var shipping Shipping
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&shipping.MethodID,
		&shipping.ProviderName,
		&shipping.ServiceType,
		&shipping.BaseRate,
		&shipping.ContactPhone,
		&shipping.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &shipping, nil
}

func (m *ShippingModel) UpdateShipping(ctx context.Context, shipping Shipping) error {
	query := `
	UPDATE shipping_methods
	SET provider_name = $1, service_type = $2, base_rate = $3, contact_phone = $4
	WHERE method_id = $5`
	result, err := m.DB.ExecContext(ctx, query,
		shipping.ProviderName,
		shipping.ServiceType,
		shipping.BaseRate,
		shipping.ContactPhone,
		shipping.MethodID,
	)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
