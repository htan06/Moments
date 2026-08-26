package infra

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/htan06/Moments/internal/errs"
	"github.com/htan06/Moments/internal/module/user/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresUserRepository struct {
	conn *pgxpool.Pool
}

func NewPostgresUserRepository(conn *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{
		conn: conn,
	}
}

func (pur *PostgresUserRepository) CreateProfile(ctx context.Context, p domain.Profile) error {
	qry := `INSERT INTO profile.users (user_id, name, username, followers_count, following_count, posts_count)
			VALUES ($1, $2, $3, $4, $5, $6);`

	_, err := pur.conn.Exec(ctx, qry, p.UserID, p.Name, p.Username, p.FollowersCount, p.FollowingCount, p.PostsCount)
	if err != nil {
		if pgerr, ok := errors.AsType[*pgconn.PgError](err); ok &&
			pgerr.Code == "23505" &&
			pgerr.ConstraintName == "users_username_key" {
			return errs.NewError(errs.Conflict, err, domain.UsernameAlreadyUsed)
		}
		return fmt.Errorf("PostgresUserRepository.CreateProfile: %w", err)
	}
	return nil
}

func (pur *PostgresUserRepository) GetAvatarIDByUserID(ctx context.Context, userID int64) (*string, error) {
	query := `SELECT avatar_id FROM profile.users WHERE user_id = $1;`

	var id *string
	if err := pur.conn.QueryRow(ctx, query, userID).Scan(&id); err != nil {
		return nil, fmt.Errorf("PostgresUserRepository.GetAvatarIDByUserID: %w", err)
	}
	return id, nil
}

func (pur *PostgresUserRepository) UpdateProfile(ctx context.Context, userID int64, fieldUpdates map[string]interface{}) error {
	queryBuilder := squirrel.Update("profile.users").
		SetMap(fieldUpdates).
		Where("id = ?", userID).
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

func (pur *PostgresUserRepository) GetSelfProfileByUsername(ctx context.Context, username string) (domain.ProfileReadModel, error) {
	query := `SELECT
					user_id,
					username,
					name,
					avatar_id,
					avatar_thumbnail_id,
					bio,

					followers_count,
					following_count,
					posts_count,
					NULL as relationship

				FROM profile.users
				WHERE username = $1;`

	row, err := pur.conn.Query(ctx, query, username)
	if err != nil {
		return domain.ProfileReadModel{}, fmt.Errorf("PostgresUserRepository.GetSelfProfileByUsername: %w", err)
	}

	profile, err := pgx.CollectOneRow[domain.ProfileReadModel](row, pgx.RowToStructByName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ProfileReadModel{}, errs.NewError(errs.NotFound, err, domain.UserNotFound)
		}
		return domain.ProfileReadModel{}, fmt.Errorf("PostgresUserRepository.GetSelfProfileByUsername: %w", err)
	}

	return profile, nil
}

func (pur *PostgresUserRepository) GetOtherProfileByUsername(ctx context.Context, currentUserID int64, targetUsername string) (domain.ProfileReadModel, error) {
	query := `SELECT
					u.user_id,
					u.username,
					u.name,
					u.avatar_id,
					u.avatar_thumbnail_id,
					u.bio,

					u.followers_count,
					u.following_count,
					u.posts_count,
					COALESCE(
						(SELECT 
							json_build_object('type', 'FOLLOWING', 'follow_id', id) 
							FROM social.follows f 
							WHERE f.follower_id = $1 AND f.following_id = u.user_id
						),
						json_build_object('type', 'NONE', 'follow_id', null)
					) as relationship

				FROM profile.users u
				WHERE username = $2;`

	row, err := pur.conn.Query(ctx, query, currentUserID, targetUsername)
	if err != nil {
		return domain.ProfileReadModel{}, fmt.Errorf("PostgresUserRepository.GetOtherProfileByUsername: %w", err)
	}

	profile, err := pgx.CollectOneRow[domain.ProfileReadModel](row, pgx.RowToStructByName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ProfileReadModel{}, errs.NewError(errs.NotFound, err, domain.UserNotFound)
		}
		return domain.ProfileReadModel{}, fmt.Errorf("PostgresUserRepository.GetOtherProfileByUsername: %w", err)
	}

	return profile, nil
}

