package repo

import (
	"github.com/0xfig521/tide/internal/db"
	"github.com/0xfig521/tide/internal/models"
)

// CategoryRepo handles category data access.
type CategoryRepo struct {
	db *db.DB
}

func NewCategoryRepo(db *db.DB) *CategoryRepo {
	return &CategoryRepo{db: db}
}

// Create inserts a new category.
func (r *CategoryRepo) Create(name, description string) (*models.Category, error) {
	result, err := r.db.Conn.Exec(`
		INSERT INTO categories (name, description) VALUES (?, ?)
	`, name, description)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return r.GetByID(id)
}

// GetByID retrieves a category by ID.
func (r *CategoryRepo) GetByID(id int64) (*models.Category, error) {
	c := &models.Category{}
	var createdAt, updatedAt string
	err := r.db.Conn.QueryRow(`
		SELECT id, name, description, created_at, updated_at
		FROM categories WHERE id = ?
	`, id).Scan(&c.ID, &c.Name, &c.Description, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	c.CreatedAt = mustParseTime(createdAt)
	c.UpdatedAt = mustParseTime(updatedAt)
	return c, nil
}

// GetByName retrieves a category by name.
func (r *CategoryRepo) GetByName(name string) (*models.Category, error) {
	c := &models.Category{}
	var createdAt, updatedAt string
	err := r.db.Conn.QueryRow(`
		SELECT id, name, description, created_at, updated_at
		FROM categories WHERE name = ?
	`, name).Scan(&c.ID, &c.Name, &c.Description, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	c.CreatedAt = mustParseTime(createdAt)
	c.UpdatedAt = mustParseTime(updatedAt)
	return c, nil
}

// List retrieves all categories with feed counts.
func (r *CategoryRepo) List() ([]*models.Category, error) {
	rows, err := r.db.Conn.Query(`
		SELECT c.id, c.name, c.description, c.created_at, c.updated_at,
			COUNT(fc.feed_id) as feed_count
		FROM categories c
		LEFT JOIN feed_categories fc ON fc.category_id = c.id
		GROUP BY c.id
		ORDER BY c.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*models.Category
	for rows.Next() {
		c := &models.Category{}
		var createdAt, updatedAt string
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &createdAt, &updatedAt, &c.FeedCount); err != nil {
			return nil, err
		}
		c.CreatedAt = mustParseTime(createdAt)
		c.UpdatedAt = mustParseTime(updatedAt)
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

// Delete removes a category.
func (r *CategoryRepo) Delete(id int64) error {
	_, err := r.db.Conn.Exec(`DELETE FROM categories WHERE id = ?`, id)
	return err
}

// DeleteByName removes a category by name.
func (r *CategoryRepo) DeleteByName(name string) error {
	_, err := r.db.Conn.Exec(`DELETE FROM categories WHERE name = ?`, name)
	return err
}

// Update modifies a category.
func (r *CategoryRepo) Update(id int64, name, description string) error {
	_, err := r.db.Conn.Exec(`
		UPDATE categories SET name = ?, description = ?, updated_at = datetime('now')
		WHERE id = ?
	`, name, description, id)
	return err
}
