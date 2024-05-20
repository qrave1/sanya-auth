package repository

import (
	"context"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"sanya-auth/internal/config"
	"sanya-auth/internal/domain"
)

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(cfg *config.Config) (*GormRepository, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.Username,
		cfg.Postgres.Password,
		cfg.Postgres.Database,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	err = db.AutoMigrate(&domain.User{}, &domain.Token{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return &GormRepository{db: db}, nil
}

func (repo *GormRepository) CreateUser(ctx context.Context, user *domain.User) error {
	return repo.db.WithContext(ctx).Create(user).Error
}

func (repo *GormRepository) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	err := repo.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (repo *GormRepository) SetToken(ctx context.Context, token *domain.Token) error {
	return repo.db.WithContext(ctx).Create(token).Error
}

func (repo *GormRepository) GetTokenByUserID(ctx context.Context, userID uint) (*domain.Token, error) {
	var token domain.Token
	err := repo.db.WithContext(ctx).Where("user_id = ?", userID).Order("expiration DESC").First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}
