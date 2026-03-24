package repository

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"github.com/sssseraphim/effective_mobile/internal/models"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(connStr string) (*Repository, error) {
	log.Info("Connecting to database...")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.WithError(err).Error("Failed to open database connection")
		return nil, err
	}

	if err := db.Ping(); err != nil {
		log.WithError(err).Error("Failed to ping database")
		return nil, err
	}

	log.Info("Successfully connected to database")
	return &Repository{db: db}, nil
}

func (r *Repository) Close() error {
	return r.db.Close()
}

func (r *Repository) Create(sub *models.Subscription) error {
	log.WithFields(log.Fields{
		"service_name": sub.ServiceName,
		"user_id":      sub.UserID,
		"price":        sub.Price,
	}).Info("Creating new subscription")

	query := `
        INSERT INTO subscriptions (id, service_name, price, user_id, start_date, end_date)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, created_at, updated_at
    `

	err := r.db.QueryRow(query, sub.ID, sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate).
		Scan(&sub.ID, &sub.CreatedAt, &sub.UpdatedAt)

	if err != nil {
		log.WithError(err).Error("Failed to create subscription")
		return err
	}

	log.WithField("id", sub.ID).Info("Subscription created successfully")
	return nil
}

func (r *Repository) GetByID(id uuid.UUID) (*models.Subscription, error) {
	log.WithField("id", id).Debug("Fetching subscription by ID")

	var sub models.Subscription
	query := `
        SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
        FROM subscriptions
        WHERE id = $1
    `

	err := r.db.QueryRow(query, id).Scan(
		&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID,
		&sub.StartDate, &sub.EndDate, &sub.CreatedAt, &sub.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		log.WithField("id", id).Warn("Subscription not found")
		return nil, nil
	}
	if err != nil {
		log.WithError(err).Error("Failed to fetch subscription")
		return nil, err
	}

	return &sub, nil
}

func (r *Repository) List(limit, offset int) ([]models.Subscription, error) {
	log.WithFields(log.Fields{
		"limit":  limit,
		"offset": offset,
	}).Debug("Listing subscriptions")

	query := `
        SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
        FROM subscriptions
        ORDER BY created_at DESC
        LIMIT $1 OFFSET $2
    `

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		log.WithError(err).Error("Failed to list subscriptions")
		return nil, err
	}
	defer rows.Close()

	var subscriptions []models.Subscription
	for rows.Next() {
		var sub models.Subscription
		err := rows.Scan(
			&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID,
			&sub.StartDate, &sub.EndDate, &sub.CreatedAt, &sub.UpdatedAt,
		)
		if err != nil {
			log.WithError(err).Error("Failed to scan subscription row")
			return nil, err
		}
		subscriptions = append(subscriptions, sub)
	}

	log.WithField("count", len(subscriptions)).Debug("Subscriptions listed")
	return subscriptions, nil
}

func (r *Repository) Update(id uuid.UUID, sub *models.Subscription) error {
	log.WithField("id", id).Info("Updating subscription")

	query := `
        UPDATE subscriptions
        SET service_name = $1, price = $2, start_date = $3, end_date = $4, updated_at = CURRENT_TIMESTAMP
        WHERE id = $5
        RETURNING updated_at
    `

	err := r.db.QueryRow(query, sub.ServiceName, sub.Price, sub.StartDate, sub.EndDate, id).
		Scan(&sub.UpdatedAt)

	if err != nil {
		log.WithError(err).Error("Failed to update subscription")
		return err
	}

	log.WithField("id", id).Info("Subscription updated successfully")
	return nil
}

func (r *Repository) Delete(id uuid.UUID) error {
	log.WithField("id", id).Info("Deleting subscription")

	query := "DELETE FROM subscriptions WHERE id = $1"
	result, err := r.db.Exec(query, id)
	if err != nil {
		log.WithError(err).Error("Failed to delete subscription")
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		log.WithField("id", id).Warn("No subscription found to delete")
		return fmt.Errorf("subscription not found")
	}

	log.WithField("id", id).Info("Subscription deleted successfully")
	return nil
}

func (r *Repository) GetTotalCost(userID uuid.UUID, serviceName, startDate, endDate string) (int, error) {
	log.WithFields(log.Fields{
		"user_id":      userID,
		"service_name": serviceName,
		"start_date":   startDate,
		"end_date":     endDate,
	}).Info("Calculating total cost")

	query := `
        SELECT COALESCE(SUM(price), 0)
        FROM subscriptions
        WHERE (user_id = $1 OR $1 IS NULL)
          AND (service_name = $2 OR $2 IS NULL)
          AND start_date >= $3
          AND (end_date IS NULL OR end_date <= $4)
    `

	var total int
	err := r.db.QueryRow(query, userID, serviceName, startDate, endDate).Scan(&total)
	if err != nil {
		log.WithError(err).Error("Failed to calculate total cost")
		return 0, err
	}

	log.WithField("total_cost", total).Info("Total cost calculated")
	return total, nil
}
