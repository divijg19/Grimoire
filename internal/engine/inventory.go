package engine

import "strings"

// ================================
// Inventory Helpers (Pure)
// ================================

// GetItemCount returns how many of an item the player has.
func GetItemCount(p *Player, itemID string) int {
	if p.Inventory == nil {
		return 0
	}
	return p.Inventory[NormalizeItemID(itemID)]
}

// AddItem adds qty of an item to the player's inventory.
func AddItem(p *Player, itemID string, qty int) {
	if qty <= 0 {
		return
	}
	itemID = NormalizeItemID(itemID)
	if itemID == "" {
		return
	}
	p.EnsureInventory()
	p.Inventory[itemID] += qty
}

// RemoveItem removes qty of an item from the player's inventory.
// If qty >= current count, the item is removed entirely.
func RemoveItem(p *Player, itemID string, qty int) {
	if qty <= 0 || p.Inventory == nil {
		return
	}
	itemID = NormalizeItemID(itemID)
	if itemID == "" {
		return
	}
	have := p.Inventory[itemID]
	if have <= qty {
		delete(p.Inventory, itemID)
		return
	}
	p.Inventory[itemID] = have - qty
}

// HasItem returns true if the player has at least qty of itemID.
func HasItem(p *Player, itemID string, qty int) bool {
	if qty <= 0 {
		return true
	}
	if p.Inventory == nil {
		return false
	}
	return p.Inventory[NormalizeItemID(itemID)] >= qty
}

// NormalizeItemID converts user/save/catalog IDs to a canonical lower_snake_case key.
func NormalizeItemID(itemID string) string {
	parts := strings.Fields(strings.ToLower(strings.ReplaceAll(itemID, "-", " ")))
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "_")
}

// NormalizeInventory deduplicates inventory keys by canonical ID and removes invalid counts.
func NormalizeInventory(inventory map[string]int) map[string]int {
	if inventory == nil {
		return map[string]int{}
	}

	normalized := make(map[string]int)
	for rawID, qty := range inventory {
		if qty <= 0 {
			continue
		}
		itemID := NormalizeItemID(rawID)
		if itemID == "" {
			continue
		}
		normalized[itemID] += qty
	}

	return normalized
}

// ================================
// Inventory + Events (Optional Helpers)
// ================================

// AddItemWithEvent adds items and emits ItemAdded.
func AddItemWithEvent(p *Player, itemID string, qty int) Events {
	if qty <= 0 {
		return nil
	}
	AddItem(p, itemID, qty)
	return Events{
		ItemAdded{
			ItemID: itemID,
			Count:  qty,
		},
	}
}

// RemoveItemWithEvent removes items and emits ItemRemoved.
func RemoveItemWithEvent(p *Player, itemID string, qty int) Events {
	if qty <= 0 {
		return nil
	}
	RemoveItem(p, itemID, qty)
	return Events{
		ItemRemoved{
			ItemID: itemID,
			Count:  qty,
		},
	}
}
