package outbound

import (
	"context"
	"errors"
	"fmt"

	"github.com/htan06/Moments/user/internal/adapter/outbound/sqlc/queries"
	"github.com/htan06/Moments/user/internal/application/readmodel"
	"github.com/htan06/Moments/user/internal/domain"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresProfileRepository struct {
	conn    *pgxpool.Pool
	queries *queries.Queries
}

func NewPostgresUserRepository(conn *pgxpool.Pool) *PostgresProfileRepository {
	return &PostgresProfileRepository{
		conn:    conn,
		queries: queries.New(conn),
	}
}

func (pur *PostgresProfileRepository) CreateProfile(ctx context.Context, p domain.Profile) error {
	err := pur.queries.CreateProfile(ctx, queries.CreateProfileParams{
		UserID:         p.UserID(),
		Name:           p.Name(),
		Username:       p.Username(),
		FollowersCount: p.FollowersCount(),
		FollowingCount: p.FollowingCount(),
		PostsCount:     p.PostsCount(),
	})
	if err != nil {
		return fmt.Errorf("PostgresUserRepository.CreateProfile: %w", err)
	}
	return nil
}

func (pur *PostgresProfileRepository) GetAvatarKeyByUserID(ctx context.Context, userID int64) (*string, error) {
	avatarID, err := pur.queries.GetAvatarKeyByUserID(ctx, userID)

	if err != nil {
		return nil, fmt.Errorf("PostgresUserRepository.GetAvatarIDByUserID: %w", err)
	}
	return &avatarID.String, nil
}

func (pur *PostgresProfileRepository) UpdateProfile(ctx context.Context, userID int64, fieldUpdates map[string]interface{}) error {
	queryBuilder := squirrel.Update("profile.users").
		SetMap(fieldUpdates).
		Where("user_id = ?", userID).
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("PostgresUserRepository.UpdateProfile: %w", err)
	}

	if _, err := pur.conn.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("PostgresUserRepository.UpdateProfile: %w", err)
	}
	return nil
}

func (pur *PostgresProfileRepository) GetSelfProfileByUsername(ctx context.Context, username string) (readmodel.Profile, error) {
	row, err := pur.queries.GetSelfProfileByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return readmodel.Profile{}, domain.ErrUserNotFound
		}
		return readmodel.Profile{}, fmt.Errorf("PostgresUserRepository.GetSelfProfileByUsername: %w", err)
	}

	return readmodel.Profile{
		IsOwner:            true,
		UserID:             row.UserID,
		Username:           row.Username,
		Name:               row.Name,
		AvatarURL:          &row.AvatarKey.String,
		AvatarThumbnailURL: &row.AvatarThumbnailKey.String,
		Bio:                &row.Bio.String,
		FollowersCount:     row.FollowersCount,
		FollowingCount:     row.FollowingCount,
		PostsCount:         row.PostsCount,
		Relationship:       "",
	}, nil
}

func (pur *PostgresProfileRepository) GetOtherProfileByUsername(ctx context.Context, currentUserID int64, targetUsername string) (readmodel.Profile, error) {
	row, err := pur.queries.GetOtherProfileByUsername(ctx, queries.GetOtherProfileByUsernameParams{
		CurrentUserID:  currentUserID,
		TargetUsername: targetUsername,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return readmodel.Profile{}, domain.ErrUserNotFound
		}
		return readmodel.Profile{}, fmt.Errorf("PostgresUserRepository.GetOtherProfileByUsername: %w", err)
	}

	return readmodel.Profile{
		IsOwner:            true,
		UserID:             row.UserID,
		Username:           row.Username,
		Name:               row.Name,
		AvatarURL:          &row.AvatarKey.String,
		AvatarThumbnailURL: &row.AvatarThumbnailKey.String,
		Bio:                &row.Bio.String,
		FollowersCount:     row.FollowersCount,
		FollowingCount:     row.FollowingCount,
		PostsCount:         row.PostsCount,
		Relationship:       row.Relationship,
	}, nil
}

func (pur *PostgresProfileRepository) UpdateAvatarKeyAndAvatarThumbnailKey(ctx context.Context, userID int64, avatarKey string, avatarThumbnailKey string) error {
	err := pur.queries.UpdateAvatarKeyAndAvatarThumbnailKey(ctx, queries.UpdateAvatarKeyAndAvatarThumbnailKeyParams{
		AvatarKey:          pgtype.Text{String: avatarKey, Valid: true},
		AvatarThumbnailKey: pgtype.Text{String: avatarThumbnailKey, Valid: true},
		UserID:             userID,
	})
	if err != nil {
		return fmt.Errorf("PostgresUserRepository.UpdateAvatarID: %w", err)
	}
	return nil
}

