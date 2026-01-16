package store

import "fmt"

func (v *VaultStore) CreateRelationship(rel *Relationship) error {
	_, err := v.db.Exec(`
		INSERT OR REPLACE INTO relationships (source_id, target_id, relation_type, strength)
		VALUES (?, ?, ?, ?)`,
		rel.SourceID, rel.TargetID, rel.RelationType, rel.Strength)
	return err
}

func (v *VaultStore) GetRelationships(itemID string) ([]Relationship, error) {
	rows, err := v.db.Query(`
		SELECT id, source_id, target_id, relation_type, strength
		FROM relationships
		WHERE source_id = ? OR target_id = ?`, itemID, itemID)
	if err != nil {
		return nil, fmt.Errorf("query relationships: %w", err)
	}
	defer rows.Close()

	var rels []Relationship
	for rows.Next() {
		var r Relationship
		if err := rows.Scan(&r.ID, &r.SourceID, &r.TargetID, &r.RelationType, &r.Strength); err != nil {
			return nil, err
		}
		rels = append(rels, r)
	}
	return rels, nil
}

// GetGraph returns all items and relationships for graph visualization
func (v *VaultStore) GetGraph() ([]Item, []Relationship, error) {
	items, err := v.ListItems(1000, 0)
	if err != nil {
		return nil, nil, err
	}

	rows, err := v.db.Query(`SELECT id, source_id, target_id, relation_type, strength FROM relationships`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var rels []Relationship
	for rows.Next() {
		var r Relationship
		rows.Scan(&r.ID, &r.SourceID, &r.TargetID, &r.RelationType, &r.Strength)
		rels = append(rels, r)
	}
	return items, rels, nil
}

func (v *VaultStore) DeleteRelationship(sourceID, targetID string) error {
	_, err := v.db.Exec(`DELETE FROM relationships WHERE source_id = ? AND target_id = ?`, sourceID, targetID)
	return err
}
