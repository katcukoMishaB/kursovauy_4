package main

import (
	"database/sql"
	"errors"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (User) TableName() string             { return "users" }
func (UserRole) TableName() string         { return "user_roles" }
func (OrganizerRequest) TableName() string { return "organizer_requests" }
func (Group) TableName() string            { return "groups" }

func notFoundOr(err error, empty bool) error {
	if errors.Is(err, gorm.ErrRecordNotFound) || empty {
		return sql.ErrNoRows
	}
	return err
}

func (r *UserRepository) CreateUser(firstName, lastName, email, hashedPassword string) (string, error) {
	var userID string
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Raw(
			`INSERT INTO users (first_name, last_name, email, password, registration_date, user_type)
			 VALUES (?, ?, ?, ?, CURRENT_DATE, 'student') RETURNING id`,
			firstName, lastName, email, hashedPassword,
		).Scan(&userID).Error; err != nil {
			return err
		}
		return tx.Exec(
			`INSERT INTO user_roles (user_id, is_participant, is_organizer, is_admin) VALUES (?, true, false, false)`,
			userID,
		).Error
	})
	return userID, err
}

func (r *UserRepository) FindUserByEmail(email string) (User, string, error) {
	var row struct {
		ID               string    `gorm:"column:id"`
		FirstName        string    `gorm:"column:first_name"`
		LastName         string    `gorm:"column:last_name"`
		Email            string    `gorm:"column:email"`
		Password         string    `gorm:"column:password"`
		RegistrationDate string    `gorm:"column:registration_date"`
		Status           bool      `gorm:"column:status"`
		UserType         string    `gorm:"column:user_type"`
		GroupID          *string   `gorm:"column:group_id"`
		GroupName        *string   `gorm:"column:group_name"`
	}
	err := r.db.Raw(
		`SELECT u.id, u.first_name, u.last_name, u.email, u.password,
		        u.registration_date::text, u.status, u.user_type, u.group_id, g.name AS group_name
		 FROM users u
		 LEFT JOIN groups g ON g.id = u.group_id
		 WHERE u.email = ?`,
		email,
	).Scan(&row).Error
	if err != nil {
		return User{}, "", err
	}
	if row.ID == "" {
		return User{}, "", sql.ErrNoRows
	}
	user := User{
		ID:        row.ID,
		FirstName: row.FirstName,
		LastName:  row.LastName,
		Email:     row.Email,
		Status:    row.Status,
		UserType:  row.UserType,
		GroupID:   row.GroupID,
		GroupName: row.GroupName,
	}
	return user, row.Password, nil
}

func (r *UserRepository) FindUserByID(id string) (User, error) {
	var user User
	err := r.db.Raw(
		`SELECT u.id, u.first_name, u.last_name, u.email, u.registration_date, u.status,
		        u.user_type, u.group_id, g.name AS group_name
		 FROM users u
		 LEFT JOIN groups g ON g.id = u.group_id
		 WHERE u.id = ?`,
		id,
	).Scan(&user).Error
	if err != nil {
		return user, err
	}
	if user.ID == "" {
		return user, sql.ErrNoRows
	}
	return user, nil
}

func (r *UserRepository) GetUserRole(userID string) (UserRole, error) {
	var role UserRole
	err := r.db.Raw(
		`SELECT user_id, is_participant, is_organizer, is_admin FROM user_roles WHERE user_id = ?`,
		userID,
	).Scan(&role).Error
	if err != nil {
		return role, err
	}
	if role.UserID == "" {
		return role, sql.ErrNoRows
	}
	return role, nil
}

func (r *UserRepository) UpdateProfile(userID, firstName, lastName string, hashedPassword *string, groupID *string) error {
	switch {
	case hashedPassword != nil && groupID != nil:
		return r.db.Exec(
			`UPDATE users SET first_name = ?, last_name = ?, password = ?, group_id = NULLIF(?, '')::uuid WHERE id = ?`,
			firstName, lastName, *hashedPassword, *groupID, userID,
		).Error
	case hashedPassword != nil:
		return r.db.Exec(
			`UPDATE users SET first_name = ?, last_name = ?, password = ? WHERE id = ?`,
			firstName, lastName, *hashedPassword, userID,
		).Error
	case groupID != nil:
		return r.db.Exec(
			`UPDATE users SET first_name = ?, last_name = ?, group_id = NULLIF(?, '')::uuid WHERE id = ?`,
			firstName, lastName, *groupID, userID,
		).Error
	default:
		return r.db.Exec(
			`UPDATE users SET first_name = ?, last_name = ? WHERE id = ?`,
			firstName, lastName, userID,
		).Error
	}
}

func (r *UserRepository) FindActiveOrganizerRequest(userID, requestType string) (string, error) {
	var id string
	err := r.db.Raw(
		`SELECT id FROM organizer_requests
		 WHERE user_id = ? AND status = 'в рассмотрении' AND request_type = ? LIMIT 1`,
		userID, requestType,
	).Scan(&id).Error
	if err != nil {
		return "", err
	}
	if id == "" {
		return "", sql.ErrNoRows
	}
	return id, nil
}

