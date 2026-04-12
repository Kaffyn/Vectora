extends Node
## Quick access hotbar for equipped items
## Manages 5 slots for quick equip/use during gameplay

class_name Hotbar

const HOTBAR_SIZE: int = 5

var slots: Array[Item] = []
var selected_slot: int = 0
var inventory: Inventory = null

signal item_equipped(item: Item, slot: int)
signal item_unequipped(slot: int)
signal selection_changed(slot: int)

func _ready() -> void:
	# Initialize empty hotbar slots
	slots.resize(HOTBAR_SIZE)

func set_inventory(inv: Inventory) -> void:
	"""Link this hotbar to an inventory"""
	inventory = inv
	if inventory:
		inventory.item_added.connect(_on_inventory_item_added)
		inventory.item_removed.connect(_on_inventory_item_removed)

func equip_item(inventory_slot: int) -> bool:
	"""Equip item from inventory to hotbar"""
	if !inventory or inventory_slot < 0 or inventory_slot >= inventory.items.size():
		return false

	var item = inventory.get_item(inventory_slot)
	if !item:
		return false

	# Find first empty hotbar slot
	for i in range(slots.size()):
		if !slots[i]:
			slots[i] = item
			item_equipped.emit(item, i)
			return true

	return false

func equip_item_to_slot(inventory_slot: int, hotbar_slot: int) -> bool:
	"""Equip item from inventory to specific hotbar slot"""
	if !inventory or inventory_slot < 0 or inventory_slot >= inventory.items.size():
		return false

	if hotbar_slot < 0 or hotbar_slot >= slots.size():
		return false

	var item = inventory.get_item(inventory_slot)
	if !item:
		return false

	slots[hotbar_slot] = item
	item_equipped.emit(item, hotbar_slot)
	return true

func unequip_item(hotbar_slot: int) -> bool:
	"""Remove item from hotbar slot"""
	if hotbar_slot < 0 or hotbar_slot >= slots.size():
		return false

	if slots[hotbar_slot]:
		slots[hotbar_slot] = null
		item_unequipped.emit(hotbar_slot)
		return true

	return false

func select_slot(slot: int) -> bool:
	"""Select hotbar slot for use"""
	if slot < 0 or slot >= slots.size():
		return false

	selected_slot = slot
	selection_changed.emit(slot)
	return true

func get_selected_item() -> Item:
	"""Get currently selected hotbar item"""
	if selected_slot >= 0 and selected_slot < slots.size():
		return slots[selected_slot]
	return null

func get_item(slot: int) -> Item:
	"""Get item at hotbar slot"""
	if slot >= 0 and slot < slots.size():
		return slots[slot]
	return null

func is_slot_empty(slot: int) -> bool:
	"""Check if hotbar slot is empty"""
	if slot >= 0 and slot < slots.size():
		return slots[slot] == null
	return true

func get_free_slots() -> int:
	"""Count empty hotbar slots"""
	var count = 0
	for item in slots:
		if !item:
			count += 1
	return count

func _on_inventory_item_added(_item: Item, _slot: int) -> void:
	"""Called when item added to inventory"""
	# Optional: Auto-equip first item
	pass

func _on_inventory_item_removed(item: Item, _slot: int) -> void:
	"""Called when item removed from inventory"""
	# Remove from hotbar if equipped there
	for i in range(slots.size()):
		if slots[i] == item:
			slots[i] = null
			item_unequipped.emit(i)
