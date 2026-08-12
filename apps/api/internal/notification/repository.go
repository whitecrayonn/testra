package notification

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/testra/testra/apps/api/internal/shared/db"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
)

type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type SQLRepository struct {
	db DBTX
}

func NewSQLRepository(sqlDB *sql.DB) *SQLRepository {
	return &SQLRepository{db: db.Wrap(sqlDB)}
}

func (r *SQLRepository) CreateNotification(ctx context.Context, n *Notification) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO notifications (id, organization_id, user_id, type, title, body, link, read, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		n.ID, n.OrganizationID, n.UserID, n.Type, n.Title, n.Body, n.Link, n.Read, n.CreatedAt, n.UpdatedAt,
	)
	return err
}

func (r *SQLRepository) GetNotification(ctx context.Context, id, userID uuid.UUID) (*Notification, error) {
	var n Notification
	err := r.db.QueryRowContext(ctx,
		`SELECT id, organization_id, user_id, type, title, body, link, read, created_at, updated_at
		 FROM notifications
		 WHERE id = $1 AND user_id = $2`,
		id, userID,
	).Scan(&n.ID, &n.OrganizationID, &n.UserID, &n.Type, &n.Title, &n.Body, &n.Link, &n.Read, &n.CreatedAt, &n.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sharederrors.ErrNotFound
	}
	return &n, err
}

func (r *SQLRepository) ListNotifications(ctx context.Context, userID uuid.UUID, read *bool, cursor string, limit int) ([]Notification, string, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var createdAtCondition string
	var args []interface{}
	args = append(args, userID)

	if cursor != "" {
		t, err := decodeTimeCursor(cursor)
		if err == nil {
			createdAtCondition = "AND created_at < $2"
			args = append(args, t)
		}
	}

	readCondition := ""
	if read != nil {
		paramIdx := len(args) + 1
		readCondition = fmt.Sprintf("AND read = $%d", paramIdx)
		args = append(args, *read)
	}

	query := fmt.Sprintf(
		`SELECT id, organization_id, user_id, type, title, body, link, read, created_at, updated_at
		 FROM notifications
		 WHERE user_id = $1 %s %s
		 ORDER BY created_at DESC, id DESC
		 LIMIT %d`,
		createdAtCondition, readCondition, limit+1,
	)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID, &n.OrganizationID, &n.UserID, &n.Type, &n.Title, &n.Body, &n.Link, &n.Read, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, "", err
		}
		notifications = append(notifications, n)
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	nextCursor := ""
	if len(notifications) > limit {
		last := notifications[limit-1]
		nextCursor = encodeTimeCursor(last.CreatedAt)
		notifications = notifications[:limit]
	}

	return notifications, nextCursor, nil
}

func (r *SQLRepository) CountUnread(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND read = false`,
		userID,
	).Scan(&count)
	return count, err
}

func (r *SQLRepository) MarkRead(ctx context.Context, id, userID uuid.UUID, read bool) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE notifications SET read = $1, updated_at = NOW() WHERE id = $2 AND user_id = $3`,
		read, id, userID,
	)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sharederrors.ErrNotFound
	}
	return nil
}

func (r *SQLRepository) DeleteNotification(ctx context.Context, id, userID uuid.UUID) error {
	res, err := r.db.ExecContext(ctx,
		`DELETE FROM notifications WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sharederrors.ErrNotFound
	}
	return nil
}

func (r *SQLRepository) GetPreferences(ctx context.Context, orgID, userID uuid.UUID) (*NotificationPreferences, error) {
	var p NotificationPreferences
	err := r.db.QueryRowContext(ctx,
		`SELECT id, organization_id, user_id, in_app_enabled, email_enabled, slack_enabled, teams_enabled, webhook_enabled, created_at, updated_at
		 FROM notification_preferences
		 WHERE organization_id = $1 AND user_id = $2`,
		orgID, userID,
	).Scan(&p.ID, &p.OrganizationID, &p.UserID, &p.InAppEnabled, &p.EmailEnabled, &p.SlackEnabled, &p.TeamsEnabled, &p.WebhookEnabled, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		// Return default preferences
		return &NotificationPreferences{
			OrganizationID: orgID,
			UserID:         userID,
			InAppEnabled:   true,
			EmailEnabled:   false,
			SlackEnabled:   false,
			TeamsEnabled:   false,
			WebhookEnabled: false,
		}, nil
	}
	return &p, err
}

func (r *SQLRepository) UpsertPreferences(ctx context.Context, p *NotificationPreferences) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO notification_preferences (id, organization_id, user_id, in_app_enabled, email_enabled, slack_enabled, teams_enabled, webhook_enabled, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 ON CONFLICT (organization_id, user_id)
		 DO UPDATE SET in_app_enabled = EXCLUDED.in_app_enabled, email_enabled = EXCLUDED.email_enabled,
		               slack_enabled = EXCLUDED.slack_enabled, teams_enabled = EXCLUDED.teams_enabled,
		               webhook_enabled = EXCLUDED.webhook_enabled, updated_at = EXCLUDED.updated_at`,
		p.ID, p.OrganizationID, p.UserID, p.InAppEnabled, p.EmailEnabled, p.SlackEnabled, p.TeamsEnabled, p.WebhookEnabled, p.CreatedAt, p.UpdatedAt,
	)
	return err
}

