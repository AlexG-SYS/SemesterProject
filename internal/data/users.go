package data

import (
	"database/sql"
	"errors"
	"time"
)

type User struct {
	UserID    int64     `json:"user_id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Never export the hash in JSON!
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	Profile   *Profile  `json:"profile,omitempty"`
}

type Profile struct {
	ProfileID   int64     `json:"profile_id"`
	UserID      int64     `json:"user_id"`
	FullName    string    `json:"full_name"`
	Phone       string    `json:"phone"`
	Address     string    `json:"address"`
	District    string    `json:"district"`
	TownVillage string    `json:"town_village"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ProfileModel struct {
	DB *sql.DB
}

type UserModel struct {
	DB *sql.DB
}

func ValidateProfile(p *Profile) map[string]string {
	errs := make(map[string]string)
	if p.FullName == "" {
		errs["full_name"] = "must be provided"
	}
	if p.Phone == "" {
		errs["phone"] = "must be provided"
	}
	if p.Address == "" {
		errs["address"] = "must be provided"
	}
	if p.District == "" {
		errs["district"] = "must be provided"
	}
	if p.TownVillage == "" {
		errs["town_village"] = "must be provided"
	}
	return errs
}

func ValidateUser(u *User) map[string]string {
	errs := make(map[string]string)
	if u.Email == "" {
		errs["email"] = "must be provided"
	}
	if u.Password == "" {
		errs["password"] = "must be provided"
	}
	if u.Role == "" {
		errs["role"] = "must be provided"
	}
	return errs
}

func (m UserModel) Insert(user *User, profile *Profile) error {
	// 1. Start the transaction
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}
	// Defer a rollback in case of error
	defer tx.Rollback()

	// 2. Insert the User
	userQuery := `
        INSERT INTO users (email, password_hash, role) 
        VALUES ($1, $2, $3) 
        RETURNING user_id, created_at`

	err = tx.QueryRow(userQuery, user.Email, user.Password, user.Role).Scan(&user.UserID, &user.CreatedAt)
	if err != nil {
		return err
	}

	profile.UserID = user.UserID // Link the profile to the newly created user

	// 3. Insert the Profile using the NEW user.UserID
	profileQuery := `
        INSERT INTO profiles (user_id, full_name, phone, address, district, town_village)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING profile_id, created_at, updated_at`

	err = tx.QueryRow(profileQuery,
		user.UserID, // Link established here
		profile.FullName,
		profile.Phone,
		profile.Address,
		profile.District,
		profile.TownVillage,
	).Scan(&profile.ProfileID, &profile.CreatedAt, &profile.UpdatedAt)

	if err != nil {
		return err
	}

	// 4. Commit the transaction
	return tx.Commit()
}

func (m ProfileModel) Get(id int64) (*User, error) {
	if id < 1 {
		return nil, errors.New("record not found")
	}

	// SQL JOIN to get both User and Profile data
	query := `
        SELECT u.user_id, u.email, u.role, u.created_at,
               p.profile_id, p.full_name, p.phone, p.address, p.district, p.town_village, p.created_at, p.updated_at
        FROM users u
        LEFT JOIN profiles p ON u.user_id = p.user_id
        WHERE u.user_id = $1`

	var user User
	var profile Profile

	err := m.DB.QueryRow(query, id).Scan(
		&user.UserID, &user.Email, &user.Role, &user.CreatedAt,
		&profile.ProfileID, &profile.FullName, &profile.Phone, &profile.Address, &profile.District, &profile.TownVillage, &profile.CreatedAt, &profile.UpdatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, errors.New("record not found")
		default:
			return nil, err
		}
	}

	// Attach the profile to the user struct
	user.Profile = &profile

	return &user, nil
}

func (m ProfileModel) Update(p *Profile) error {
	query := `
        UPDATE profiles 
        SET full_name = $1, phone = $2, address = $3, district = $4, town_village = $5, updated_at = NOW()
        WHERE profile_id = $6
        RETURNING updated_at`

	args := []any{
		p.FullName,
		p.Phone,
		p.Address,
		p.District,
		p.TownVillage,
		p.ProfileID,
	}

	return m.DB.QueryRow(query, args...).Scan(&p.UpdatedAt)
}
