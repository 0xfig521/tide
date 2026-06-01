package repo

import (
	"regexp"

	"github.com/0xfig521/tide/internal/db"
	"github.com/0xfig521/tide/internal/models"
)

type Rule struct {
	ID          int64  `json:"id"`
	Priority    int    `json:"priority"`
	IsActive    bool   `json:"is_active"`
	MatchField  string `json:"match_field"`
	MatchRegex  string `json:"match_regex"`
	Action      string `json:"action"`
	ActionValue string `json:"action_value"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type RuleRepo struct {
	db *db.DB
}

func NewRuleRepo(db *db.DB) *RuleRepo {
	return &RuleRepo{db: db}
}

func (r *RuleRepo) Create(rule *Rule) (int64, error) {
	result, err := r.db.Conn.Exec(`
		INSERT INTO rules (priority, is_active, match_field, match_regex, action, action_value)
		VALUES (?, ?, ?, ?, ?, ?)
	`, rule.Priority, boolToInt(rule.IsActive), rule.MatchField, rule.MatchRegex, rule.Action, rule.ActionValue)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *RuleRepo) List() ([]*Rule, error) {
	rows, err := r.db.Conn.Query(`
		SELECT id, priority, is_active, match_field, match_regex, action, action_value,
		       COALESCE(created_at,''), COALESCE(updated_at,'')
		FROM rules
		ORDER BY priority DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []*Rule
	for rows.Next() {
		rule := &Rule{}
		var isActive int
		if err := rows.Scan(
			&rule.ID, &rule.Priority, &isActive,
			&rule.MatchField, &rule.MatchRegex,
			&rule.Action, &rule.ActionValue,
			&rule.CreatedAt, &rule.UpdatedAt,
		); err != nil {
			return nil, err
		}
		rule.IsActive = isActive != 0
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}

func (r *RuleRepo) Delete(id int64) error {
	_, err := r.db.Conn.Exec(`DELETE FROM rules WHERE id = ?`, id)
	return err
}

func (r *RuleRepo) Apply(entry *models.Entry) (map[string]string, error) {
	rules, err := r.List()
	if err != nil {
		return nil, err
	}

	applied := make(map[string]string)
	for _, rule := range rules {
		if !rule.IsActive {
			continue
		}

		re, err := regexp.Compile(rule.MatchRegex)
		if err != nil {
			continue
		}

		fieldValue := getEntryField(entry, rule.MatchField)
		if !re.MatchString(fieldValue) {
			continue
		}

		switch rule.Action {
		case "tag":
			applied["tag"] = rule.ActionValue
		case "state":
			applied["state"] = rule.ActionValue
		case "ignore":
			applied["state"] = "ignored"
			return applied, nil
		case "priority":
			applied["priority"] = rule.ActionValue
		case "category":
			applied["category"] = rule.ActionValue
		}
	}
	return applied, nil
}

func getEntryField(entry *models.Entry, field string) string {
	switch field {
	case "title":
		return entry.Title
	case "description":
		return entry.Description
	case "content":
		return entry.Content
	case "author":
		return entry.AuthorName
	case "category":
		return entry.Categories
	default:
		return entry.Title
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