func (r *SQLRepository) CreateChannel(ctx context.Context, ch *NotificationChannel) error {
	config, err := json.Marshal(ch.Config)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx,
		`INSERT INTO notification_channels (id, organization_id, workspace_id, type, name, config, created_by, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		ch.ID, ch.OrganizationID, ch.WorkspaceID, ch.Type, ch.Name, config, ch.CreatedBy, ch.CreatedAt, ch.UpdatedAt,
	)
	return err
}

func (r *SQLRepository) GetChannel(ctx context.Context, id uuid.UUID) (*NotificationChannel, error) {
	var ch NotificationChannel
	var config []byte
	err := r.db.QueryRowContext(ctx,
		`SELECT id, organization_id, workspace_id, type, name, config, created_by, created_at, updated_at
		 FROM notification_channels
		 WHERE id = $1`,
		id,
	).Scan(&ch.ID, &ch.OrganizationID, &ch.WorkspaceID, &ch.Type, &ch.Name, &config, &ch.CreatedBy, &ch.CreatedAt, &ch.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sharederrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(config, &ch.Config); err != nil {
		return nil, err
	}
	return &ch, nil
}

func (r *SQLRepository) ListChannels(ctx context.Context, workspaceID uuid.UUID, cursor string, limit int) ([]NotificationChannel, error) {
	var rows *sql.Rows
	var err error

	if cursor != "" {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, organization_id, workspace_id, type, name, config, created_by, created_at, updated_at
			 FROM notification_channels
			 WHERE workspace_id = $1 AND id < $2
			 ORDER BY id DESC LIMIT $3`,
			workspaceID, cursor, limit,
		)
	} else {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, organization_id, workspace_id, type, name, config, created_by, created_at, updated_at
			 FROM notification_channels
			 WHERE workspace_id = $1
			 ORDER BY id DESC LIMIT $2`,
			workspaceID, limit,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []NotificationChannel
	for rows.Next() {
		var ch NotificationChannel
		var config []byte
		if err := rows.Scan(&ch.ID, &ch.OrganizationID, &ch.WorkspaceID, &ch.Type, &ch.Name, &config, &ch.CreatedBy, &ch.CreatedAt, &ch.UpdatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(config, &ch.Config); err != nil {
			return nil, err
		}
		channels = append(channels, ch)
	}
	return channels, rows.Err()
}

func (r *SQLRepository) UpdateChannel(ctx context.Context, ch *NotificationChannel) error {
	config, err := json.Marshal(ch.Config)
	if err != nil {
		return err
	}
	res, err := r.db.ExecContext(ctx,
		`UPDATE notification_channels
		 SET type = $1, name = $2, config = $3, updated_at = $4
		 WHERE id = $5`,
		ch.Type, ch.Name, config, ch.UpdatedAt, ch.ID,
	)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sharederrors.ErrNotFound
	}
	return nil
}

func (r *SQLRepository) DeleteChannel(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.ExecContext(ctx,
		`DELETE FROM notification_channels WHERE id = $1`,
		id,
	)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sharederrors.ErrNotFound
	}
	return nil
}

const cursorTimeFormat = time.RFC3339Nano

func encodeTimeCursor(t time.Time) string {
	return base64.URLEncoding.EncodeToString([]byte(t.UTC().Format(cursorTimeFormat)))
}

func decodeTimeCursor(cursor string) (time.Time, error) {
	b, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, err
	}
	return time.Parse(cursorTimeFormat, string(b))
}

func (r *SQLRepository) CreateTemplate(ctx context.Context, t *NotificationTemplate) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO notification_templates (id, organization_id, name, event_type, channel_type, subject, body, created_by, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		t.ID, t.OrganizationID, t.Name, t.EventType, t.ChannelType, t.Subject, t.Body, t.CreatedBy, t.CreatedAt, t.UpdatedAt)
	return err
}

func (r *SQLRepository) GetTemplate(ctx context.Context, id uuid.UUID) (*NotificationTemplate, error) {
	var t NotificationTemplate
	err := r.db.QueryRowContext(ctx,
		`SELECT id, organization_id, name, event_type, channel_type, subject, body, created_by, created_at, updated_at
		 FROM notification_templates WHERE id = $1`, id,
	).Scan(&t.ID, &t.OrganizationID, &t.Name, &t.EventType, &t.ChannelType, &t.Subject, &t.Body, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sharederrors.ErrNotFound
	}
	return &t, err
}

func (r *SQLRepository) ListTemplates(ctx context.Context, orgID uuid.UUID, eventType, channelType string, limit int) ([]NotificationTemplate, error) {
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	query := `SELECT id, organization_id, name, event_type, channel_type, subject, body, created_by, created_at, updated_at
		 FROM notification_templates WHERE organization_id = $1 `
	args := []interface{}{orgID}
	if eventType != "" {
		query += ` AND event_type = $` + fmt.Sprintf("%d ", len(args)+1)
		args = append(args, eventType)
	}
	if channelType != "" {
		query += ` AND channel_type = $` + fmt.Sprintf("%d ", len(args)+1)
		args = append(args, channelType)
	}
	query += ` ORDER BY updated_at DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1)
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []NotificationTemplate
	for rows.Next() {
		var t NotificationTemplate
		if err := rows.Scan(&t.ID, &t.OrganizationID, &t.Name, &t.EventType, &t.ChannelType, &t.Subject, &t.Body, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, rows.Err()
}

func (r *SQLRepository) UpdateTemplate(ctx context.Context, t *NotificationTemplate) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE notification_templates SET name = $2, event_type = $3, channel_type = $4, subject = $5, body = $6, updated_at = $7 WHERE id = $1`,
		t.ID, t.Name, t.EventType, t.ChannelType, t.Subject, t.Body, t.UpdatedAt)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sharederrors.ErrNotFound
	}
	return nil
}

func (r *SQLRepository) DeleteTemplate(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM notification_templates WHERE id = $1`, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sharederrors.ErrNotFound
	}
	return nil
}