func (pur *PostgresUserRepository) UpdateAvatarIDAndAvatarThumbnailID(ctx context.Context, userID int64, avatarID string, avatarThumbnailID string) error {
	query := `UPDATE profile.users SET avatar_id = $1, avatar_thumbnail_id = $2 WHERE user_id = $3;`

	if _, err := pur.conn.Exec(ctx, query, avatarID, avatarThumbnailID, userID); err != nil {
		return fmt.Errorf("PostgresUserRepository.UpdateAvatarID: %w", err)
	}
	return nil
}

func (pur *PostgresUserRepository) FindProfilesByUsername(ctx context.Context, username string) ([]domain.ProfileSummaryReadModel, error) {
	query := `SELECT user_id, username, name, avatar_thumbnail_id FROM profile.users WHERE username LIKE $1 ||'%' LIMIT 10;`

	rows, err := pur.conn.Query(ctx, query, username)
	if err != nil {
		return []domain.ProfileSummaryReadModel{}, fmt.Errorf("PostgresUserRepository.FindProfilesByUsername: %w", err)
	}

	profileSummaries, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.ProfileSummaryReadModel])
	return profileSummaries, nil
}

func (p *PostgresUserRepository) IncPostCount(ctx context.Context, userID int64) error {
	qry := `UPDATE profile.users SET posts_count = posts_count + 1 WHERE user_id = $1;`

	tag, err := p.conn.Exec(ctx, qry, userID)
	if err != nil {
		return fmt.Errorf("PostgresUserRepository.IncPostCount: %w", err)
	}

	if tag.RowsAffected() != 1 {
		return fmt.Errorf("PostgresUserRepository.IncPostCount: %w", err)
	}

	return nil
}

func (p *PostgresUserRepository) DecPostCount(ctx context.Context, userID int64) error {
	qry := `UPDATE profile.users SET posts_count = posts_count - 1 WHERE user_id = $1;`

	tag, err := p.conn.Exec(ctx, qry, userID)
	if err != nil {
		return fmt.Errorf("PostgresUserRepository.DeccPostCount: %w", err)
	}

	if tag.RowsAffected() != 1 {
		return fmt.Errorf("PostgresUserRepository.DecPostCount: %w", err)
	}

	return nil
}

func (p *PostgresUserRepository) IncFollowCount(ctx context.Context, followerID int64, followingID int64) error {
	tx, err := p.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("PostgresUserRepository.IncFollowCount: %w", err)
	}
	defer tx.Rollback(ctx)

	incFollowing := `UPDATE profile.users SET following_count = following_count + 1 WHERE user_id = $1;`
	incFollowers := `UPDATE profile.users SET followers_count = followers_count + 1 WHERE user_id = $1;`

	batch := pgx.Batch{}
	batch.Queue(incFollowing, followerID)
	batch.Queue(incFollowers, followingID)

	if err := tx.SendBatch(ctx, &batch).Close(); err != nil {
		return fmt.Errorf("PostgresUserRepository.IncFollowCount: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("PostgresUserRepository.IncFollowCount: %w", err)
	}

	return nil
}

func (p *PostgresUserRepository) DecFollowCount(ctx context.Context, followerID int64, followingID int64) error {
	tx, err := p.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("PostgresUserRepository.DecFollowCount: %w", err)
	}
	defer tx.Rollback(ctx)

	decFollowing := `UPDATE profile.users SET following_count = following_count - 1 WHERE user_id = $1;`
	decFollowers := `UPDATE profile.users SET followers_count = followers_count - 1 WHERE user_id = $1;`

	batch := pgx.Batch{}
	batch.Queue(decFollowing, followerID)
	batch.Queue(decFollowers, followingID)

	if err := tx.SendBatch(ctx, &batch).Close(); err != nil {
		return fmt.Errorf("PostgresUserRepository.DecFollowCount: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("PostgresUserRepository.DecFollowCount: %w", err)
	}
	
	return nil
}