func (pur *PostgresProfileRepository) FindProfilesByUsername(ctx context.Context, username string) ([]readmodel.ProfileSummary, error) {
	rows, err := pur.queries.FindProfilesByUsername(ctx, pgtype.Text{
		String: username,
		Valid:  true,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []readmodel.ProfileSummary{}, domain.ErrUserNotFound
		}
		return []readmodel.ProfileSummary{}, fmt.Errorf("PostgresUserRepository.GetOtherProfileByUsername: %w", err)
	}

	var profileSummaries []readmodel.ProfileSummary
	for _, r := range rows {
		profileSummaries = append(profileSummaries, readmodel.ProfileSummary{
			UserID:             r.UserID,
			Username:           r.Username,
			Name:               r.Name,
			AvatarThumbnailURL: &r.AvatarThumbnailKey.String,
		})
	}
	return profileSummaries, nil
}
func (p *PostgresProfileRepository) GetProfilesSummary(ctx context.Context, currentUserId int64, ids []int64) ([]readmodel.ProfileSummary, error) {
	rows, err := p.queries.GetProfilesByUserIDs(ctx, queries.GetProfilesByUserIDsParams{
		CurrentUserID: currentUserId,
		Ids:           ids,
	})

	if err != nil {
		return nil, fmt.Errorf("PostgresProfileRepository.GetProfilesSummary: %w", err)
	}

	var profiles []readmodel.ProfileSummary

	for _, p := range rows {
		profiles = append(profiles, readmodel.ProfileSummary{
			UserID:             p.UserID,
			Username:           p.Username,
			Name:               p.Name,
			AvatarThumbnailURL: &p.AvatarThumbnailKey.String,
		})
	}
	return profiles, nil
}

// func (p *PostgresUserRepository) IncPostCount(ctx context.Context, userID int64) error {
// 	qry := `UPDATE profile.users SET posts_count = posts_count + 1 WHERE user_id = $1;`

// 	tag, err := p.conn.Exec(ctx, qry, userID)
// 	if err != nil {
// 		return fmt.Errorf("PostgresUserRepository.IncPostCount: %w", err)
// 	}

// 	if tag.RowsAffected() != 1 {
// 		return fmt.Errorf("PostgresUserRepository.IncPostCount: %w", err)
// 	}

// 	return nil
// }

// func (p *PostgresUserRepository) DecPostCount(ctx context.Context, userID int64) error {
// 	qry := `UPDATE profile.users SET posts_count = posts_count - 1 WHERE user_id = $1;`

// 	tag, err := p.conn.Exec(ctx, qry, userID)
// 	if err != nil {
// 		return fmt.Errorf("PostgresUserRepository.DeccPostCount: %w", err)
// 	}

// 	if tag.RowsAffected() != 1 {
// 		return fmt.Errorf("PostgresUserRepository.DecPostCount: %w", err)
// 	}

// 	return nil
// }

// func (p *PostgresUserRepository) IncFollowCount(ctx context.Context, followerID int64, followingID int64) error {
// 	tx, err := p.conn.Begin(ctx)
// 	if err != nil {
// 		return fmt.Errorf("PostgresUserRepository.IncFollowCount: %w", err)
// 	}
// 	defer tx.Rollback(ctx)

// 	incFollowing := `UPDATE profile.users SET following_count = following_count + 1 WHERE user_id = $1;`
// 	incFollowers := `UPDATE profile.users SET followers_count = followers_count + 1 WHERE user_id = $1;`

// 	batch := pgx.Batch{}
// 	batch.Queue(incFollowing, followerID)
// 	batch.Queue(incFollowers, followingID)

// 	if err := tx.SendBatch(ctx, &batch).Close(); err != nil {
// 		return fmt.Errorf("PostgresUserRepository.IncFollowCount: %w", err)
// 	}

// 	if err := tx.Commit(ctx); err != nil {
// 		return fmt.Errorf("PostgresUserRepository.IncFollowCount: %w", err)
// 	}

// 	return nil
// }

// func (p *PostgresUserRepository) DecFollowCount(ctx context.Context, followerID int64, followingID int64) error {
// 	tx, err := p.conn.Begin(ctx)
// 	if err != nil {
// 		return fmt.Errorf("PostgresUserRepository.DecFollowCount: %w", err)
// 	}
// 	defer tx.Rollback(ctx)

// 	decFollowing := `UPDATE profile.users SET following_count = following_count - 1 WHERE user_id = $1;`
// 	decFollowers := `UPDATE profile.users SET followers_count = followers_count - 1 WHERE user_id = $1;`

// 	batch := pgx.Batch{}
// 	batch.Queue(decFollowing, followerID)
// 	batch.Queue(decFollowers, followingID)

// 	if err := tx.SendBatch(ctx, &batch).Close(); err != nil {
// 		return fmt.Errorf("PostgresUserRepository.DecFollowCount: %w", err)
// 	}

// 	if err := tx.Commit(ctx); err != nil {
// 		return fmt.Errorf("PostgresUserRepository.DecFollowCount: %w", err)
// 	}

// 	return nil
// }