func (r *SQLRepository) CreateHistory(ctx context.Context, h *NotificationHistory) error {
	chID := uuid.NullUUID{}
	if h.ChannelID != nil {
		chID = uuid.NullUUID{UUID: *h.ChannelID, Valid: true}
	}
	notifID := uuid.NullUUID{UUID: h.NotificationID, Valid: h.NotificationID != uuid.Nil}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO notification_history (id, organization_id, notification_id, channel_id, channel_type, status, error_message, retry_count, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		h.ID, h.OrganizationID, notifID, chID, h.ChannelType, h.Status, h.ErrorMessage, h.RetryCount, h.CreatedAt, h.UpdatedAt)
	return err
}

func (r *SQLRepository) UpdateHistory(ctx context.Context, h *NotificationHistory) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE notification_history SET status = $2, error_message = $3, retry_count = $4, updated_at = $5 WHERE id = $1`,
		h.ID, h.Status, h.ErrorMessage, h.RetryCount, h.UpdatedAt)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sharederrors.ErrNotFound
	}
	return nil
}

func (r *SQLRepository) ListHistory(ctx context.Context, notificationID uuid.UUID, limit int) ([]NotificationHistory, error) {
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, organization_id, notification_id, channel_id, channel_type, status, error_message, retry_count, created_at, updated_at
		 FROM notification_history WHERE notification_id = $1 ORDER BY created_at DESC LIMIT $2`,
		notificationID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []NotificationHistory
	for rows.Next() {
		var h NotificationHistory
		var chID sql.NullString
		if err := rows.Scan(&h.ID, &h.OrganizationID, &h.NotificationID, &chID, &h.ChannelType, &h.Status, &h.ErrorMessage, &h.RetryCount, &h.CreatedAt, &h.UpdatedAt); err != nil {
			return nil, err
		}
		if chID.Valid {
			id, _ := uuid.Parse(chID.String)
			h.ChannelID = &id
		}
		list = append(list, h)
	}
	return list, rows.Err()
}