func (r *UserRepository) GetLatestRequest(userID, requestType string) (OrganizerRequest, error) {
	var rq OrganizerRequest
	err := r.db.Raw(
		`SELECT id, user_id, experience_description, resume_url, submission_date, status, request_type
		 FROM organizer_requests
		 WHERE user_id = ? AND request_type = ?
		 ORDER BY submission_date DESC LIMIT 1`,
		userID, requestType,
	).Scan(&rq).Error
	if err != nil {
		return rq, err
	}
	if rq.ID == "" {
		return rq, sql.ErrNoRows
	}
	return rq, nil
}

func (r *UserRepository) CreateOrganizerRequest(userID, experience, resumeURL, requestType string) (string, error) {
	var resArg interface{}
	if resumeURL != "" {
		resArg = resumeURL
	}
	var id string
	err := r.db.Raw(
		`INSERT INTO organizer_requests (user_id, experience_description, resume_url, submission_date, status, request_type)
		 VALUES (?, ?, ?, CURRENT_DATE, 'в рассмотрении', ?) RETURNING id`,
		userID, experience, resArg, requestType,
	).Scan(&id).Error
	return id, err
}

func (r *UserRepository) ListOrganizerRequests(filterType string) ([]OrganizerRequestExtended, error) {
	q := `
		SELECT or_.id, or_.user_id, or_.experience_description, or_.resume_url,
			or_.submission_date, or_.status, or_.request_type,
			COALESCE(u.first_name, ''), COALESCE(u.last_name, ''), COALESCE(u.email, '')
		FROM organizer_requests or_
		LEFT JOIN users u ON u.id = or_.user_id`
	args := []interface{}{}
	if filterType != "" {
		q += " WHERE or_.request_type = ?"
		args = append(args, filterType)
	}
	q += " ORDER BY or_.submission_date DESC"
	rows, err := r.db.Raw(q, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []OrganizerRequestExtended{}
	for rows.Next() {
		var rq OrganizerRequestExtended
		if err := rows.Scan(&rq.ID, &rq.UserID, &rq.ExperienceDesc, &rq.ResumeURL,
			&rq.SubmissionDate, &rq.Status, &rq.RequestType,
			&rq.FirstName, &rq.LastName, &rq.Email); err != nil {
			continue
		}
		skills := []string{}
		_ = r.db.Raw(
			`SELECT tc.name FROM user_skills us JOIN tag_catalog tc ON tc.id = us.tag_id
			 WHERE us.user_id = ? ORDER BY tc.name`, rq.UserID).Scan(&skills).Error
		rq.Skills = skills
		interests := []string{}
		_ = r.db.Raw(
			`SELECT pc.name FROM user_interests ui JOIN project_categories pc ON pc.id = ui.category_id
			 WHERE ui.user_id = ? ORDER BY pc.name`, rq.UserID).Scan(&interests).Error
		rq.Interests = interests
		out = append(out, rq)
	}
	return out, nil
}

func (r *UserRepository) ApproveOrganizerRequest(requestID string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var userID, requestType string
		row := tx.Raw(`SELECT user_id, request_type FROM organizer_requests WHERE id = ?`, requestID).Row()
		if err := row.Scan(&userID, &requestType); err != nil {
			return err
		}
		if userID == "" {
			return sql.ErrNoRows
		}
		if err := tx.Exec(`UPDATE organizer_requests SET status = 'одобрена' WHERE id = ?`, requestID).Error; err != nil {
			return err
		}
		if err := tx.Exec(`UPDATE user_roles SET is_organizer = true WHERE user_id = ?`, userID).Error; err != nil {
			return err
		}
		if requestType == "teacher" {
			if err := tx.Exec(`UPDATE users SET user_type = 'teacher' WHERE id = ?`, userID).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *UserRepository) RejectOrganizerRequest(requestID string) error {
	return r.db.Exec(
		`UPDATE organizer_requests SET status = 'отклонена' WHERE id = ?`,
		requestID,
	).Error
}

func (r *UserRepository) ListUsers() ([]User, error) {
	users := []User{}
	err := r.db.Raw(
		`SELECT u.id, u.first_name, u.last_name, u.email, u.registration_date, u.status,
		        u.user_type, u.group_id, g.name AS group_name
		 FROM users u
		 LEFT JOIN groups g ON g.id = u.group_id
		 ORDER BY u.registration_date DESC`,
	).Scan(&users).Error
	return users, err
}

func (r *UserRepository) UpdateUserStatus(userID string, status bool) error {
	return r.db.Exec(`UPDATE users SET status = ? WHERE id = ?`, status, userID).Error
}

func (r *UserRepository) ListGroups() ([]Group, error) {
	groups := []Group{}
	err := r.db.Raw(`SELECT id, name FROM groups ORDER BY name`).Scan(&groups).Error
	return groups, err
}

func (r *UserRepository) CreateGroup(name string) (string, error) {
	var id string
	err := r.db.Raw(`INSERT INTO groups (name) VALUES (?) RETURNING id`, name).Scan(&id).Error
	return id, err
}

func (r *UserRepository) UpdateGroup(id, name string) error {
	return r.db.Exec(`UPDATE groups SET name = ? WHERE id = ?`, name, id).Error
}

func (r *UserRepository) DeleteGroup(id string) error {
	return r.db.Exec(`DELETE FROM groups WHERE id = ?`, id).Error
}
